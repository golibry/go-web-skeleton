package testkit

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
)

type FilesystemCleanupOptions struct {
	Paths []string
}

func FilesystemCleaner[C any](options FilesystemCleanupOptions) CleanupFunc[C] {
	return func(ctx context.Context, _ *Bootstrap[C]) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		return RemovePaths(options.Paths...)
	}
}

func RemovePaths(paths ...string) error {
	for _, path := range paths {
		cleanPath := filepath.Clean(path)
		if cleanPath == "." || cleanPath == string(filepath.Separator) {
			return fmt.Errorf("refusing to remove unsafe path %q", path)
		}
		if err := os.RemoveAll(cleanPath); err != nil {
			return fmt.Errorf("could not remove path %q: %w", path, err)
		}
	}

	return nil
}

type EmptyDirectoryOptions struct {
	Path string
}

func DirectoryCleaner[C any](options EmptyDirectoryOptions) CleanupFunc[C] {
	return func(ctx context.Context, _ *Bootstrap[C]) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		return EmptyDirectory(options.Path)
	}
}

func EmptyDirectory(path string) error {
	cleanPath := filepath.Clean(path)
	if cleanPath == "." || cleanPath == string(filepath.Separator) {
		return fmt.Errorf("refusing to empty unsafe directory %q", path)
	}

	entries, err := os.ReadDir(cleanPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("could not read directory %q: %w", path, err)
	}

	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(cleanPath, entry.Name())); err != nil {
			return fmt.Errorf("could not remove directory entry %q: %w", entry.Name(), err)
		}
	}

	return nil
}
