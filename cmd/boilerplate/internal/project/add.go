package project

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/fatih/color"

	"github.com/ravilushqa/boilerplate/cmd/boilerplate/internal/base"
)

// repoAddIgnores are template files that make no sense inside an existing module.
var repoAddIgnores = []string{
	".git", ".github", ".claude", "api", "README.md", "AGENTS.md", "CLAUDE.md", "LICENSE",
	"go.mod", "go.sum", "tools", "buf.yaml", "buf.gen.yaml", ".golangci.yml", "renovate.json", ".gitignore",
}

// Add adds the service into the existing Go module mod.
func (p *Project) Add(ctx context.Context, in io.Reader, out io.Writer, dir, mod string, opts options) error {
	to := path.Join(dir, p.Path)

	if err := prepareDir(in, out, to, opts.force); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(out, "🚀 Adding service %s from %s, please wait a moment.\n\n", p.Name, opts.repoURL)

	repo, err := base.NewRepo(opts.repoURL, opts.branch)
	if err != nil {
		return err
	}

	if err := repo.CopyToV2(ctx, to, path.Join(mod, p.Path), repoAddIgnores, []string{path.Join(p.Path, "api"), "api"}); err != nil {
		return err
	}

	if err := base.RemoveTemplateOnly(to); err != nil {
		return err
	}

	base.Tree(out, to, dir)

	_, _ = fmt.Fprintf(out, "\n🍺 Service %s added\n", color.GreenString(p.Name))
	_, _ = fmt.Fprintln(out, color.WhiteString("$ go build ./%s/...", p.Path))

	return nil
}
