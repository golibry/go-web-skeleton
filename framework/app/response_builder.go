package app

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	"github.com/golibry/go-common-domain/domain"
	httplib "github.com/golibry/go-http/http"
)

type ResponseBuilder struct {
	logger          *slog.Logger
	errorCategories func() []*httplib.ErrorCategory
}

func NewResponseBuilderService(
	logger *slog.Logger,
	errorCategories func() []*httplib.ErrorCategory,
) *ResponseBuilder {
	if errorCategories == nil {
		errorCategories = defaultErrorCategories
	}

	return &ResponseBuilder{
		logger:          logger,
		errorCategories: errorCategories,
	}
}

func (rbs *ResponseBuilder) NewBuilder(w http.ResponseWriter) *httplib.ResponseBuilder {
	return httplib.NewResponseBuilder(w)
}

func (rbs *ResponseBuilder) NewErrorBuilder(
	responseWriter http.ResponseWriter,
	request *http.Request,
) *httplib.ErrorResponseBuilder {
	return rbs.NewBuilder(responseWriter).Error().
		WithLogger(rbs.logger).
		WithErrorCategories(rbs.errorCategories()...).
		WithContext(request.Context())
}

func (rbs *ResponseBuilder) JSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}

func (rbs *ResponseBuilder) Text(w http.ResponseWriter, status int, s string) error {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	_, err := io.WriteString(w, s)
	return err
}

func (rbs *ResponseBuilder) HTML(w http.ResponseWriter, status int, s string) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, err := io.WriteString(w, s)
	return err
}

func (rbs *ResponseBuilder) NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

func (rbs *ResponseBuilder) Redirect(
	w http.ResponseWriter,
	r *http.Request,
	url string,
	status int,
) {
	if status == 0 {
		status = http.StatusFound
	}

	http.Redirect(w, r, url, status)
}

func defaultErrorCategories() []*httplib.ErrorCategory {
	badRequestCategory := httplib.NewErrorCategory(http.StatusBadRequest)
	httplib.AddErrorType[*domain.Error](badRequestCategory)

	categories := make([]*httplib.ErrorCategory, 0)
	categories = append(categories, badRequestCategory)

	return categories
}
