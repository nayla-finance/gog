package main

import (
	"fmt"
	"os"

	"github.com/mohamedalosaili/gog"
	new_cmd "github.com/mohamedalosaili/gog/cmd/new"
	"github.com/mohamedalosaili/gog/cmd/swag"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:     "gog [command]",
	Short:   "gog is a tool for generating Go projects",
	Version: gog.Version,
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func main() {
	rootCmd.AddCommand(new_cmd.NewCmd(), swag.NewSwag())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
