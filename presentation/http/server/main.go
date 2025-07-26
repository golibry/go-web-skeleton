package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golibry/go-web-skeleton/infrastructure/registry"
	"github.com/golibry/go-web-skeleton/presentation/http/routes"
)

func main() {
	baseErr := "Could not start web server."
	container, err := registry.NewContainer()
	if err != nil {
		panic(fmt.Sprintf("%s Error building container registry: %s", baseErr, err))
	}

	fmt.Println("Starting web server")

	router := http.NewServeMux()
	routes.RegisterRoutes(router, container)
	startServer(container, router)
}

func startServer(
	container *registry.Container,
	router *http.ServeMux,
) {
	addr := container.Config().HttpServer.BindAddress + ":" +
		container.Config().HttpServer.BindPort

	// Server run context
	serverCtx, serverStopCtx := context.WithCancel(context.Background())

	// The HTTP Server
	server := &http.Server{
		Addr:           addr,
		Handler:        router,
		MaxHeaderBytes: container.Config().HttpServer.MaxHeaderBytes,
		WriteTimeout:   container.Config().HttpServer.RequestTimeout,
	}

	// Listen for syscall signals for process to interrupt/quit
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	go func() {
		<-sig

		// Shutdown signal with grace period of 30 seconds
		shutdownCtx, gracePeriodCancel := context.WithTimeout(serverCtx, 30*time.Second)
		defer gracePeriodCancel()

		go func() {
			<-shutdownCtx.Done()
			if errors.Is(shutdownCtx.Err(), context.DeadlineExceeded) {
				fmt.Println("Graceful shutdown timed out ... forcing exit")
			}
		}()

		// Trigger graceful shutdown
		err := server.Shutdown(shutdownCtx)
		if err != nil {
			container.Logger().Error(err.Error())
			fmt.Println("[ERROR]", err.Error())
		}
		serverStopCtx()
	}()

	fmt.Println("HTTP server now listening on " + addr)

	// Run the server
	err := server.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		container.Logger().Error(err.Error())
		fmt.Println("[ERROR]", err.Error())
	}

	// Wait for server context to be stopped
	<-serverCtx.Done()
}
