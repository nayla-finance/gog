package migrate

import (
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func newMigrationUp() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "up",
		Short:                 "Run pending migrations",
		DisableFlagsInUseLine: true,
		RunE:                  runMigrationUp,
	}

	cmd.Flags().StringP("config", "c", "config.yaml", "config file")

	return cmd
}

func runMigrationUp(cmd *cobra.Command, args []string) error {
	cfg, db, err := setupMigration(cmd)
	if err != nil {
		return fmt.Errorf("❌ Failed to setup migration: %v", err)
	}
	defer db.Close()

	goose.SetTableName(cfg.Database.MigrateTable)
	fmt.Printf("🔄 Running migrations from directory: %s\n", cfg.Database.MigrationsDir)
	if err := goose.Up(db, cfg.Database.MigrationsDir); err != nil {
		return fmt.Errorf("❌ Migration failed: %v", err)
	}
	fmt.Println("✅ Migrations completed successfully")

	return nil
}
