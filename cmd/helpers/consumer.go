package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/eroshiva/cloudtalk/pkg/logger"
	"github.com/eroshiva/cloudtalk/pkg/rabbitmq"
)

var zlog = logger.NewLogger("cloudtalk-consumer")

func main() {
	msgs, conn, ch, err := rabbitmq.Connect()
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to connect to RabbitMQ")
	}
	defer rabbitmq.CloseConnection(conn)
	defer rabbitmq.CloseChannel(ch)

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	go func() {
		for d := range msgs {
			zlog.Info().Msgf("Received a message: %s", d.Body)
		}
	}()

	zlog.Info().Msgf("Listening for messages")
	<-sigChan
	zlog.Info().Msg("Shutdown signal is received. Wrapping up...")
}
