package cases

import (
	"context"

	"project/internal/entities"
)

//go:generate mockgen -source=crypto_provider.go -destination=./testdata/crypto_provider.go -package=testdata

type CryptoProvider interface {
	GetActualRates(ctx context.Context, titles []string) ([]entities.Coin, error)
}
