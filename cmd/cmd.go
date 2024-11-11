package cmd

import (
	"fmt"
	"os"

	new_cmd "github.com/mohamedalosaili/gog/cmd/new"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "gog [command] [flags]",
	Short: "gog is a tool for generating Go projects",
	// Example: "gog new <project name> <path>",
	CompletionOptions: cobra.CompletionOptions{
		DisableDefaultCmd: true,
	},
}

func Run() {
	rootCmd.AddCommand(new_cmd.NewCmd())
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
