package migrate

import (
	"fmt"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func newMigrationNew() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "new [name]",
		Short:                 "Create a new migration",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("❌ name is missing")
			}

			configFile, err := cmd.Flags().GetString("config")
			if err != nil {
				return fmt.Errorf("❌ Failed to get config file: %v", err)
			}

			cfg, err := config.Load(configFile)
			if err != nil {
				return fmt.Errorf("❌ Failed to load configuration")
			}

			fmt.Println("🔄 Creating new migration...")
			if err := goose.Create(nil, cfg.Database.MigrationsDir, args[0], "sql"); err != nil {
				return fmt.Errorf("❌ Failed to create migration: %v", err)
			}
			fmt.Println("✅ Migration created successfully")

			return nil
		},
	}

	cmd.Flags().StringP("config", "c", "config.yaml", "config file")

	return cmd
}
