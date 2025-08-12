package routes

import (
	"net/http"

	"github.com/golibry/go-http/http/router"
	"github.com/golibry/go-http/http/router/middleware"
	"github.com/golibry/go-web-skeleton/infrastructure/registry"
)

// RegisterRoutes registers all application routes on a dedicated ServerMuxWrapper
// and mounts it under "/" on the provided root router. All registered routes
// get the project's default middlewares by default via the wrapper.
//
// Default middlewares (inner to outer):
// - Timeout (from golibry/go-http)
func RegisterRoutes(root *http.ServeMux, container *registry.Container) {
	logger := container.Logger()

	// Build default middleware list for the mux wrapper
	defaultMiddlewares := []router.NamedMiddleware{
		{
			Name: "timeout",
			Middleware: func(next http.Handler) http.Handler {
				return middleware.NewTimeoutMiddleware(
					next,
					logger,
					middleware.TimeoutOptions{
						Timeout: container.Config().HttpServer.RequestTimeout,
					},
				)
			},
		},
	}

	// Create mux wrapper with defaults
	mux := router.NewServerMuxWrapper(defaultMiddlewares)

	// Load all feature routes on the mux wrapper
	(&StaticUiController{Router: mux, Container: container}).LoadRoutes()

	// Mount the mux wrapper under a root path so every route is covered
	root.Handle("/", mux)
}
