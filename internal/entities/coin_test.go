package entities

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestNewCoin(t *testing.T) {
	testTable := []struct {
		Name     string
		Title    string
		Cost     float64
		ActualAt time.Time
		WantErr  bool
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
		},
		{
			Name:     "Invalid cost",
			Title:    "Bitcoin",
			Cost:     0,
			ActualAt: time.Now(),
			WantErr:  true,
		},
		{
			Name:     "Zero ActualAt",
			Title:    "Bitcoin",
			Cost:     125.2,
			ActualAt: time.Time{},
			WantErr:  true,
		},
	}
	for _, testCase := range testTable {
		t.Run(testCase.Name, func(*testing.T) {
			_, err := NewCoin(testCase.Title, testCase.Cost, testCase.ActualAt)
			if testCase.WantErr {
				assert.Error(t, err, "Expected an error for test case:"+testCase.Name)
			} else {
				assert.NoError(t, err, "Did not expect an error for test case:"+testCase.Name)
			}
		})
	}
}
