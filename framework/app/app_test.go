package app

import (
	"context"
	"log/slog"
	"reflect"
	"sync"
	"testing"

	"github.com/golibry/go-web-skeleton/framework/config"
)

type testStandardConfig struct {
	log config.Log
}

func (c testStandardConfig) LogConfig() config.Log {
	return c.log
}

func (c testStandardConfig) DatabaseConfig() *config.Database {
	return nil
}

func TestCloseRunsContextCleanupOnceInReverseOrder(t *testing.T) {
	app := New(struct{}{})
	order := make([]string, 0)

	app.RegisterContextCleanup(func(ctx context.Context) error {
		if ctx.Value("cleanup") != "yes" {
			t.Fatal("cleanup context value missing")
		}
		order = append(order, "first")
		return nil
	})
	app.RegisterCleanup(func() error {
		order = append(order, "second")
		return nil
	})

	ctx := context.WithValue(context.Background(), "cleanup", "yes")
	if err := app.CloseContext(ctx); err != nil {
		t.Fatalf("CloseContext() error = %v", err)
	}
	if err := app.CloseContext(ctx); err != nil {
		t.Fatalf("second CloseContext() error = %v", err)
	}

	expected := []string{"second", "first"}
	if !reflect.DeepEqual(order, expected) {
		t.Fatalf("cleanup order = %v, want %v", order, expected)
	}
}

func TestCloseIsConcurrencySafe(t *testing.T) {
	app := New(struct{}{})
	var mu sync.Mutex
	calls := 0
	app.RegisterCleanup(func() error {
		mu.Lock()
		defer mu.Unlock()
		calls++
		return nil
	})

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := app.Close(); err != nil {
				t.Errorf("Close() error = %v", err)
			}
		}()
	}
	wg.Wait()

	if calls != 1 {
		t.Fatalf("cleanup calls = %d, want 1", calls)
	}
}

func TestContainerRegistersCoreServices(t *testing.T) {
	container, err := NewContainer(
		struct{}{},
		ContainerOptions{
			Log: config.Log{
				LogLevel: slog.LevelInfo,
				LogPath:  "stdout",
			},
			DisableDefaultLogger: true,
		},
	)
	if err != nil {
		t.Fatalf("NewContainer() error = %v", err)
	}
	defer func() { _ = container.Close() }()

	logger, ok := Service[struct{}, *slog.Logger](container)
	if !ok || logger == nil {
		t.Fatal("slog.Logger service was not registered")
	}

	responseBuilder, ok := Service[struct{}, *ResponseBuilder](container)
	if !ok || responseBuilder == nil {
		t.Fatal("ResponseBuilder service was not registered")
	}
}

func TestNewContainerFromConfigUsesStandardConfig(t *testing.T) {
	container, err := NewContainerFromConfig(
		testStandardConfig{
			log: config.Log{
				LogLevel: slog.LevelInfo,
				LogPath:  "stdout",
			},
		},
	)
	if err != nil {
		t.Fatalf("NewContainerFromConfig() error = %v", err)
	}
	defer func() { _ = container.Close() }()

	if container.Logger() == nil {
		t.Fatal("Logger() = nil")
	}
}
