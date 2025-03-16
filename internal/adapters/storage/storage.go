package storage

import (
	"context"
	"crypto-project/internal/cases"
	"crypto-project/internal/entities"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/pkg/errors"
	"time"
)

type Storage struct {
	pool *pgxpool.Pool
}

func NewStorage(pool *pgxpool.Pool) (*Storage, error) {
	if pool == nil {
		return nil, errors.Wrap(entities.ErrInvalidParam, "pool cannot be nil")
	}

	return &Storage{
		pool: pool,
	}, nil
}

func (s *Storage) Store(ctx context.Context, coins []*entities.Coin) error {
	query := `INSERT INTO coins (title, cost, actual_at) values ($1, $2, $3);`

	for _, coin := range coins {
		_, err := s.pool.Exec(ctx, query, coin.Title, coin.Cost, coin.ActualAt)
		if err != nil {
			return errors.Wrap(err, "failed to execute query")
		}
	}

	return nil
}

func (s *Storage) GetCoinsList(ctx context.Context) ([]string, error) {
	query := `SELECT title FROM coins;`

	rows, err := s.pool.Query(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()

	var titles []string

	for rows.Next() {
		var title string

		if err = rows.Scan(&title); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		titles = append(titles, title)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error during to row iteration")
	}

	return titles, nil
}

func (s *Storage) GetActualCoin(ctx context.Context, titles []string) ([]*entities.Coin, error) {
	query := `SELECT title, cost, actual_at FROM coins WHERE title = ANY($1);`

	rows, err := s.pool.Query(ctx, query, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()

	var actualCoins []*entities.Coin

	for rows.Next() {
		var title string
		var cost float64
		var actualAt time.Time

		if err = rows.Scan(&title, &cost, &actualAt); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		coin, err := entities.NewCoin(title, cost, actualAt)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new coin")
		}

		actualCoins = append(actualCoins, coin)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error during to row iteration")
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

	query += ` FROM coins WHERE title = $1 GROUP BY title;`

	rows, err := s.pool.Query(ctx, query, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed to execute query")
	}
	defer rows.Close()

	var aggregateCoins []*entities.Coin

	for rows.Next() {
		var title string
		var cost float64
		var actualAt time.Time

		if err = rows.Scan(&title, &cost, &actualAt); err != nil {
			return nil, errors.Wrap(err, "failed to scan row")
		}

		coin, err := entities.NewCoin(title, cost, actualAt)
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new coin")
		}

		aggregateCoins = append(aggregateCoins, coin)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "error during to row iteration")
	}

	return aggregateCoins, nil
}
