package migrate

import (
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func newMigrationUp() *cobra.Command {
	return &cobra.Command{
		Use:                   "up",
		Short:                 "Run pending migrations",
		DisableFlagsInUseLine: true,
		RunE:                  runMigrationUp,
	}
}

func runMigrationUp(cmd *cobra.Command, args []string) error {
	cfg, db, err := setupMigration()
	if err != nil {
		return fmt.Errorf("❌ Failed to setup migration: %v", err)
	}
	defer db.Close()

	goose.SetTableName(cfg.DatabaseMigrateTable)
	fmt.Printf("🔄 Running migrations from directory: %s\n", cfg.DatabaseMigrationsDir)
	if err := goose.Up(db, cfg.DatabaseMigrationsDir); err != nil {
		return fmt.Errorf("❌ Migration failed: %v", err)
	}
	fmt.Println("✅ Migrations completed successfully")

	return nil
}
