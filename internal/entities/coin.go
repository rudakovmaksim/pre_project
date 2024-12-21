package entities

import (
	"github.com/pkg/errors"
)

type Coin struct {
	Title            string
	Cost             float64
	RatePercentDelta float32
}

func NewCoin(title string, cost float64) (*Coin, error) {
	if title == "" {
		return nil, errors.Wrap(ErrInvalidParam, "coin name is empty")
	}

	if cost <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "negative coin rate")
	}

	return &Coin{
		Title:            title,
		Cost:             cost,
		RatePercentDelta: 0,
	}, nil
}
