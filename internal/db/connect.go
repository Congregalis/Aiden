package db

import (
	"database/sql"
	"fmt"

	"github.com/congregalis/aiden/internal/config"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func Open(databaseConfig config.DatabaseConfig) (*sql.DB, error) {
	db, err := sql.Open("pgx", databaseConfig.DSN)
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	ApplyPoolSettings(db, PoolSettingsFromConfig(databaseConfig))
	return db, nil
}
