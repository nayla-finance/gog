package migrate

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/project-name/internal/config"
)

func Run() {
	fmt.Println("🔄 Loading configuration...")
	cfg, err := config.Load()
	if err != nil {
		fmt.Println("❌ Failed to load configuration")
		panic(err)
	}
	fmt.Println("✅ Configuration loaded successfully")

	fmt.Println("🔄 Connecting to database...")
	db, err := ConnectToDB(cfg)
	if err != nil {
		fmt.Println("❌ Failed to connect to database")
		panic(err)
	}
	defer func() {
		db.Close()
		fmt.Println("✅ Database connection closed")
	}()

	fmt.Println("✅ Database connection established")

	fmt.Printf("🔄 Setting migrations table name to '%s'...\n", cfg.DatabaseMigrateTable)
	goose.SetTableName(cfg.DatabaseMigrateTable)
	fmt.Println("✅ Migrations table name set")

	fmt.Printf("🔄 Running migrations from directory: %s\n", cfg.DatabaseMigrationsDir)
	if err := goose.Up(db, cfg.DatabaseMigrationsDir); err != nil {
		fmt.Println("❌ Migration failed")
		panic(err)
	}
	fmt.Println("✅ Migrations completed successfully")
}

func ConnectToDB(cfg *config.Config) (*sql.DB, error) {
	dbString := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%d sslmode=disable", cfg.DatabaseUsername, cfg.DatabasePassword, cfg.DatabaseName, cfg.DatabaseHost, cfg.DatabasePort)

	return sql.Open(cfg.DatabaseDriver, dbString)
}
