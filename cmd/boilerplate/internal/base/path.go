package base

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/fatih/color"
)

// templateIgnoreFile lists paths (one per line, relative to the repository
// root) that belong to the template repository only and are removed from
// generated projects. The file itself is removed too.
const templateIgnoreFile = ".templateignore"

// cacheDir returns (and creates) a directory under ~/.boilerplate.
func cacheDir(dir string) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	p := path.Join(home, ".boilerplate", dir)
	if err := os.MkdirAll(p, 0o700); err != nil {
		return "", err
	}

	return p, nil
}

// isGenerated reports whether a file is generated code that must not be
// edited textually. Protobuf descriptors embedded in *.pb.go are
// length-prefixed, so a string replacement would corrupt them;
// such files are regenerated with `make generate` instead.
func isGenerated(name string) bool {
	return strings.HasSuffix(name, ".pb.go")
}

func copyFile(src, dst string, replaces []string) error {
	srcinfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	buf, err := os.ReadFile(src)
	if err != nil {
		return err
	}

	if !isGenerated(src) {
		buf = replace(buf, replaces)
	}

	return os.WriteFile(dst, buf, srcinfo.Mode()) //nolint:gosec // paths come from the template being copied
}

// replace applies old/new pairs in order.
func replace(buf []byte, replaces []string) []byte {
	for i := 0; i+1 < len(replaces); i += 2 {
		buf = bytes.ReplaceAll(buf, []byte(replaces[i]), []byte(replaces[i+1]))
	}

	return buf
}

func copyDir(src, dst string, replaces, ignores []string) error {
	srcinfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dst, srcinfo.Mode()); err != nil {
		return err
	}

	fds, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, fd := range fds {
		if slices.Contains(ignores, fd.Name()) {
			continue
		}

		srcfp := path.Join(src, fd.Name())
		dstfp := path.Join(dst, fd.Name())

		if fd.IsDir() {
			err = copyDir(srcfp, dstfp, replaces, ignores)
		} else {
			err = copyFile(srcfp, dstfp, replaces)
		}

		if err != nil {
			return err
		}
	}

	return nil
}

// Tree prints the files created under root, relative to dir.
func Tree(out io.Writer, root, dir string) {
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil //nolint:nilerr // best-effort listing
		}

		info, err := d.Info()
		if err != nil {
			return nil //nolint:nilerr // best-effort listing
		}

		_, _ = fmt.Fprintf(out, "%s %s (%v bytes)\n", color.GreenString("CREATED"), strings.TrimPrefix(p, dir+"/"), info.Size())

		return nil
	})
}

// ReplaceName replaces the template service name with the new one in all
// non-Go files under root (Helm chart, Makefile, Dockerfile, docs, ...).
// Go code refers to the module path, which is replaced separately.
func ReplaceName(root, oldName, newName string) error {
	return filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			return nil
		}

		switch filepath.Ext(p) {
		case ".go", ".mod", ".sum":
			return nil
		}

		return copyFile(p, p, []string{oldName, newName})
	})
}

var templateBlockRe = regexp.MustCompile(`(?s)<!-- template:start -->.*?<!-- template:end -->\n*`)

// StripTemplateBlocks removes text between `<!-- template:start -->` and
// `<!-- template:end -->` markers from Markdown files under root.
func StripTemplateBlocks(root string) error {
	r, err := os.OpenRoot(root)
	if err != nil {
		return err
	}

	defer r.Close()

	return fs.WalkDir(r.FS(), ".", func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || filepath.Ext(p) != ".md" {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return err
		}

		b, err := r.ReadFile(p)
		if err != nil {
			return err
		}

		stripped := templateBlockRe.ReplaceAll(b, nil)
		if bytes.Equal(b, stripped) {
			return nil
		}

		return r.WriteFile(p, stripped, info.Mode())
	})
}

// RemoveTemplateOnly removes the paths listed in .templateignore under root.
func RemoveTemplateOnly(root string) error {
	f, err := os.Open(filepath.Join(root, templateIgnoreFile))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	if err != nil {
		return err
	}

	defer f.Close()

	paths := []string{templateIgnoreFile}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		if !filepath.IsLocal(line) {
			return fmt.Errorf("%s: path %q must be relative and inside the repository", templateIgnoreFile, line)
		}

		paths = append(paths, line)
	}

	if err := sc.Err(); err != nil {
		return err
	}

	for _, p := range paths {
		if err := os.RemoveAll(filepath.Join(root, p)); err != nil {
			return err
		}
	}

	return nil
}
