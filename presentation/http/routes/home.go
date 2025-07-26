package routes

import (
	"fmt"
	"github.com/golibry/go-web-skeleton/infrastructure/registry"
	"net/http"
	"path"
)

type HomeController struct {
	Router    *http.ServeMux
	Container *registry.Container
}

func (c *HomeController) LoadRoutes() {
	c.HelloWorldQuery()
	c.StaticFileQuery()
}

func (c *HomeController) StaticFileQuery() {
	filesPath := path.Join(c.Container.Config().AppBaseDir, "presentation", "http", "ui", "public")
	c.Router.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir(filesPath))),
	)
}

type HelloCommand struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"LastName"`
}

func BuildHelloCommandFromRequest(request *http.Request) HelloCommand {
	firstName := request.FormValue("firstName")
	lastName := request.FormValue("lastName")
	return HelloCommand{firstName, lastName}
}

func (c *HomeController) HelloWorldQuery() {
	c.Router.HandleFunc(
		"/", func(responseWriter http.ResponseWriter, request *http.Request) {
			command := BuildHelloCommandFromRequest(request)

			_, _ = responseWriter.Write(
				[]byte(
					fmt.Sprintf("Hello %s %s!", command.FirstName, command.LastName),
				),
			)
		},
	)
}
