package main

import (
	"context"
	"fmt"
	"io"
	"path"

	"github.com/fatih/color"
)

// templateName is the service name used in the template; it is replaced
// with the new service name in non-Go files (Helm chart, Makefile, docs...).
const templateName = "boilerplate"

// New creates a new project from the template repository.
func (p *Project) New(ctx context.Context, in io.Reader, out io.Writer, dir string, opts options) error {
	to := path.Join(dir, p.Name)

	if err := prepareDir(in, out, to, opts.force); err != nil {
		return err
	}

	_, _ = fmt.Fprintf(out, "🚀 Creating service %s from %s, please wait a moment.\n\n", p.Name, opts.repoURL)

	repo, err := NewRepo(opts.repoURL, opts.branch)
	if err != nil {
		return err
	}

	if err := repo.CopyTo(ctx, to, p.Path, []string{".git"}); err != nil {
		return err
	}

	// Template-only files go first: .templateignore mentions the template name.
	if err := RemoveTemplateOnly(to); err != nil {
		return err
	}

	if err := StripTemplateBlocks(to); err != nil {
		return err
	}

	if err := ReplaceName(to, templateName, p.Name); err != nil {
		return err
	}

	Tree(out, to, dir)

	_, _ = fmt.Fprintf(out, "\n🍺 Project %s created\n", color.GreenString(p.Name))
	_, _ = fmt.Fprint(out, "💻 Next steps:\n\n")
	_, _ = fmt.Fprintln(out, color.WhiteString("$ cd %s", p.Name))
	_, _ = fmt.Fprintln(out, color.WhiteString("$ git init"))
	_, _ = fmt.Fprintln(out, color.WhiteString("$ make generate   # regenerate protobuf code for the new module path"))
	_, _ = fmt.Fprintln(out, color.WhiteString("$ make check      # format, lint and test"))
	_, _ = fmt.Fprintln(out, color.WhiteString("$ make run"))
	_, _ = fmt.Fprintln(out, "\n📖 Read AGENTS.md (also for humans) to learn the layout and conventions.")

	return nil
}
