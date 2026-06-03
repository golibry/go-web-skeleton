package testkit

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRemovePaths(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := RemovePaths(path); err != nil {
		t.Fatalf("RemovePaths() error = %v", err)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("Stat() error = %v, want not exists", err)
	}
}

func TestRemovePathsRejectsUnsafePath(t *testing.T) {
	if err := RemovePaths("."); err == nil {
		t.Fatal("RemovePaths() error = nil, want error")
	}
}

func TestEmptyDirectory(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "file.txt")
	if err := os.WriteFile(path, []byte("test"), 0644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	if err := EmptyDirectory(dir); err != nil {
		t.Fatalf("EmptyDirectory() error = %v", err)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("ReadDir() error = %v", err)
	}
	if len(entries) != 0 {
		t.Fatalf("len(entries) = %d, want 0", len(entries))
	}
}
