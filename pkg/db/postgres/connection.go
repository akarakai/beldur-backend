package postgres

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPgxPool(ctx context.Context, cfg Config) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}

type Config struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

func (c Config) DSN() string {
	ssl := c.SSLMode
	if ssl == "" {
		ssl = "disable"
	}
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		c.User, c.Password, c.Host, c.Port, c.DBName, ssl,
	)
}

func ConfigFromEnv() (Config, error) {
	cfg := Config{
		DBName:   os.Getenv("POSTGRES_DB"),
		User:     os.Getenv("POSTGRES_USER"),
		Password: os.Getenv("POSTGRES_PASSWORD"),
		Host:     os.Getenv("POSTGRES_HOST"),
		Port:     os.Getenv("POSTGRES_PORT"),
		SSLMode:  os.Getenv("POSTGRES_SSLMODE"),
	}

	var missing []string
	if cfg.DBName == "" {
		missing = append(missing, "POSTGRES_DB")
	}
	if cfg.User == "" {
		missing = append(missing, "POSTGRES_USER")
	}
	if cfg.Password == "" {
		missing = append(missing, "POSTGRES_PASSWORD")
	}
	if cfg.Host == "" {
		missing = append(missing, "POSTGRES_HOST")
	}
	if cfg.Port == "" {
		missing = append(missing, "POSTGRES_PORT")
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing env vars: %v", missing)
	}

	return cfg, nil
}
