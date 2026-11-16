package main

import (
	"os"
	"path/filepath"
	"testing"
)

func writeFile(t *testing.T, name, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(name), 0o750); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(name, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
}

func readFile(t *testing.T, name string) string {
	t.Helper()

	b, err := os.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}

	return string(b)
}

func TestCopyDirSkipsGenerated(t *testing.T) {
	t.Parallel()

	src, dst := t.TempDir(), filepath.Join(t.TempDir(), "out")
	writeFile(t, filepath.Join(src, "main.go"), "import \"example.com/old/x\"")
	writeFile(t, filepath.Join(src, "api", "x.pb.go"), "\"\\x14example.com/old/api\"")
	writeFile(t, filepath.Join(src, ".git", "HEAD"), "ref")

	if err := copyDir(src, dst, []string{"example.com/old", "example.com/new"}, []string{".git"}); err != nil {
		t.Fatal(err)
	}

	if got := readFile(t, filepath.Join(dst, "main.go")); got != "import \"example.com/new/x\"" {
		t.Errorf("main.go: %s", got)
	}

	if got := readFile(t, filepath.Join(dst, "api", "x.pb.go")); got != "\"\\x14example.com/old/api\"" {
		t.Errorf("generated file must be kept as is: %s", got)
	}

	if _, err := os.Stat(filepath.Join(dst, ".git")); !os.IsNotExist(err) {
		t.Errorf(".git must be ignored: %v", err)
	}
}

func TestReplaceName(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "chart", "Chart.yaml"), "name: boilerplate")
	writeFile(t, filepath.Join(root, "main.go"), "// boilerplate")
	writeFile(t, filepath.Join(root, ".claude", "skills", "go", "SKILL.md"), "no boilerplate constructors")

	if err := ReplaceName(root, "boilerplate", "billing"); err != nil {
		t.Fatal(err)
	}

	if got := readFile(t, filepath.Join(root, "chart", "Chart.yaml")); got != "name: billing" {
		t.Errorf("Chart.yaml: %s", got)
	}

	if got := readFile(t, filepath.Join(root, "main.go")); got != "// boilerplate" {
		t.Errorf("Go files must not be changed: %s", got)
	}

	if got := readFile(t, filepath.Join(root, ".claude", "skills", "go", "SKILL.md")); got != "no boilerplate constructors" {
		t.Errorf("vendored files must not be changed: %s", got)
	}
}

func TestRemoveTemplateOnly(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, templateIgnoreFile), "# comment\n\ncmd\n.github/workflows/release.yml\n")
	writeFile(t, filepath.Join(root, "cmd", "boilerplate", "main.go"), "")
	writeFile(t, filepath.Join(root, "main.go"), "")
	writeFile(t, filepath.Join(root, ".github", "workflows", "release.yml"), "")
	writeFile(t, filepath.Join(root, ".github", "workflows", "ci.yml"), "")

	if err := RemoveTemplateOnly(root); err != nil {
		t.Fatal(err)
	}

	for p, wantExists := range map[string]bool{
		templateIgnoreFile:              false,
		"cmd":                           false,
		".github/workflows/release.yml": false,
		"main.go":                       true,
		".github/workflows/ci.yml":      true,
	} {
		_, err := os.Stat(filepath.Join(root, p))
		if exists := err == nil; exists != wantExists {
			t.Errorf("%s: exists=%v, want %v", p, exists, wantExists)
		}
	}
}

func TestRemoveTemplateOnlyRejectsEscapes(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, templateIgnoreFile), "../outside\n")

	if err := RemoveTemplateOnly(root); err == nil {
		t.Fatal("expected an error")
	}
}

func TestStripTemplateBlocks(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	writeFile(t, filepath.Join(root, "README.md"), "# x\n\n<!-- template:start -->\nabout the template\n<!-- template:end -->\n\nservice docs\n")
	writeFile(t, filepath.Join(root, "notes.txt"), "<!-- template:start -->keep<!-- template:end -->")

	if err := StripTemplateBlocks(root); err != nil {
		t.Fatal(err)
	}

	if got, want := readFile(t, filepath.Join(root, "README.md")), "# x\n\nservice docs\n"; got != want {
		t.Errorf("README.md: got %q, want %q", got, want)
	}

	if got := readFile(t, filepath.Join(root, "notes.txt")); got != "<!-- template:start -->keep<!-- template:end -->" {
		t.Errorf("non-Markdown files must not change: %q", got)
	}
}
