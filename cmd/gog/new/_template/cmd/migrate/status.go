package migrate

import (
	"fmt"

	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func newMigrationStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "status",
		Short:                 "Show the status of the migrations",
		DisableFlagsInUseLine: true,
		RunE:                  runMigrationStatus,
	}

	cmd.Flags().StringP("config", "c", "config.yaml", "config file")

	return cmd
}

func runMigrationStatus(cmd *cobra.Command, args []string) error {
	fmt.Println("🔄 Getting migration status...")

	cfg, db, err := setupMigration(cmd)
	if err != nil {
		return fmt.Errorf("❌ Failed to setup migration: %v", err)
	}
	defer db.Close()

	if err := goose.Status(db, cfg.Database.MigrationsDir); err != nil {
		return fmt.Errorf("❌ Migration failed: %v", err)
	}
	fmt.Println("✅ Migration status retrieved successfully")

	return nil
}
