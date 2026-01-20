// Package testing contains helper functions to run unit tests within this repository.
package testing

import (
	"sync"
	"time"

	apiv1 "github.com/eroshiva/cloudtalk/api/v1"
	"github.com/eroshiva/cloudtalk/internal/ent"
	"github.com/eroshiva/cloudtalk/internal/server"
	"github.com/eroshiva/cloudtalk/pkg/client/db"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// DefaultTestTimeout defines default test timeout for a testing package
const (
	DefaultTestTimeout           = time.Second * 1
	defaultGRPCTestServerAddress = "localhost:50051"
	defaultHTTPTestServerAddress = "localhost:50052"
)

// Setup function sets up testing environment. Currently, only uploading schema to the DB.
func Setup() (*ent.Client, error) {
	return db.RunSchemaMigration()
}

// SetupFull function sets up testing environment.It uploads schema to the DB and starts gRPC and HTTP reverse proxy servers.
func SetupFull(grpcServerAddress, httpServerAddress string) (*ent.Client, apiv1.ProductReviewsServiceClient, *sync.WaitGroup, chan bool, chan bool, error) {
	if grpcServerAddress == "" {
		grpcServerAddress = defaultGRPCTestServerAddress
	}

	if httpServerAddress == "" {
		httpServerAddress = defaultHTTPTestServerAddress
	}

	client, err := db.RunSchemaMigration()
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	wg := &sync.WaitGroup{}
	termChan := make(chan bool, 1)
	readyChan := make(chan bool, 1)
	reverseProxyReadyChan := make(chan bool, 1)
	reverseProxyTermChan := make(chan bool, 1)

	wg.Go(func() {
		server.StartServer(grpcServerAddress, httpServerAddress, client, wg, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan)
	})
	// Waiting until both servers are up and running
	<-readyChan
	<-reverseProxyReadyChan

	// creating gRPC testing client
	conn, err := grpc.NewClient(grpcServerAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, nil, nil, nil, nil, err
	}

	grpcClient := apiv1.NewProductReviewsServiceClient(conn)

	return client, grpcClient, wg, termChan, reverseProxyTermChan, nil
}

// TeardownFull function tears down testing suite including DB connection, gRPC and HTTP reverse proxy servers.
func TeardownFull(client *ent.Client, wg *sync.WaitGroup, termChan, reverseProxyTermChan chan bool) {
	close(termChan)
	close(reverseProxyTermChan)
	err := db.GracefullyCloseDBClient(client)
	if err != nil {
		panic(err)
	}
	wg.Wait()
}
