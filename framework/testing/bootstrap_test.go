package testkit

import (
	"context"
	"os"
	"testing"

	"github.com/golibry/go-web-skeleton/framework/app"
	"github.com/golibry/go-web-skeleton/framework/config"
)

func TestEnsureTestEnvSetsDefault(t *testing.T) {
	t.Setenv("APP_ENV", "")

	if err := EnsureTestEnv(); err != nil {
		t.Fatalf("EnsureTestEnv() error = %v", err)
	}
	if got := os.Getenv("APP_ENV"); got != "test" {
		t.Fatalf("APP_ENV = %q, want test", got)
	}
}

func TestSetupPassesContextToBeforeSetup(t *testing.T) {
	t.Setenv("APP_ENV", "test")

	type contextKey string
	const key contextKey = "test-key"
	ctx := context.WithValue(context.Background(), key, "expected")
	called := false

	bootstrap, err := Setup(
		Options[struct{}]{
			Context: ctx,
			BeforeSetup: []EnvironmentResetFunc[struct{}]{
				func(ctx context.Context, _ struct{}) error {
					called = true
					if got := ctx.Value(key); got != "expected" {
						t.Fatalf("context value = %v, want expected", got)
					}
					return nil
				},
			},
			ContainerOptions: func(struct{}) app.ContainerOptions {
				return app.ContainerOptions{
					Log: config.Log{
						LogPath: "stdout",
					},
				}
			},
		},
	)
	if err != nil {
		t.Fatalf("Setup() error = %v", err)
	}
	defer func() { _ = bootstrap.Close() }()

	if !called {
		t.Fatal("BeforeSetup hook was not called")
	}
}
