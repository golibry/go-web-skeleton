package app

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"reflect"
	"sync"

	httplib "github.com/golibry/go-http/http"
	"github.com/golibry/go-web-skeleton/framework/config"
)

type Cleanup func() error
type ContextCleanup func(context.Context) error

type App[C any] struct {
	config   C
	cleanup  []ContextCleanup
	services map[reflect.Type]any
	mu       sync.Mutex
	closed   bool
}

type ContainerOptions struct {
	Log                  config.Log
	Database             *config.Database
	ErrorCategories      func() []*httplib.ErrorCategory
	DisableDefaultLogger bool
	PingDBOnStartup      bool
}

type Container[C any] struct {
	*App[C]
	loggerService   *LoggerService
	dbService       *SQLDBService
	responseBuilder *ResponseBuilder
}

type StandardConfig interface {
	LogConfig() config.Log
	DatabaseConfig() *config.Database
}

func New[C any](config C) *App[C] {
	return &App[C]{
		config: config,
	}
}

func NewContainerFromConfig[C StandardConfig](cfg C) (*Container[C], error) {
	return NewContainer(
		cfg,
		ContainerOptions{
			Log:      cfg.LogConfig(),
			Database: cfg.DatabaseConfig(),
		},
	)
}

func NewContainer[C any](cfg C, options ContainerOptions) (*Container[C], error) {
	root := New(cfg)

	loggerService, err := NewLoggerService(
		options.Log.LogPath,
		options.Log.LogLevel,
		LoggerOptions{
			SetDefault: !options.DisableDefaultLogger,
		},
	)
	if err != nil {
		return nil, err
	}
	root.RegisterCleanup(loggerService.Close)

	container := &Container[C]{
		App:             root,
		loggerService:   loggerService,
		responseBuilder: NewResponseBuilderService(loggerService.Logger(), options.ErrorCategories),
	}
	RegisterService(container, loggerService)
	RegisterService(container, loggerService.Logger())
	RegisterService(container, container.responseBuilder)

	if options.Database != nil {
		dbService, err := NewDBService(
			*options.Database,
			SQLDBOptions{
				PingOnStartup: options.PingDBOnStartup,
			},
		)
		if err != nil {
			_ = container.Close()
			return nil, err
		}
		container.dbService = dbService
		root.RegisterCleanup(dbService.Close)
		RegisterService(container, dbService)
		RegisterService(container, dbService.DB())
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

func (c *Container[C]) DBService() *SQLDBService {
	return c.dbService
}

func (c *Container[C]) DbService() *SQLDBService {
	return c.DBService()
}

func (c *Container[C]) DB() *sql.DB {
	if c == nil || c.dbService == nil {
		return nil
	}

	return c.dbService.DB()
}

func (c *Container[C]) Db() *sql.DB {
	return c.DB()
}

func (c *Container[C]) ResponseBuilder() *ResponseBuilder {
	return c.responseBuilder
}

func RegisterService[C any, T any](container *Container[C], service T) {
	if container == nil || container.App == nil {
		return
	}

	container.mu.Lock()
	defer container.mu.Unlock()

	if container.services == nil {
		container.services = make(map[reflect.Type]any)
	}
	container.services[serviceType[T]()] = service
}

func Service[C any, T any](container *Container[C]) (T, bool) {
	var zero T
	if container == nil || container.App == nil {
		return zero, false
	}

	container.mu.Lock()
	defer container.mu.Unlock()

	service, ok := container.services[serviceType[T]()]
	if !ok {
		return zero, false
	}

	typed, ok := service.(T)
	if !ok {
		return zero, false
	}

	return typed, true
}

func serviceType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func (a *App[C]) RegisterCleanup(cleanup Cleanup) {
	if cleanup == nil {
		return
	}

	a.RegisterContextCleanup(func(context.Context) error {
		return cleanup()
	})
}

func (a *App[C]) RegisterContextCleanup(cleanup ContextCleanup) {
	if a == nil || cleanup == nil {
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	if a.closed {
		return
	}

	a.cleanup = append(a.cleanup, cleanup)
}

func (a *App[C]) Close() error {
	return a.CloseContext(context.Background())
}

func (a *App[C]) CloseContext(ctx context.Context) error {
	if a == nil {
		return nil
	}

	a.mu.Lock()
	if a.closed {
		a.mu.Unlock()
		return nil
	}
	a.closed = true
	cleanup := append([]ContextCleanup(nil), a.cleanup...)
	a.cleanup = nil
	a.mu.Unlock()

	if ctx == nil {
		ctx = context.Background()
	}

	errs := make([]error, 0)
	for i := len(cleanup) - 1; i >= 0; i-- {
		if err := cleanup[i](ctx); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
