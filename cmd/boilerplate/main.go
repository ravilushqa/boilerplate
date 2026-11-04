// Command boilerplate creates Go services from the boilerplate template.
//
//	go install github.com/ravilushqa/boilerplate/cmd/boilerplate@latest
//	boilerplate new github.com/acme/billing
package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"

	"github.com/ravilushqa/boilerplate/cmd/boilerplate/internal/project"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	root := &cobra.Command{
		Use:   "boilerplate",
		Short: "Create Go services from the boilerplate template.",
	}
	root.AddCommand(project.NewCmd())

	if err := root.ExecuteContext(ctx); err != nil {
		os.Exit(1) //nolint:gocritic // stop() is not needed on exit
	}
}
