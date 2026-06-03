package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golibry/go-http/http/router/middleware"
	"github.com/golibry/go-web-skeleton/framework/config"
)

type Options struct {
	ServerConfig   config.HttpServer
	Logger         *slog.Logger
	RegisterRoutes func(router *nethttp.ServeMux)

	// BuildGlobalMiddlewareChain wraps the router with middleware components, handlers
	BuildGlobalMiddlewareChain func(
		router *nethttp.ServeMux,
		logger *slog.Logger,
		ctx context.Context,
	) nethttp.Handler
}

func Start(options Options) {
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

	// Initialize HTTP router and register application routes
	router := nethttp.NewServeMux()
	options.RegisterRoutes(router)

	// Start the HTTP server with a graceful shutdown
	addr := net.JoinHostPort(options.ServerConfig.BindAddress, options.ServerConfig.BindPort)
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	var handler nethttp.Handler
	if options.BuildGlobalMiddlewareChain != nil {
		handler = options.BuildGlobalMiddlewareChain(router, options.Logger, serverCtx)
	} else {
		handler = buildGlobalMiddlewareChain(router, options.Logger)
	}

	// Configure an HTTP server with security and performance settings
	server := &nethttp.Server{
		Addr:              addr,
		Handler:           handler,
		MaxHeaderBytes:    options.ServerConfig.MaxHeaderBytes,
		ReadHeaderTimeout: options.ServerConfig.RequestTimeout,
		IdleTimeout:       options.ServerConfig.RequestTimeout,
		ReadTimeout:       options.ServerConfig.RequestTimeout,
	}

	// Set up graceful shutdown handling
	setupGracefulShutdown(
		server, serverCtx, serverStopCtx, options.Logger, options.ServerConfig.RequestTimeout,
	)

	options.Logger.Info("HTTP server started", "address", addr)

	// Start the HTTP server
	err := server.ListenAndServe()
	if err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
		options.Logger.Error("HTTP server failed to start", "error", err, "address", addr)
		serverStopCtx()
		return
	}

	// Wait for a graceful shutdown to complete
	<-serverCtx.Done()
	options.Logger.Info("HTTP server shutdown complete")
}

// buildGlobalMiddlewareChain wraps the router with middleware components from golibry/go-http
func buildGlobalMiddlewareChain(
	router *nethttp.ServeMux,
	logger *slog.Logger,
) nethttp.Handler {
	// Start with the base mux as the handler
	handler := nethttp.Handler(router)

	// Wrap with Access Logger middleware
	accessLogOptions := middleware.AccessLogOptions{
		LogClientIp: true,
	}
	handler = middleware.NewHTTPAccessLogger(handler, logger, accessLogOptions)

	// Wrap with path normalizer
	handler = middleware.NewPathNormalizer(handler)

	// Wrap with Recoverer middleware (outermost layer)
	handler = middleware.NewRecoverer(handler, context.Background(), logger)

	return handler
}

// setupGracefulShutdown configures signal handling for graceful server shutdown.
// It listens for SIGINT, SIGTERM, and SIGQUIT signals.
func setupGracefulShutdown(
	server *nethttp.Server,
	serverCtx context.Context,
	serverStopCtx context.CancelFunc,
	logger *slog.Logger,
	gracePeriod time.Duration,
) {
	// Create a signal channel and register for shutdown signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	go func() {
		defer signal.Stop(sigChan)
		defer close(sigChan)

		// Wait for a shutdown signal or context cancellation
		select {
		case sig := <-sigChan:
			logger.Info("Shutdown signal received", "signal", sig.String())
		case <-serverCtx.Done():
			return
		}

		// Create a shutdown context with timeout
		shutdownCtx, cancel := context.WithTimeout(context.Background(), gracePeriod)
		defer cancel()

		// Monitor for shutdown timeout
		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				logger.Warn("Graceful shutdown timed out", "timeout", gracePeriod)
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
