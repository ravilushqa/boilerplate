package main

import (
	"log"

	"github.com/ravilushqa/boilerplate/cmd/boilerplate/internal/project"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "boilerplate",
	Short: "Boilerplate: An elegant toolkit for Go microservices.",
}

func init() {
	rootCmd.AddCommand(project.CmdNew)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
