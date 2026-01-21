// Package cmd is a main entry point to our service.
package main

import (
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/eroshiva/cloudtalk/internal/server"
	"github.com/eroshiva/cloudtalk/pkg/client/db"
	"github.com/eroshiva/cloudtalk/pkg/logger"
	"github.com/eroshiva/cloudtalk/pkg/rabbitmq"
)

var zlog = logger.NewLogger("main")

func main() { //nolint:unused // this is a main entry point to our service.
	zlog.Info().Msgf("Starting product review service")

	// channels to handle termination and capture signals
	termChan := make(chan bool)
	reverseProxyTermChan := make(chan bool)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	readyChan := make(chan bool, 1)
	reverseProxyReadyChan := make(chan bool, 1)

	// connecting to DB
	dbClient, err := db.RunSchemaMigration()
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to instantiate connection with PostgreSQL DB")
	}

	// connecting to RabbitMQ
	rabbitMQConn, rabbitMQCh, err := rabbitmq.Connect()
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}

	// creating waitgroup so main will wait for servers to exit cleanly
	wg := &sync.WaitGroup{}
	wg.Go(func() {
		<-sigChan
		zlog.Info().Msg("Shutdown signal received. Closing channels...")
		close(termChan)
		close(reverseProxyTermChan)
	})

	// starting NB API server.
	server.StartServer(server.GetGRPCServerAddress(), server.GetHTTPServerAddress(), dbClient, rabbitMQCh, wg, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan)

	wg.Wait()
	zlog.Info().Msg("Shutting down product reviews service")

	// gracefully closing clients
	err = db.GracefullyCloseDBClient(dbClient)
	if err != nil {
		zlog.Error().Err(err).Msg("Failed to gracefully close DB connection")
	}
	rabbitmq.CloseConnection(rabbitMQConn)
	rabbitmq.CloseChannel(rabbitMQCh)

	zlog.Info().Msgf("Shutdown is complete. Goodbye!")
}
