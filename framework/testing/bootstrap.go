package testkit

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/golibry/go-web-skeleton/framework/app"
)

type ConfigLoader[C any] func(config *C) error

type ContainerOptionsFactory[C any] func(config C) app.ContainerOptions

type MigrateFunc[C any] func(context.Context, C, *app.Container[C]) error

type EnvironmentResetFunc[C any] func(context.Context, C) error

type CleanupFunc[C any] func(context.Context, *Bootstrap[C]) error

type Options[C any] struct {
	Context          context.Context
	Config           C
	LoadConfig       ConfigLoader[C]
	ContainerOptions ContainerOptionsFactory[C]
	BeforeSetup      []EnvironmentResetFunc[C]
	Migrate          MigrateFunc[C]
	Cleanup          []CleanupFunc[C]
}

type Bootstrap[C any] struct {
	Config       C
	Container    *app.Container[C]
	cleanupFuncs []CleanupFunc[C]
}

func EnsureTestEnv() error {
	if os.Getenv("APP_ENV") != "" {
		return nil
	}

	return os.Setenv("APP_ENV", "test")
}

func Setup[C any](options Options[C]) (*Bootstrap[C], error) {
	ctx := options.Context
	if ctx == nil {
		ctx = context.Background()
	}

	if err := EnsureTestEnv(); err != nil {
		return nil, fmt.Errorf("could not setup test environment: %w", err)
	}

	cfg := options.Config
	if options.LoadConfig != nil {
		if err := options.LoadConfig(&cfg); err != nil {
			return nil, fmt.Errorf("could not load test config: %w", err)
		}
	}

	if err := runBeforeSetup(ctx, cfg, options.BeforeSetup); err != nil {
		return nil, err
	}

	containerOptions := app.ContainerOptions{}
	if options.ContainerOptions != nil {
		containerOptions = options.ContainerOptions(cfg)
	}

	container, err := app.NewContainer(cfg, containerOptions)
	if err != nil {
		return nil, fmt.Errorf("could not build test container: %w", err)
	}

	bootstrap := &Bootstrap[C]{
		Config:       cfg,
		Container:    container,
		cleanupFuncs: options.Cleanup,
	}

	if options.Migrate != nil {
		if err := options.Migrate(ctx, cfg, container); err != nil {
			_ = bootstrap.Close()
			return nil, fmt.Errorf("could not run test migrations: %w", err)
		}
	}

	return bootstrap, nil
}

func runBeforeSetup[C any](
	ctx context.Context,
	cfg C,
	resetters []EnvironmentResetFunc[C],
) error {
	errs := make([]error, 0)
	for _, reset := range resetters {
		if reset == nil {
			continue
		}
		if err := reset(ctx, cfg); err != nil {
			errs = append(errs, err)
		}
	}

	if err := errors.Join(errs...); err != nil {
		return fmt.Errorf("could not run test setup hooks: %w", err)
	}

	return nil
}

func (b *Bootstrap[C]) Cleanup(ctx context.Context) error {
	if b == nil {
		return nil
	}
	if ctx == nil {
		ctx = context.Background()
	}

	errs := make([]error, 0)
	for _, cleanup := range b.cleanupFuncs {
		if cleanup == nil {
			continue
		}
		if err := cleanup(ctx, b); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}

func (b *Bootstrap[C]) Close() error {
	if b == nil || b.Container == nil {
		return nil
	}

	return b.Container.Close()
}
