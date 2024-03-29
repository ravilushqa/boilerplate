package project

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/ravilushqa/boilerplate/cmd/boilerplate/internal/base"

	"github.com/AlecAivazis/survey/v2"
	"github.com/fatih/color"
)

// Project is a project template.
type Project struct {
	Name string
	Path string
}

// New new a project from remote repo.
func (p *Project) New(ctx context.Context, dir string, layout string, branch string) error {
	to := path.Join(dir, p.Name)
	if _, err := os.Stat(to); !os.IsNotExist(err) {
		fmt.Printf("🚫 %s already exists\n", p.Name)
		override := false
		prompt := &survey.Confirm{
			Message: "📂 Do you want to override the folder ?",
			Help:    "Delete the existing folder and create the project.",
		}
		e := survey.AskOne(prompt, &override)
		if e != nil {
			return e
		}
		if !override {
			return err
		}
		os.RemoveAll(to)
	}
	fmt.Printf("🚀 Creating service %s, layout repo is %s, please wait a moment.\n\n", p.Name, layout)
	repo := base.NewRepo(layout, branch)
	if err := repo.CopyTo(ctx, to, p.Path, []string{".git", ".github"}); err != nil {
		return err
	}

	// replace substring in files in the chart folder, Makefile and README.md
	base.ReplaceInFiles(
		[]string{path.Join(to, "chart"), path.Join(to, "Makefile")},
		[]string{"boilerplate", p.Name},
	)

	// remove cmd folder
	err := os.RemoveAll(path.Join(to, "cmd"))
	if err != nil {
		return err
	}

	base.Tree(to, dir)

	fmt.Printf("\n🍺 Project creation succeeded %s\n", color.GreenString(p.Name))
	fmt.Print("💻 Use the following command to start the project 👇:\n\n")

	fmt.Println(color.WhiteString("$ cd %s", p.Name))
	fmt.Println(color.WhiteString("$ make build"))
	fmt.Println(color.WhiteString("$ ./bin/app"))
	fmt.Println(color.WhiteString("$ Help: ./bin/app --help"))
	return nil
}
