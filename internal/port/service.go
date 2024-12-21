package port

import (
	"context"

	"project/internal/entities"
)

type Service interface {
	GetLastRates(ctx context.Context, titles []string) ([]entities.Coin, error)
	GetMaxRates(ctx context.Context, titles []string) ([]entities.Coin, error)
	GetMinRates(ctx context.Context, titles []string) ([]entities.Coin, error)
	GetAvgRates(ctx context.Context, titles []string) ([]entities.Coin, error)
	GetPercentRates(ctx context.Context, titles []string) ([]entities.Coin, error)
}
