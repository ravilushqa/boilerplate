// Package project implements the `boilerplate new` command.
package project

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/ravilushqa/boilerplate/cmd/boilerplate/internal/base"
)

const defaultRepoURL = "https://github.com/ravilushqa/boilerplate.git"

// Project is a project template.
type Project struct {
	Name string // directory and service name, e.g. "billing"
	Path string // module path or path relative to the current module, e.g. "github.com/acme/billing"
}

type options struct {
	repoURL string
	branch  string
	timeout time.Duration
	nomod   bool
	force   bool
}

// NewCmd returns the `new` command.
func NewCmd() *cobra.Command {
	opts := options{
		repoURL: defaultRepoURL,
		timeout: 60 * time.Second,
	}
	if v := os.Getenv("LAYOUT_REPO"); v != "" {
		opts.repoURL = v
	}

	cmd := &cobra.Command{
		Use:   "new <module-path>",
		Short: "Create a service from the template",
		Long: `Create a new service from the template repository.

The argument is the Go module path of the new service; its last element
becomes the directory and service name.`,
		Example: `  boilerplate new github.com/acme/billing
  boilerplate new github.com/acme/billing --force   # overwrite without asking (CI, AI agents)
  boilerplate new internal/billing --nomod          # add to the current module`,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: false,
		RunE: func(cmd *cobra.Command, args []string) error {
			return run(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), args[0], opts)
		},
	}

	f := cmd.Flags()
	f.StringVarP(&opts.repoURL, "repo-url", "r", opts.repoURL, "template repository (env LAYOUT_REPO)")
	f.StringVarP(&opts.branch, "branch", "b", opts.branch, "template repository branch")
	f.DurationVarP(&opts.timeout, "timeout", "t", opts.timeout, "timeout for fetching the template")
	f.BoolVar(&opts.nomod, "nomod", opts.nomod, "add the service into the current Go module instead of creating a new one")
	f.BoolVarP(&opts.force, "force", "f", opts.force, "overwrite an existing directory without asking")

	return cmd
}

func run(ctx context.Context, in io.Reader, out io.Writer, name string, opts options) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, opts.timeout)
	defer cancel()

	p := &Project{Name: path.Base(name), Path: name}

	if !opts.nomod {
		err = p.New(ctx, in, out, wd, opts)
	} else {
		var mod string

		mod, err = base.ModulePath(path.Join(wd, "go.mod"))
		if err != nil {
			return fmt.Errorf("--nomod requires go.mod in the current directory: %w", err)
		}

		err = p.Add(ctx, in, out, wd, mod, opts)
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return fmt.Errorf("project creation timed out after %s (use --timeout): %w", opts.timeout, err)
	}

	return err
}

// prepareDir makes sure dir does not exist, removing it if allowed.
func prepareDir(in io.Reader, out io.Writer, dir string, force bool) error {
	_, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	if !force {
		if !isTerminal(in) {
			return fmt.Errorf("%s already exists; pass --force to overwrite it", dir)
		}

		ok, err := confirm(in, out, fmt.Sprintf("📂 %s already exists. Delete it and create the project?", dir))
		if err != nil {
			return err
		}

		if !ok {
			return fmt.Errorf("%s already exists", dir)
		}
	}

	return os.RemoveAll(dir)
}

func confirm(in io.Reader, out io.Writer, question string) (bool, error) {
	_, _ = fmt.Fprintf(out, "%s [y/N]: ", question)

	answer, err := bufio.NewReader(in).ReadString('\n')
	if err != nil && !errors.Is(err, io.EOF) {
		return false, err
	}

	switch strings.ToLower(strings.TrimSpace(answer)) {
	case "y", "yes":
		return true, nil
	default:
		return false, nil
	}
}

func isTerminal(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}

	fi, err := f.Stat()

	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
