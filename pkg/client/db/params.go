package db

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	"entgo.io/ent/dialect"
	"github.com/eroshiva/cloudtalk/internal/ent"
	_ "github.com/lib/pq" // SQL driver, necessary for DB interaction
)

const (
	defaultHost     = "localhost"
	defaultPort     = 5432
	defaultUser     = "admin"
	defaultPassword = "pass"
	defaultDatabase = "postgres"
	defaultSSLMode  = "disable"

	envUser     = "PGUSER"
	envPass     = "PGPASSWORD"
	envHost     = "PGHOST"
	envPort     = "PGPORT"
	envDatabase = "PGDATABASE"
	envSSLMode  = "PGSSLMODE"
)

var (
	host     = ""
	port     = -1
	user     = ""
	dbName   = ""
	password = ""
	sslmode  = ""
)

func init() {
	// parsing env variables needed for the DB connectivity.
	host = os.Getenv(envHost)
	if host == "" {
		host = defaultHost
	}
	portVar := os.Getenv(envPort)
	if portVar == "" {
		port = defaultPort
	} else {
		convPort, err := strconv.Atoi(portVar)
		if err != nil {
			zlog.Fatal().Err(err).Msg("invalid port value")
		}
		port = convPort
	}
	user = os.Getenv(envUser)
	if user == "" {
		user = defaultUser
	}
	password = os.Getenv(envPass)
	if password == "" {
		password = defaultPassword
	}
	dbName = os.Getenv(envDatabase)
	if dbName == "" {
		dbName = defaultDatabase
	}
	sslmode = os.Getenv(envSSLMode)
	if sslmode == "" {
		sslmode = defaultSSLMode
	}
}

// getDataSourceName returns data source name variable based on set environment parameters.
func getDataSourceName() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbName, sslmode)
}

// RunSchemaMigration instantiates connection to the DB and performs automatic migration.
func RunSchemaMigration() (*ent.Client, error) {
	zlog.Info().Msgf("Opening connection to PostreSQL...")
	client, err := ent.Open(dialect.Postgres, getDataSourceName())
	if err != nil {
		zlog.Error().Err(err).Msg("failed opening connection to postgres")
		return nil, err
	}

	zlog.Info().Msgf("Migrating database schema...")
	// Run the auto migration tool.
	if err = client.Schema.Create(context.Background()); err != nil {
		zlog.Error().Err(err).Msg("failed creating schema resources")
		// gracefully closing client
		newErr := GracefullyCloseDBClient(client)
		if newErr != nil {
			return nil, errors.Join(err, newErr)
		}
		return nil, err
	}

	return client, nil
}

// GracefullyCloseDBClient gracefully closes connection with the DB.
func GracefullyCloseDBClient(client *ent.Client) error {
	zlog.Info().Msg("Gracefully closing connection to the DB")
	err := client.Close()
	if err != nil {
		zlog.Error().Err(err).Msgf("failed closing connection to postgres")
		return err
	}
	return nil
}
