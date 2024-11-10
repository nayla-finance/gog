package main

import (
	"os"

	"github.com/PROJECT_NAME/cmd/migrate"
	"github.com/PROJECT_NAME/cmd/serve"
	"github.com/spf13/cobra"
)

func main() {
	cmd := &cobra.Command{
		Use:   "PROJECT_NAME",
		Short: "PROJECT_NAME CLI",
		CompletionOptions: cobra.CompletionOptions{
			DisableDefaultCmd: true,
		},
	}

	cmd.AddCommand(serve.NewServeCmd(), migrate.NewMigrateCmd())
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
