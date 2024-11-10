package migrate

import (
	"fmt"

	"github.com/PROJECT_NAME/internal/config"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func newMigrationNew() *cobra.Command {
	return &cobra.Command{
		Use:                   "new [name]",
		Short:                 "Create a new migration",
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return fmt.Errorf("❌ name is missing")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("❌ Failed to load configuration")
			}

			fmt.Println("🔄 Creating new migration...")
			if err := goose.Create(nil, cfg.DatabaseMigrationsDir, args[0], "sql"); err != nil {
				return fmt.Errorf("❌ Failed to create migration: %v", err)
			}
			fmt.Println("✅ Migration created successfully")

			return nil
		},
	}
}
