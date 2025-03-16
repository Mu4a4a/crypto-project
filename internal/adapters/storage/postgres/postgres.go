package postgres

import (
	"context"
	"crypto-project/config"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
)

func ConnectDB(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	connString := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Postgres.Host, cfg.Postgres.Port, cfg.Postgres.User, cfg.Postgres.Password, cfg.Postgres.DBName, cfg.Postgres.SSLMode)

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new pool connection")
	}

	if err = pool.Ping(ctx); err != nil {
		defer pool.Close()
		return nil, errors.Wrap(err, "failed to ping pool connections")
	}

	return pool, nil
}
