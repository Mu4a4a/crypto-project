package http

import (
	"context"

	"crypto-project/internal/entities"
)

type Service interface {
	GetLastRates(ctx context.Context, title []string) ([]*entities.Coin, error)
	GetAggregateRates(ctx context.Context, title []string, aggFunc string) ([]*entities.Coin, error)
}
