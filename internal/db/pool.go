package db

import (
	"database/sql"
	"time"

	"github.com/congregalis/aiden/internal/config"
)

type PoolSettings struct {
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

func PoolSettingsFromConfig(cfg config.DatabaseConfig) PoolSettings {
	return PoolSettings{
		MaxOpenConns:    cfg.MaxOpenConns,
		MaxIdleConns:    cfg.MaxIdleConns,
		ConnMaxLifetime: cfg.ConnMaxLifetime,
	}
}

func ApplyPoolSettings(db *sql.DB, settings PoolSettings) {
	if db == nil {
		return
	}

	db.SetMaxOpenConns(settings.MaxOpenConns)
	db.SetMaxIdleConns(settings.MaxIdleConns)
	db.SetConnMaxLifetime(settings.ConnMaxLifetime)
}
