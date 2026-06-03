package migrations

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	"github.com/golibry/go-migrations/execution"
	gomigration "github.com/golibry/go-migrations/migration"
)

func TestMigrationsCommandReturnsExitError(t *testing.T) {
	args := os.Args
	os.Args = []string{"app", "migrations", "missing-command"}
	defer func() { os.Args = args }()

	dirPath, err := gomigration.NewMigrationsDirPath(t.TempDir())
	if err != nil {
		t.Fatalf("NewMigrationsDirPath() error = %v", err)
	}

	originalNewRuntime := newRuntime
	newRuntime = func(Options) (*Runtime, error) {
		return &Runtime{
			Context:    context.Background(),
			DB:         &sql.DB{},
			Repository: &execution.InMemoryRepository{},
			Registry:   gomigration.NewDirMigrationsRegistry(dirPath, nil),
			DirPath:    dirPath,
		}, nil
	}
	defer func() { newRuntime = originalNewRuntime }()

	command := &Migrations{
		Options: Options{},
	}

	err = command.Exec(&bytes.Buffer{})

	var exitErr *MigrationsExitError
	if !errors.As(err, &exitErr) {
		t.Fatalf("Exec() error = %v, want MigrationsExitError", err)
	}
	if exitErr.Code != 1 {
		t.Fatalf("MigrationsExitError.Code = %d, want 1", exitErr.Code)
	}
}
