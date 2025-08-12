package routes

import (
	"errors"
	"net/http"
	"path"
	"time"

	"github.com/golibry/go-web-skeleton/infrastructure/registry"
)

// handlerRegistrar is a minimal interface for registering handlers on a mux-like router.
// Both *http.ServeMux and golibry's ServerMuxWrapper satisfy this interface.
type handlerRegistrar interface {
	Handle(pattern string, handler http.Handler)
}

type StaticUiController struct {
	Router    handlerRegistrar
	Container *registry.Container
}

func (c *StaticUiController) LoadRoutes() {
	c.StaticFileQuery()

	// Add demonstration API endpoints showcasing ResponseBuilder usage
	c.ApiDemo()
}

// ApiDemo adds demonstration API endpoints showcasing ResponseBuilder service usage
func (c *StaticUiController) ApiDemo() {
	// JSON response demo
	c.Router.Handle("/api/demo/success", http.HandlerFunc(c.handleSuccess))

	// Text response demo
	c.Router.Handle("/api/demo/text", http.HandlerFunc(c.handleText))

	// HTML response demo
	c.Router.Handle("/api/demo/html", http.HandlerFunc(c.handleHtml))

	// Error response demos
	c.Router.Handle("/api/demo/error/400", http.HandlerFunc(c.handleBadRequest))
	c.Router.Handle("/api/demo/error/404", http.HandlerFunc(c.handleNotFound))
	c.Router.Handle("/api/demo/error/500", http.HandlerFunc(c.handleInternalError))

	// Created response demo
	c.Router.Handle("/api/demo/created", http.HandlerFunc(c.handleCreated))
}

func (c *StaticUiController) StaticFileQuery() {
	filesPath := path.Join(c.Container.Config().AppBaseDir, "presentation", "http", "ui", "public")
	c.Router.Handle(
		"/static/",
		http.StripPrefix("/static/", http.FileServer(http.Dir(filesPath))),
	)
}

// Handler methods demonstrating ResponseBuilder service usage

// handleSuccess demonstrates JSON success response using ResponseBuilder
func (c *StaticUiController) handleSuccess(w http.ResponseWriter, r *http.Request) {
	responseData := map[string]interface{}{
		"message":   "ResponseBuilder service is working correctly",
		"timestamp": time.Now().UTC(),
		"status":    "success",
		"service":   "ResponseBuilder",
	}

	// Use ResponseBuilder service from container
	err := c.Container.ResponseBuilder.Success(w, responseData).Send()
	if err != nil {
		// Fallback error handling
		c.Container.ResponseBuilder.InternalServerError(w, err).Send()
	}
}

// handleText demonstrates plain text response
func (c *StaticUiController) handleText(w http.ResponseWriter, r *http.Request) {
	textContent := "This is a plain text response generated using the ResponseBuilder service!"

	err := c.Container.ResponseBuilder.Text(w).ContentString(textContent).Send()
	if err != nil {
		c.Container.ResponseBuilder.InternalServerError(w, err).Send()
	}
}

// handleHtml demonstrates HTML response
func (c *StaticUiController) handleHtml(w http.ResponseWriter, r *http.Request) {
	htmlContent := `<!DOCTYPE html>
<html>
<head>
    <title>ResponseBuilder Demo</title>
</head>
<body>
    <h1>ResponseBuilder Service Demo</h1>
    <p>This HTML response was generated using the ResponseBuilder service.</p>
    <p>Generated at: ` + time.Now().Format("2006-01-02 15:04:05") + `</p>
</body>
</html>`

	err := c.Container.ResponseBuilder.HTML(w).ContentString(htmlContent).Send()
	if err != nil {
		c.Container.ResponseBuilder.InternalServerError(w, err).Send()
	}
}

// handleBadRequest demonstrates 400 error response
func (c *StaticUiController) handleBadRequest(w http.ResponseWriter, r *http.Request) {
	err := c.Container.ResponseBuilder.BadRequest(
		w,
		"This is a demonstration of a 400 Bad Request error",
	).Send()
	if err != nil {
		// Fallback to basic error handling
		http.Error(w, "Failed to send error response", http.StatusInternalServerError)
	}
}

// handleNotFound demonstrates 404 error response
func (c *StaticUiController) handleNotFound(w http.ResponseWriter, r *http.Request) {
	err := c.Container.ResponseBuilder.NotFound(
		w,
		"The requested demo resource was not found",
	).Send()
	if err != nil {
		http.Error(w, "Failed to send error response", http.StatusInternalServerError)
	}
}

// handleInternalError demonstrates 500 error response with structured error logging
func (c *StaticUiController) handleInternalError(w http.ResponseWriter, r *http.Request) {
	// Simulate an internal error
	simulatedError := errors.New("demonstration of internal server error with structured logging")

	err := c.Container.ResponseBuilder.InternalServerError(w, simulatedError).Send()
	if err != nil {
		http.Error(w, "Failed to send error response", http.StatusInternalServerError)
	}
}

// handleCreated demonstrates 201 Created response
func (c *StaticUiController) handleCreated(w http.ResponseWriter, r *http.Request) {
	createdResource := map[string]interface{}{
		"id":        "demo-123",
		"name":      "Demo Resource",
		"createdAt": time.Now().UTC(),
		"status":    "created",
	}

	err := c.Container.ResponseBuilder.Created(w, createdResource).Send()
	if err != nil {
		c.Container.ResponseBuilder.InternalServerError(w, err).Send()
	}
}
