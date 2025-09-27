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

	"github.com/golibry/go-http/http/router/middleware"
	"github.com/golibry/go-web-skeleton/config"
)

type ServerOptions struct {
	ServerConfig   config.HttpServerConfig
	Logger         *slog.Logger
	RegisterRoutes func(router *http.ServeMux)
}

func StartServer(options ServerOptions) {
	if options.RegisterRoutes == nil {
		panic(
			fmt.Errorf(
				"failed to start web server: %s",
				"register routes function is required",
			),
		)
	}

	if options.Logger == nil {
		panic(
			fmt.Errorf(
				"failed to start web server: %s",
				"logger is required",
			),
		)
	}

	options.Logger.Info("Starting web server")

	// Initialize HTTP router and register application routes
	router := http.NewServeMux()
	options.RegisterRoutes(router)

	// Start the HTTP server with a graceful shutdown
	addr := options.ServerConfig.BindAddress + ":" + options.ServerConfig.BindPort
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// Configure an HTTP server with security and performance settings
	server := &http.Server{
		Addr:              addr,
		Handler:           buildGlobalMiddlewareChain(router, options.Logger, serverCtx),
		MaxHeaderBytes:    options.ServerConfig.MaxHeaderBytes,
		ReadHeaderTimeout: options.ServerConfig.RequestTimeout,
		IdleTimeout:       options.ServerConfig.RequestTimeout,
	}

	// Set up graceful shutdown handling
	setupGracefulShutdown(
		server, serverCtx, serverStopCtx, options.Logger, options.ServerConfig.RequestTimeout,
	)

	// Start the HTTP server
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		options.Logger.Error("HTTP server failed to start", "error", err, "address", addr)
		return
	}

	options.Logger.Info("HTTP server started", "address", addr)

	// Wait for a graceful shutdown to complete
	<-serverCtx.Done()
	options.Logger.Info("HTTP server shutdown complete")
}

// buildGlobalMiddlewareChain wraps the router with middleware components from golibry/go-http
func buildGlobalMiddlewareChain(
	router *http.ServeMux,
	logger *slog.Logger,
	ctx context.Context,
) http.Handler {
	// Start with the base mux as the handler
	handler := http.Handler(router)

	// Wrap with Access Logger middleware
	accessLogOptions := middleware.AccessLogOptions{
		LogClientIp: true,
	}
	handler = middleware.NewHTTPAccessLogger(handler, logger, accessLogOptions)

	// Wrap with path normalizer
	handler = middleware.NewPathNormalizer(handler)

	// Wrap with Recoverer middleware (outermost layer)
	handler = middleware.NewRecoverer(handler, ctx, logger)

	return handler
}

// setupGracefulShutdown configures signal handling for graceful server shutdown.
// It listens for SIGINT, SIGTERM, SIGHUP, and SIGQUIT signals.
func setupGracefulShutdown(
	server *http.Server,
	serverCtx context.Context,
	serverStopCtx context.CancelFunc,
	logger *slog.Logger,
	gracePeriod time.Duration,
) {
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
