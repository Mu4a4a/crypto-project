package cryptocompare

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"crypto-project/internal/cases"
	"crypto-project/internal/entities"
)

const (
	fsyms  = "fsyms"
	tsyms  = "tsyms"
	apiKey = "api_key"
)

type Client struct {
	client   *http.Client
	logger   cases.Logger
	apiKey   string
	baseURL  string
	currency string
}

type Options func(*Client)

func NewClient(options ...Options) (*Client, error) {
	client := &Client{
		client:   &http.Client{},
		apiKey:   "",
		baseURL:  "",
		currency: "",
	}

	for _, opt := range options {
		opt(client)
	}

	if client.apiKey == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "api key cannot be empty")
	}

	if client.baseURL == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "base url cannot be empty")
	}

	if client.currency == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "currency cannot be empty")
	}

	if client.logger == nil || client.logger == cases.Logger(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "logger not set")
	}

	return client, nil
}

func WithApiKey(apiKey string) Options {
	return func(client *Client) {
		client.apiKey = apiKey
	}
}

func WithBaseURL(baseURL string) Options {
	return func(client *Client) {
		client.baseURL = baseURL
	}
}

func WithCurrency(currency string) Options {
	return func(client *Client) {
		client.currency = currency
	}
}

func WithLogger(logger cases.Logger) Options {
	return func(client *Client) {
		client.logger = logger
	}
}

func (c *Client) GetActualRates(ctx context.Context, titles []string) ([]*entities.Coin, error) {
	u, err := url.Parse(c.baseURL)
	if err != nil {
		c.logger.Error("failed to parse base url",
			zap.String("method", "GetActualRates"),
			zap.String("base url", c.baseURL),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "failed to parse url: %v", err)
	}

	params := url.Values{}
	params.Add(fsyms, strings.Join(titles, ","))
	params.Add(tsyms, c.currency)
	params.Add(apiKey, c.apiKey)

	u.RawQuery = params.Encode()
	fullURL := u.String()

	req, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
	if err != nil {
		c.logger.Error("failed to create new request",
			zap.String("method", "GetActualRates"),
			zap.String("request method", "GET"),
			zap.String("full url", fullURL),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "failed to create new request: %v", err)
	}

	response, err := c.client.Do(req)
	if err != nil {
		c.logger.Error("failed to do request",
			zap.String("method", "GetActualRates"),
			zap.Any("request", req),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "failed to do request: %v", err)
	}
	defer response.Body.Close()

	c.logger.Info("response body data",
		zap.Any("data", response.Body))

	if response.StatusCode != http.StatusOK {
		return nil, errors.Wrap(entities.ErrInternal, "status code must be ok")
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		c.logger.Error("failed to read body",
			zap.String("method", "GetActualRates"),
			zap.Any("body", response.Body),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "failed to read response body: %v", err)
	}

	if len(body) == 0 {
		c.logger.Warn("body is empty",
			zap.String("method", "GetActualRates"))

		return nil, errors.Wrap(entities.ErrInvalidParam, "response body cannot be empty")
	}

	var rowData map[string]map[string]float64

	err = json.Unmarshal(body, &rowData)
	if err != nil {
		c.logger.Error("failed to unmarshal body",
			zap.String("method", "GetActualRates"),
			zap.Any("body", body),
			zap.Error(err))

		return nil, errors.Wrapf(entities.ErrInternal, "failed to parse json: %v", err)
	}

	coins := make([]*entities.Coin, 0, len(titles))

	for title, data := range rowData {
		cost := data[c.currency]

		coin, err := entities.NewCoin(title, cost, time.Now())
		if err != nil {
			c.logger.Error("failed to create new coin",
				zap.String("method", "GetActualRates"),
				zap.String("title", title),
				zap.Float64("cost", cost),
				zap.Error(err))

			return nil, errors.Wrap(err, "failed to create new coin")
		}
		coins = append(coins, coin)
	}

	return coins, nil
}
