package migrate

import (
	"database/sql"
	"fmt"

	"github.com/PROJECT_NAME/internal/config"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/spf13/cobra"
)

func NewMigrateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate",
		Short: "Run database migrations",
	}

	cmd.AddCommand(
		newMigrationUp(),
		newMigrationNew(),
		newMigrationStatus(),
	)

	return cmd
}

// setupMigration handles common migration setup tasks
func setupMigration() (*config.Config, *sql.DB, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, nil, fmt.Errorf("❌ Failed to load configuration: %v", err)
	}

	fmt.Println("🔄 Connecting to database...")

	dbString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable", cfg.DatabaseUsername, cfg.DatabasePassword, cfg.DatabaseName, cfg.DatabaseHost, cfg.DatabasePort)
	db, err := sql.Open(cfg.DatabaseDriver, dbString)
	if err != nil {
		return nil, nil, fmt.Errorf("❌ Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Database connection established")
	goose.SetTableName(cfg.DatabaseMigrateTable)

	return cfg, db, nil
}
