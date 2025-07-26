package routes

import (
	"github.com/golibry/go-web-skeleton/infrastructure/registry"
	"net/http"
)

func RegisterRoutes(router *http.ServeMux, container *registry.Container) {
	(&HomeController{router, container}).LoadRoutes()
}
