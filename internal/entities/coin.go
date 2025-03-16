package entities

import (
	"time"

	"github.com/pkg/errors"
)

type Coin struct {
	Title    string
	Cost     float64
	ActualAt time.Time
}

func NewCoin(title string, cost float64, actualAt time.Time) (*Coin, error) {
	if title == "" {
		return nil, errors.Wrap(ErrInvalidParam, "title cannot be empty")
	}

	if cost <= 0 {
		return nil, errors.Wrap(ErrInvalidParam, "cost cannot be negative or zero")
	}

	if actualAt.IsZero() {
		return nil, errors.Wrap(ErrInvalidParam, "actualAt cannot be zero")
	}

	return &Coin{
		Title:    title,
		Cost:     cost,
		ActualAt: actualAt,
	}, nil
}
