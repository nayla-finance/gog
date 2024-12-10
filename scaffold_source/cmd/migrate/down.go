package migrate

import (
	"fmt"
	"strconv"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func newMigrationDown() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "down [version]",
		Short: "Rollback migrations",
		Long: `Rollback migrations
migrate down - Rollback single migration
migrate down 20241108133703 - Rollback to version 20241108133703
migrate down 0 - Rollback all migrations`,
		Example: "migrate down\nmigrate down 20241108133703\nmigrate down 0",
		RunE:    runMigrationDown,
	}

	cmd.Flags().StringP("config", "c", "config.yaml", "config file")

	return cmd
}

func runMigrationDown(cmd *cobra.Command, args []string) error {
	cfg, db, err := setupMigration(cmd)
	if err != nil {
		return fmt.Errorf("❌ Failed to setup migration: %v", err)
	}
	defer db.Close()

	goose.SetTableName(cfg.Database.MigrateTable)
	fmt.Printf("🔄 Rolling back migrations from directory: %s\n", cfg.Database.MigrationsDir)

	if len(args) == 0 {
		if err := goose.Down(db, cfg.Database.MigrationsDir); err != nil {
			return fmt.Errorf("❌ Rolling back failed: %v", err)
		}
	} else {
		to, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil {
			return fmt.Errorf("❌ Invalid migration version: %v", err)
		}

		if err := goose.DownTo(db, cfg.Database.MigrationsDir, to); err != nil {
			return fmt.Errorf("❌ Rolling back failed: %v", err)
		}
	}
	fmt.Println("✅ Migrations completed successfully")

	return nil
}
