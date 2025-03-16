package cryptocompare

import (
	"context"
	"crypto-project/config"
	"crypto-project/internal/entities"
	"encoding/json"
	"github.com/pkg/errors"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type CryptoCompareClient struct {
	cfg    *config.Config
	Client *http.Client
}

func NewCryptoCompareClient(cfg *config.Config) (*CryptoCompareClient, error) {
	if cfg == nil {
		return nil, errors.Wrap(entities.ErrInvalidParam, "config cannot be nil")
	}

	return &CryptoCompareClient{
		cfg: cfg,
		Client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (c *CryptoCompareClient) GetActualRates(ctx context.Context, titles []string) ([]*entities.Coin, error) {
	u, err := url.Parse(c.cfg.CryptoCompare.BaseURL)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse url")
	}

	params := url.Values{}
	params.Add("fsyms", strings.Join(titles, ","))
	params.Add("tsyms", c.cfg.CryptoCompare.Currency)
	params.Add("api_key", c.cfg.CryptoCompare.ApiKey)

	u.RawQuery = params.Encode()
	fullURL := u.String()

	req, err := http.NewRequest("GET", fullURL, nil)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create new request")
	}

	response, err := c.Client.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, "failed to do request")
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.Wrap(err, "status code cannot be not ok")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Wrap(err, "failed to read response body")
	}

	if len(body) == 0 {
		return nil, errors.Wrap(err, "body cannot be empty")
	}

	var rowData map[string]map[string]float64

	err = json.Unmarshal(body, &rowData)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse json")
	}

	var coins []*entities.Coin

	for title, data := range rowData {
		cost := data[c.cfg.CryptoCompare.Currency]

		coin, err := entities.NewCoin(title, cost, time.Now())
		if err != nil {
			return nil, errors.Wrap(err, "failed to create new coin")
		}
		coins = append(coins, coin)
	}

	return coins, nil
}
