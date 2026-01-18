// Package server contains main server code for gRPC API and HTTP reverse proxy.
package server

import (
	"context"
	"net"
	"net/http"
	"os"
	"sync"

	apiv1 "github.com/eroshiva/cloudtalk/api/v1"
	"github.com/eroshiva/cloudtalk/internal/ent"
	"github.com/eroshiva/cloudtalk/pkg/logger"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	// server configuration-related constants
	tcpNetwork               = "tcp"
	defaultServerAddress     = "localhost:50051"
	envServerAddress         = "GRPC_SERVER_ADDRESS" // must be in form address:port, e.g., localhost:50051.
	envHTTPServerAddress     = "HTTP_SERVER_ADDRESS" // must be in form address:port, e.g., localhost:80.
	defaultHTTPServerAddress = "localhost:50052"
)

var zlog = logger.NewLogger("server")

// server implements the DeviceMonitoringServiceServer interface.
type server struct {
	apiv1.ProductReviewsServiceServer

	// client for interactions with DB
	dbClient *ent.Client
}

// Options structure defines server's features enablement.
type Options struct {
	EnableInterceptor bool
}

func getServerOptions(_ *Options) ([]grpc.ServerOption, error) {
	// parse server options from configuration
	optionsList := make([]grpc.ServerOption, 0)
	return optionsList, nil
}

func serve(grpcAddress, httpAddress string, dbClient *ent.Client, wg *sync.WaitGroup, serverOptions []grpc.ServerOption,
	termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan chan bool,
) {
	grpcReadyChan := make(chan bool, 1)
	lis, err := net.Listen(tcpNetwork, grpcAddress)
	if err != nil {
		zlog.Fatal().Err(err).Msgf("Failed to listen on %s", grpcAddress)
	}

	// Create a new gRPC server instance.
	s := grpc.NewServer(serverOptions...)

	gRPCServer := &server{
		dbClient: dbClient,
	}

	// Register our server implementation with the gRPC server.
	apiv1.RegisterProductReviewsServiceServer(s, gRPCServer)

	// Start the server.
	zlog.Info().Msgf("gRPC server listening at %v", lis.Addr())

	go func() {
		// On testing will be nil
		if readyChan != nil {
			readyChan <- true
			grpcReadyChan <- true
		}
		if err := s.Serve(lis); err != nil {
			zlog.Fatal().Err(err).Msgf("Failed to serve")
		}
	}()

	// starting reverse proxy
	wg.Add(1)
	go func() {
		wg.Add(1) //nolint:staticcheck
		startReverseProxy(grpcAddress, httpAddress, wg, grpcReadyChan, reverseProxyReadyChan, reverseProxyTermChan)
		wg.Done()
	}()

	// handle termination signals
	termSig := <-termChan
	if termSig {
		zlog.Info().Msg("Gracefully stopping gRPC server")
		s.Stop()
	}
	// report to waitgroup that process is finished
	wg.Done()
}

// startReverseProxy starts the gRPC reverse proxy server which is connected to the HTTP handler.
func startReverseProxy(grpcServerAddress, httpServerAddress string, wg *sync.WaitGroup, grocReadyChan, reverseProxyReadyChan, reverseProxyTermChan chan bool) {
	// waiting for the gRPC server to start first
	<-grocReadyChan
	zlog.Info().Msg("Starting reverse HTTP proxy")

	// creating the gRPC-Gateway reverse proxy.
	conn, err := grpc.NewClient(
		grpcServerAddress, // The address of the gRPC server
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to dial to gRPC server")
	}

	mux := runtime.NewServeMux()

	// Registering HTTP handler for our service and connecting the gateway to our gRPC server.
	if err = apiv1.RegisterProductReviewsServiceHandler(context.Background(), mux, conn); err != nil {
		zlog.Fatal().Err(err).Msg("Failed to register HTTP gateway")
	}

	// now, create and start the HTTP server (i.e., our gateway).
	gwServer := &http.Server{
		Addr:    httpServerAddress,
		Handler: mux,
	}

	go func() {
		// On testing will be nil
		if reverseProxyReadyChan != nil {
			reverseProxyReadyChan <- true
		}
		if err = gwServer.ListenAndServe(); err != nil {
			zlog.Fatal().Err(err).Msg("Failed to serve HTTP gateway")
		}
	}()

	// handle termination signals
	termSig := <-reverseProxyTermChan
	if termSig {
		zlog.Info().Msg("Gracefully stopping HTTP server")
		err = gwServer.Shutdown(context.Background())
		if err != nil {
			zlog.Fatal().Err(err).Msg("Failed to gracefully shutdown HTTP gateway")
		}
	}
	// report to waitgroup that process is finished
	wg.Done()
}

// GetGRPCServerAddress function reads environmental variable and returns a gRPC server address.
func GetGRPCServerAddress() string {
	// read env variable, where gRPC server is running
	serverAddress := os.Getenv(envServerAddress)
	if serverAddress == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, using default gRPC server address: %s",
			envServerAddress, defaultServerAddress)
		serverAddress = defaultServerAddress
	}
	return serverAddress
}

// GetHTTPServerAddress function reads environmental variable and returns an HTTP server address.
func GetHTTPServerAddress() string {
	// read env variable, where HTTP server is running
	httpServerAddress := os.Getenv(envHTTPServerAddress)
	if httpServerAddress == "" {
		zlog.Warn().Msgf("Environment variable \"%s\" is not set, using default address: %s",
			envHTTPServerAddress, defaultHTTPServerAddress)
		httpServerAddress = defaultHTTPServerAddress
	}
	return httpServerAddress
}

// StartServer function configures and brings up gRPC server.
func StartServer(gRPCServerAddress, httpServerAddress string, dbClient *ent.Client, wg *sync.WaitGroup, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan chan bool) {
	zlog.Info().Msgf("Starting gRPC server...")

	// get server options
	serverOptions, err := getServerOptions(nil)
	if err != nil {
		zlog.Fatal().Err(err).Msg("Failed to get server options")
	}

	// start server
	serve(gRPCServerAddress, httpServerAddress, dbClient, wg, serverOptions, termChan, readyChan, reverseProxyReadyChan, reverseProxyTermChan)
}
