package entities

import (
	"fmt"
	"time"
)

type Coin struct {
	Title    string
	Cost     float64
	ActualAt time.Time
}

func NewCoin(title string, cost float64, actualAt time.Time) (*Coin, error) {
	if title == "" {
		return nil, fmt.Errorf("title cannot be empty")
	} else if cost <= 0 {
		return nil, fmt.Errorf("cost must be positiv")
	} else if actualAt.IsZero() {
		return nil, fmt.Errorf("actualAt cannot be zero")
	}
	return &Coin{
		Title:    title,
		Cost:     cost,
		ActualAt: actualAt,
	}, nil
}
