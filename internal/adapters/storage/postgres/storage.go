package postgres

import (
	"context"
	"fmt"
	"go.uber.org/zap"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"

	"crypto-project/internal/cases"
	"crypto-project/internal/entities"
)

type Storage struct {
	pool   *pgxpool.Pool
	logger cases.Logger
}

type Config struct {
	host     string
	port     string
	database string
	user     string
	password string
}

type Options func(*Config)

func NewStorage(logger cases.Logger, options ...Options) (*Storage, error) {
	config := &Config{}

	for _, opt := range options {
		opt(config)
	}

	if config.host == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "host cannot be empty")
	}

	if config.port == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "port cannot be empty")
	}

	if config.database == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "database name cannot be empty")
	}

	if config.user == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "user cannot be empty")
	}

	if config.password == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "password cannot be empty")
	}

	if logger == nil || logger == cases.Logger(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "logger not set")
	}

	connString := fmt.Sprintf("host=%s port=%s database=%s user=%s password=%s", config.host, config.port, config.database, config.user, config.password)

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, errors.Wrapf(entities.ErrInternal, "failed to create new pgxpool: %v", err)
	}

	return &Storage{
		pool:   pool,
		logger: logger,
	}, nil
}

func WithHost(host string) Options {
	return func(config *Config) {
		config.host = host
	}
}

func WithPort(port string) Options {
	return func(config *Config) {
		config.port = port
	}
}

func WithDatabase(database string) Options {
	return func(config *Config) {
		config.database = database
	}
}

func WithUser(user string) Options {
	return func(config *Config) {
		config.user = user
	}
}

func WithPassword(password string) Options {
	return func(config *Config) {
		config.password = password
	}
}

func (s *Storage) Store(ctx context.Context, coins []*entities.Coin) error {
	query := `INSERT INTO coins (title, cost, actual_at) values ($1, $2, $3);`

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		s.logger.Error("failed to begin tx",
			zap.String("method", "Store"),
			zap.Error(err))

		return errors.Wrapf(entities.ErrInternal, "failed to begin a transaction: %v", err)
	}
	defer tx.Rollback(ctx)

	for _, coin := range coins {
		params := []any{coin.Title, coin.Cost, coin.ActualAt}
		_, err = s.pool.Exec(ctx, query, params...)
		if err != nil {
			s.logger.Error("failed to exec coin",
				zap.String("method", "Store"),
				zap.Any("params query", params),
				zap.Error(err))

			return errors.Wrapf(entities.ErrInternal, "failed to execute query: %v", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		s.logger.Error("failed to commit tx",
			zap.String("method", "Store"),
			zap.Error(err))

		return errors.Wrapf(entities.ErrInternal, "failed to commit a transaction: %v", err)
	}

	return nil
}

func (s *Storage) GetCoinsList(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT title FROM coins;`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		s.logger.Error("failed to process query",
			zap.String("method", "GetCoinsList"),
			zap.String("query", query),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "failed to execute query: %v", err)
	}
	defer rows.Close()

	var titles []string

	for rows.Next() {
		var title string

		if err = rows.Scan(&title); err != nil {
			s.logger.Error("failed to scan titles",
				zap.String("method", "GetCoinsList"),
				zap.Error(err))

			return nil, errors.Wrapf(entities.ErrInternal, "failed to scan row: %v", err)
		}

		titles = append(titles, title)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("failed to check errors",
			zap.String("method", "GetCoinsList"),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "error during to row iteration: %v", err)
	}

	return titles, nil
}

func (s *Storage) GetActualCoin(ctx context.Context, titles []string) ([]*entities.Coin, error) {
	query := `SELECT DISTINCT ON (title) title, cost, actual_at FROM coins WHERE title = ANY($1) ORDER BY title, actual_at DESC;`

	rows, err := s.pool.Query(ctx, query, titles)
	if err != nil {
		s.logger.Error("failed to process query",
			zap.String("method", "GetActualCoin"),
			zap.String("query", query),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "failed to execute query: %v", err)
	}
	defer rows.Close()

	actualCoins := make([]*entities.Coin, 0, len(titles))

	for rows.Next() {
		var title string
		var cost float64
		var actualAt time.Time

		if err = rows.Scan(&title, &cost, &actualAt); err != nil {
			s.logger.Error("failed to scan coin data",
				zap.String("method", "GetActualCoin"),
				zap.Error(err))

			return nil, errors.Wrapf(entities.ErrInternal, "failed to scan row: %v", err)
		}

		coin, err := entities.NewCoin(title, cost, actualAt)
		if err != nil {
			s.logger.Error("failed to create new coin",
				zap.String("method", "GetActualCoin"),
				zap.String("title", title),
				zap.Float64("cost", cost),
				zap.Time("actual at", actualAt),
				zap.Error(err))

			return nil, errors.Wrap(err, "failed to create new coin")
		}

		actualCoins = append(actualCoins, coin)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("failed to check errors",
			zap.String("method", "GetActualCoins"),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "error during to row iteration: %v", err)
	}

	return actualCoins, nil
}

func (s *Storage) GetAggregateCoins(ctx context.Context, titles []string, aggType string) ([]*entities.Coin, error) {
	query := `SELECT title, `

	switch aggType {
	case cases.AggTypeMax:
		query += `MAX(cost) AS max_cost`
	case cases.AggTypeMin:
		query += `MIN(cost) AS min_cost`
	case cases.AggTypeAvg:
		query += `AVG(cost) AS avg_cost`
	}

	query += ` FROM coins WHERE title = ANY($1) GROUP BY title;`

	rows, err := s.pool.Query(ctx, query, titles)
	if err != nil {
		s.logger.Error("failed to process query",
			zap.String("method", "GetAggregateCoins"),
			zap.String("query", query),
			zap.String("aggregation type", aggType),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "failed to execute query: %v", err)
	}
	defer rows.Close()

	aggregateCoins := make([]*entities.Coin, 0, len(titles))

	for rows.Next() {
		var title string
		var cost float64

		if err = rows.Scan(&title, &cost); err != nil {
			s.logger.Error("failed to scan coin params",
				zap.String("method", "GetAggregateCoins"),
				zap.Error(err))

			return nil, errors.Wrapf(entities.ErrInternal, "failed to scan row: %v", err)
		}

		coin, err := entities.NewCoin(title, cost, time.Now())
		if err != nil {
			s.logger.Error("failed to create new coin",
				zap.String("method", "GetAggregateCoins"),
				zap.String("title", title),
				zap.Float64("cost", cost),
				zap.Error(err))

			return nil, errors.Wrap(err, "failed to create new coin")
		}

		aggregateCoins = append(aggregateCoins, coin)
	}

	if err = rows.Err(); err != nil {
		s.logger.Error("failed to check errors",
			zap.String("method", "GetAggregateCoins"),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "error during to row iteration: %v", err)
	}

	return aggregateCoins, nil
}

func (s *Storage) Close() {
	if s.pool != nil {
		defer s.pool.Close()
	}
}
