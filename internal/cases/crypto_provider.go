package cases

import (
	"context"

	"crypto-project/internal/entities"
)

//go:generate mockgen -source=crypto_provider.go -destination=mocks/crypto_provider_mock.go -package=mocks
type CryptoProvider interface {
	GetActualRates(ctx context.Context, titles []string) ([]*entities.Coin, error)
}
