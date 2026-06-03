package app

import (
	"database/sql"
	"errors"
	"log/slog"

	httplib "github.com/golibry/go-http/http"
	"github.com/golibry/go-web-skeleton/framework/config"
)

type Cleanup func() error

type App[C any] struct {
	config   C
	cleanup  []Cleanup
	isClosed bool
}

type ContainerOptions struct {
	Log             config.Log
	Database        *config.Database
	ErrorCategories func() []*httplib.ErrorCategory
}

type Container[C any] struct {
	*App[C]
	loggerService   *LoggerService
	dbService       *SqlDbService
	responseBuilder *ResponseBuilder
}

func New[C any](config C) *App[C] {
	return &App[C]{
		config: config,
	}
}

func NewContainer[C any](cfg C, options ContainerOptions) (*Container[C], error) {
	root := New(cfg)

	loggerService, err := NewLoggerService(options.Log.LogPath, options.Log.LogLevel)
	if err != nil {
		return nil, err
	}
	root.RegisterCleanup(loggerService.Close)

	container := &Container[C]{
		App:             root,
		loggerService:   loggerService,
		responseBuilder: NewResponseBuilderService(loggerService.Logger(), options.ErrorCategories),
	}

	if options.Database != nil {
		dbService, err := NewDbService(*options.Database)
		if err != nil {
			_ = container.Close()
			return nil, err
		}
		container.dbService = dbService
		root.RegisterCleanup(dbService.Close)
	}

	return container, nil
}

func (a *App[C]) Config() C {
	return a.config
}

func (c *Container[C]) LoggerService() *LoggerService {
	return c.loggerService
}

func (c *Container[C]) Logger() *slog.Logger {
	if c == nil || c.loggerService == nil {
		return nil
	}

	return c.loggerService.Logger()
}

func (c *Container[C]) DbService() *SqlDbService {
	return c.dbService
}

func (c *Container[C]) Db() *sql.DB {
	if c == nil || c.dbService == nil {
		return nil
	}

	return c.dbService.Db()
}

func (c *Container[C]) ResponseBuilder() *ResponseBuilder {
	return c.responseBuilder
}

func (a *App[C]) RegisterCleanup(cleanup Cleanup) {
	if cleanup == nil {
		return
	}

	a.cleanup = append(a.cleanup, cleanup)
}

func (a *App[C]) Close() error {
	if a == nil || a.isClosed {
		return nil
	}
	a.isClosed = true

	errs := make([]error, 0)
	for i := len(a.cleanup) - 1; i >= 0; i-- {
		if err := a.cleanup[i](); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
