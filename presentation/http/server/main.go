package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golibry/go-web-skeleton/infrastructure/registry"
	"github.com/golibry/go-web-skeleton/presentation/http/routes"
)

// main initializes the web server with a dependency injection container,
// sets up routes, and starts the HTTP server with graceful shutdown handling.
func main() {
	// Initialize dependency injection container
	container, err := registry.NewContainer()
	if err != nil {
		panic(fmt.Sprintf("Could not start web server. Error building container registry: %s", err))
	}

	// Ensure proper cleanup of resources on exit
	defer func() {
		if err := container.Close(); err != nil {
			container.Logger().Error("Failed to close container during shutdown", "error", err)
		}
	}()

	container.Logger().Info("Starting web server")

	// Initialize HTTP router and register application routes
	router := http.NewServeMux()
	routes.RegisterRoutes(router, container)

	// Start the HTTP server with a graceful shutdown
	startServer(container, router)
}

// startServer configures and starts the HTTP server with graceful shutdown handling.
// It listens for system signals and performs a graceful shutdown when received.
func startServer(container *registry.Container, router *http.ServeMux) {
	httpConfig := container.Config().HttpServer
	addr := httpConfig.BindAddress + ":" + httpConfig.BindPort
	logger := container.Logger()

	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Configure an HTTP server with security and performance settings
	server := &http.Server{
		Addr:           addr,
		Handler:        router,
		MaxHeaderBytes: httpConfig.MaxHeaderBytes,
		WriteTimeout:   httpConfig.RequestTimeout,
		ReadTimeout:    httpConfig.RequestTimeout,
		IdleTimeout:    httpConfig.RequestTimeout,
	}

	// Set up graceful shutdown handling
	setupGracefulShutdown(server, serverCtx, serverStopCtx, logger)

	logger.Info("HTTP server starting", "address", addr)

	// Start the HTTP server
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("HTTP server failed to start", "error", err, "address", addr)
		return
	}

	// Wait for a graceful shutdown to complete
	<-serverCtx.Done()
	logger.Info("HTTP server shutdown complete")
}

// setupGracefulShutdown configures signal handling for graceful server shutdown.
// It listens for SIGINT, SIGTERM, SIGHUP, and SIGQUIT signals.
func setupGracefulShutdown(
	server *http.Server,
	serverCtx context.Context,
	serverStopCtx context.CancelFunc,
	logger *slog.Logger,
) {
	const gracePeriod = 30 * time.Second

	// Create a signal channel and register for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP, syscall.SIGQUIT)

	go func() {
		// Wait for a shutdown signal
		sig := <-sigChan
		logger.Info("Shutdown signal received", "signal", sig.String())

		// Create a shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(serverCtx, gracePeriod)
		defer cancel()

		// Monitor for shutdown timeout
		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				logger.Warn("Graceful shutdown timed out, forcing exit", "timeout", gracePeriod)
			}
		}()

		// Perform a graceful shutdown
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("Error during server shutdown", "error", err)
		} else {
			logger.Info("Server shutdown initiated successfully")
		}

		// Signal that shutdown is complete
		serverStopCtx()
	}()
}
