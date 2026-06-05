//go:build ignore

package http

import (
	"net/http"

	frameworkapp "github.com/golibry/go-web-skeleton/framework/app"
)

func RegisterRoutes[C any](container *frameworkapp.Container[C]) func(*http.ServeMux) {
	return func(router *http.ServeMux) {
		router.HandleFunc("GET /health", func(w http.ResponseWriter, _ *http.Request) {
			_ = container.ResponseBuilder().JSON(w, http.StatusOK, map[string]string{
				"status": "ok",
			})
		})
	}
}
