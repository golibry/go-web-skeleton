package routes

import (
	"github.com/golibry/go-web-skeleton/infrastructure/registry"
	"net/http"
	"path"
)

type StaticUiController struct {
	Router    *http.ServeMux
	Container *registry.Container
}

func (c *StaticUiController) LoadRoutes() {
	c.StaticFileQuery()
}

func (c *StaticUiController) StaticFileQuery() {
	filesPath := path.Join(c.Container.Config().AppBaseDir, "presentation", "http", "ui", "public")
	c.Router.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir(filesPath))),
	)
}
