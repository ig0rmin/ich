package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/Boostport/migration"
	"github.com/Boostport/migration/driver/postgres"
	_ "github.com/jackc/pgx/v5/stdlib"
)

//go:embed migrations
var embedFS embed.FS

type Config struct {
	Host     string `env:"ICH_DB_HOST, required"`
	Port     string `env:"ICH_DB_PORT, required"`
	User     string `env:"ICH_DB_USER, required"`
	Password string `env:"ICH_DB_PASSWORD, required"`
	Name     string `env:"ICH_DB_NAME, required"`
}

func Connect(cfg *Config) (*sql.DB, error) {
	conStr := fmt.Sprintf("host=%s port=%s user=%s password=%s database=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.Name,
	)
	db, err := sql.Open("pgx", conStr)
	if err != nil {
		return nil, err
	}
	return db, nil
}

func Migrate(db *sql.DB) error {
	driver, err := postgres.NewFromDB(db)
	if err != nil {
		return err
	}
	source := &migration.EmbedMigrationSource{
		EmbedFS: embedFS,
		Dir:     "migrations",
	}
	applied, err := migration.Migrate(driver, source, migration.Up, 0)
	if err != nil {
		return err
	}
	log.Printf("Migrated %v migrations \n", applied)
	return nil
}
