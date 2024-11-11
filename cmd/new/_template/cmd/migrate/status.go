package migrate

import (
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func newMigrationStatus() *cobra.Command {
	return &cobra.Command{
		Use:                   "status",
		Short:                 "Show the status of the migrations",
		DisableFlagsInUseLine: true,
		RunE:                  runMigrationStatus,
	}
}

func runMigrationStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Getting migration status...")

	cfg, db, err := setupMigration()
	if err != nil {
		return fmt.Errorf("❌ Failed to setup migration: %v", err)
	}
	defer db.Close()

	if err := goose.Status(db, cfg.DatabaseMigrationsDir); err != nil {
		return fmt.Errorf("❌ Migration failed: %v", err)
	}
	fmt.Println("✅ Migration status retrieved successfully")

	return nil
}
