package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestModulePath(t *testing.T) {
	t.Parallel()

	file := filepath.Join(t.TempDir(), "go.mod")
	if err := os.WriteFile(file, []byte("module github.com/ravilushqa/boilerplate\n\ngo 1.27.0\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got, err := ModulePath(file)
	if err != nil {
		t.Fatal(err)
	}

	if want := "github.com/ravilushqa/boilerplate"; got != want {
		t.Fatalf("want %q, got %q", want, got)
	}
}
