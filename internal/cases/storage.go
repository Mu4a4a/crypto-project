package cases

import (
	"context"

	"crypto-project/internal/entities"
)

//go:generate mockgen -source=storage.go -destination=mocks/storage_mock.go -package=mocks
type Storage interface {
	Store(ctx context.Context, coins []*entities.Coin) error
	GetCoinsList(ctx context.Context) ([]string, error)
	GetActualCoin(ctx context.Context, titles []string) ([]*entities.Coin, error)
	GetAggregateCoins(ctx context.Context, titles []string, aggType string) ([]*entities.Coin, error)
}
