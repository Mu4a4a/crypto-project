package entities_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"crypto-project/internal/entities"
)

func TestNewCoin(t *testing.T) {
	t.Parallel()

	testTable := []struct {
		Name     string
		Title    string
		Cost     float64
		ActualAt time.Time
		WantErr  bool
		ResErr   error
	}{
		{
			Name:     "Valid data",
			Title:    "Bitcoin",
			Cost:     125.2,
			ActualAt: time.Now(),
			WantErr:  false,
		},
		{
			Name:     "Empty name",
			Title:    "",
			Cost:     125.2,
			ActualAt: time.Now(),
			WantErr:  true,
			ResErr:   entities.ErrInvalidParam,
		},
		{
			Name:     "Invalid cost",
			Title:    "Bitcoin",
			Cost:     0,
			ActualAt: time.Now(),
			WantErr:  true,
			ResErr:   entities.ErrInvalidParam,
		},
		{
			Name:     "Zero ActualAt",
			Title:    "Bitcoin",
			Cost:     125.2,
			ActualAt: time.Time{},
			WantErr:  true,
			ResErr:   entities.ErrInvalidParam,
		},
	}
	for _, tc := range testTable {
		t.Run(tc.Name, func(t *testing.T) {
			coin, err := entities.NewCoin(tc.Title, tc.Cost, tc.ActualAt)
			if tc.WantErr {
				require.ErrorIs(t, err, tc.ResErr)
				require.Nil(t, coin)
				return
			}
			require.NoError(t, err)
			require.Equal(t, &entities.Coin{
				Title:    tc.Title,
				Cost:     tc.Cost,
				ActualAt: tc.ActualAt,
			}, coin)
		})
	}
}
