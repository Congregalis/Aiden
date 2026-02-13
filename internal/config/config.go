package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

const defaultEnvFile = ".env"

type Config struct {
	AppEnv   string
	HTTP     HTTPConfig
	Database DatabaseConfig
	Telegram TelegramConfig
	Log      LogConfig
}

type HTTPConfig struct {
	Port            string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	ShutdownTimeout time.Duration
}

type DatabaseConfig struct {
	DSN             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type TelegramConfig struct {
	BotToken       string
	PollTimeoutSec int
	PollIntervalMS int
	AllowedUpdates string
}

type LogConfig struct {
	Level     string
	AddSource bool
}

func Load(envFilePath string) (Config, error) {
	if envFilePath == "" {
		envFilePath = defaultEnvFile
	}

	if err := loadDotEnv(envFilePath); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load env file: %w", err)
	}

	cfg, err := readFromEnv()
	if err != nil {
		return Config{}, err
	}

	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}

	return cfg, nil
}

func (c Config) Validate() error {
	var missing []string
	if strings.TrimSpace(c.Database.DSN) == "" {
		missing = append(missing, "DB_DSN")
	}
	if strings.TrimSpace(c.Telegram.BotToken) == "" {
		missing = append(missing, "TELEGRAM_BOT_TOKEN")
	}
	if strings.TrimSpace(c.HTTP.Port) == "" {
		missing = append(missing, "HTTP_PORT")
	}

	if len(missing) > 0 {
		return fmt.Errorf("missing required env: %s", strings.Join(missing, ", "))
	}

	if c.Telegram.PollTimeoutSec <= 0 {
		return fmt.Errorf("TELEGRAM_POLL_TIMEOUT_SEC must be > 0")
	}
	if c.Telegram.PollIntervalMS < 0 {
		return fmt.Errorf("TELEGRAM_POLL_INTERVAL_MS must be >= 0")
	}
	if c.HTTP.ReadTimeout <= 0 || c.HTTP.WriteTimeout <= 0 || c.HTTP.ShutdownTimeout <= 0 {
		return fmt.Errorf("HTTP timeouts must be > 0")
	}
	if c.Database.MaxOpenConns <= 0 {
		return fmt.Errorf("DB_MAX_OPEN_CONNS must be > 0")
	}
	if c.Database.MaxIdleConns < 0 {
		return fmt.Errorf("DB_MAX_IDLE_CONNS must be >= 0")
	}
	if c.Database.MaxIdleConns > c.Database.MaxOpenConns {
		return fmt.Errorf("DB_MAX_IDLE_CONNS must be <= DB_MAX_OPEN_CONNS")
	}
	if c.Database.ConnMaxLifetime <= 0 {
		return fmt.Errorf("DB_CONN_MAX_LIFETIME must be > 0")
	}

	return nil
}

func readFromEnv() (Config, error) {
	readTimeout, err := getEnvDuration("HTTP_READ_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	writeTimeout, err := getEnvDuration("HTTP_WRITE_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	shutdownTimeout, err := getEnvDuration("HTTP_SHUTDOWN_TIMEOUT", 10*time.Second)
	if err != nil {
		return Config{}, err
	}

	pollTimeout, err := getEnvInt("TELEGRAM_POLL_TIMEOUT_SEC", 50)
	if err != nil {
		return Config{}, err
	}

	pollInterval, err := getEnvInt("TELEGRAM_POLL_INTERVAL_MS", 200)
	if err != nil {
		return Config{}, err
	}

	maxOpenConns, err := getEnvInt("DB_MAX_OPEN_CONNS", 20)
	if err != nil {
		return Config{}, err
	}

	maxIdleConns, err := getEnvInt("DB_MAX_IDLE_CONNS", 10)
	if err != nil {
		return Config{}, err
	}

	connMaxLifetime, err := getEnvDuration("DB_CONN_MAX_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}

	addSource, err := getEnvBool("LOG_ADD_SOURCE", false)
	if err != nil {
		return Config{}, err
	}

	cfg := Config{
		AppEnv: getEnv("APP_ENV", "development"),
		HTTP: HTTPConfig{
			Port:            getEnv("HTTP_PORT", "8080"),
			ReadTimeout:     readTimeout,
			WriteTimeout:    writeTimeout,
			ShutdownTimeout: shutdownTimeout,
		},
		Database: DatabaseConfig{
			DSN:             getEnv("DB_DSN", ""),
			MaxOpenConns:    maxOpenConns,
			MaxIdleConns:    maxIdleConns,
			ConnMaxLifetime: connMaxLifetime,
		},
		Telegram: TelegramConfig{
			BotToken:       getEnv("TELEGRAM_BOT_TOKEN", ""),
			PollTimeoutSec: pollTimeout,
			PollIntervalMS: pollInterval,
			AllowedUpdates: getEnv("TELEGRAM_ALLOWED_UPDATES", "message"),
		},
		Log: LogConfig{
			Level:     getEnv("LOG_LEVEL", "info"),
			AddSource: addSource,
		},
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) (int, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s as int: %w", key, err)
	}

	return parsed, nil
}

func getEnvDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s as duration: %w", key, err)
	}

	return parsed, nil
}

func getEnvBool(key string, fallback bool) (bool, error) {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("parse %s as bool: %w", key, err)
	}

	return parsed, nil
}
