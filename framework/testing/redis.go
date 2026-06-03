package testkit

import (
	"context"
	"fmt"
)

type RedisClient interface {
	FlushDB(context.Context) error
	Del(context.Context, ...string) error
}

type RedisCleanupOptions[C any] struct {
	Client func(config C, bootstrap *Bootstrap[C]) RedisClient
	Keys   []string
}

func RedisCleaner[C any](options RedisCleanupOptions[C]) CleanupFunc[C] {
	return func(ctx context.Context, bootstrap *Bootstrap[C]) error {
		if options.Client == nil {
			return nil
		}

		client := options.Client(bootstrap.Config, bootstrap)
		if client == nil {
			return nil
		}

		if len(options.Keys) == 0 {
			if err := client.FlushDB(ctx); err != nil {
				return fmt.Errorf("could not flush Redis test database: %w", err)
			}
			return nil
		}

		if err := client.Del(ctx, options.Keys...); err != nil {
			return fmt.Errorf("could not delete Redis test keys: %w", err)
		}

		return nil
	}
}
