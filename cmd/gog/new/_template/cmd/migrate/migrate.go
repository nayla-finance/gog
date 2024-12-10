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
		newMigrationDown(),
	)

	return cmd
}

// setupMigration handles common migration setup tasks
func setupMigration(cmd *cobra.Command) (*config.Config, *sql.DB, error) {
	configFile, err := cmd.Flags().GetString("config")
	if err != nil {
		return nil, nil, fmt.Errorf("❌ Failed to get config file: %v", err)
	}

	cfg, err := config.Load(configFile)
	if err != nil {
		return nil, nil, fmt.Errorf("❌ Failed to load configuration: %v", err)
	}

	fmt.Println("🔄 Connecting to database...")

	dbString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=%s", cfg.Database.Username, cfg.Database.Password, cfg.Database.Name, cfg.Database.Host, cfg.Database.Port, cfg.Database.SSLMode)
	db, err := sql.Open(cfg.Database.Driver, dbString)
	if err != nil {
		return nil, nil, fmt.Errorf("❌ Failed to connect to database: %v", err)
	}

	fmt.Println("✅ Database connection established")
	goose.SetTableName(cfg.Database.MigrateTable)

	return cfg, db, nil
}
