package base

import (
	"path/filepath"
	"testing"
)

func TestRepo(t *testing.T) {
	urls := []string{
		"ssh://git@github.com:7875/ravilushqa/boilerplate.git",
		"git://github.com:7875/ravilushqa/boilerplate.git",
		"https://github.com:7875/ravilushqa/boilerplate.git",
		"ftps://github.com:7875/ravilushqa/boilerplate.git",
		"git@github.com:ravilushqa/boilerplate.git",
		"ssh://git@github.com:7875/ravilushqa/boilerplate.git",
		"git://github.com:7875/ravilushqa/boilerplate.git",
		"git@github.com:ravilushqa/boilerplate.git",
		"//github.com/ravilushqa/boilerplate.git",
		"file://./github.com/ravilushqa/boilerplate.git",
	}
	for _, url := range urls {
		dir := repoDir(url)
		if dir != "github.com/ravilushqa" && dir != "/ravilushqa" {
			t.Fatal(url, "repoDir test failed", dir)
		}
	}
}

func TestRepoClone(t *testing.T) {
	if testing.Short() {
		t.Skip("requires network access")
	}

	r, err := NewRepo("https://github.com/ravilushqa/boilerplate.git", "")
	if err != nil {
		t.Fatal(err)
	}

	if err := r.Clone(t.Context()); err != nil {
		t.Fatal(err)
	}

	if err := r.CopyTo(t.Context(), filepath.Join(t.TempDir(), "repo"), "github.com/ravilushqa/boilerplate-layout", nil); err != nil {
		t.Fatal(err)
	}
}
