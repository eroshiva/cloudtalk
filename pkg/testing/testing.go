// Package testing contains helper functions to run unit tests within this repository.
package testing

import (
	"sync"
	"time"

	"github.com/eroshiva/cloudtalk/internal/ent"
	"github.com/eroshiva/cloudtalk/pkg/client/db"
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
