package registry

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/golibry/go-common-domain/domain"
	httplib "github.com/golibry/go-http/http"
)

// ResponseBuilder provides a factory service for building HTTP responses
// using the golibry/go-http response builders
type ResponseBuilder struct {
	logger *slog.Logger
	ctx    context.Context
}

// NewResponseBuilderService creates a new response builder service
func NewResponseBuilderService(loggerService *LoggerService) *ResponseBuilder {
	return &ResponseBuilder{
		logger: loggerService.Logger(),
		ctx:    context.Background(),
	}
}

// NewBuilder creates a new base response builder
func (rbs *ResponseBuilder) NewBuilder(w http.ResponseWriter) *httplib.ResponseBuilder {
	return httplib.NewResponseBuilder(w)
}

// NewErrorBuilder Error creates an error response builder with structured logging support
// and predefined error categories
func (rbs *ResponseBuilder) NewErrorBuilder(
	w http.ResponseWriter,
	r *http.Request,
) *httplib.ErrorResponseBuilder {
	return rbs.NewBuilder(w).Error().
		WithLogger(rbs.logger).
		WithErrorCategories(setupErrorCategories()...).
		WithContext(r.Context())
}

func setupErrorCategories() []*httplib.ErrorCategory {
	badRequestCategory := httplib.NewErrorCategory(http.StatusBadRequest)
	httplib.AddErrorType[*domain.Error](badRequestCategory)

	categories := make([]*httplib.ErrorCategory, 0)
	categories = append(categories, badRequestCategory)

	return categories
}
