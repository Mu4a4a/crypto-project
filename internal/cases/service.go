package cases

import (
	"context"

	"github.com/pkg/errors"

	"crypto-project/internal/entities"
)

type Service struct {
	Provider CryptoProvider
	Storage  Storage
}

func NewService(provider CryptoProvider, storage Storage) (*Service, error) {
	if provider == nil || provider == CryptoProvider(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "crypto provider not set")
	}

	if storage == nil || storage == Storage(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "storage not set")
	}

	return &Service{
		Provider: provider,
		Storage:  storage,
	}, nil
}

const (
	AggTypeMax = "max"
	AggTypeMin = "min"
	AggTypeAvg = "avg"
)

//TODO: COMMENTS IN CODE

func (s *Service) GetLastRates(ctx context.Context, titles []string) ([]*entities.Coin, error) {
	if len(titles) == 0 {
		return nil, errors.Wrap(entities.ErrInvalidParam, "titles cannot be empty")
	}

	if err := s.processNotExistingTitles(ctx, titles); err != nil {
		return nil, errors.Wrap(err, "failed to process not existing titles")
	}

	actualCoins, err := s.Storage.GetActualCoin(ctx, titles)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get actual coin")
	}

	return actualCoins, nil
}

func (s *Service) GetMaxRates(ctx context.Context, titles []string) ([]*entities.Coin, error) {
	if len(titles) == 0 {
		return nil, errors.Wrap(entities.ErrInvalidParam, "titles cannot be empty")
	}

	if err := s.processNotExistingTitles(ctx, titles); err != nil {
		return nil, errors.Wrap(err, "failed to process not existing titles")
	}

	aggregateCoins, err := s.Storage.GetAggregateCoins(ctx, titles, AggTypeMax)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aggregate coins")
	}

	return aggregateCoins, nil
}

func (s *Service) GetMinRates(ctx context.Context, titles []string) ([]*entities.Coin, error) {
	if len(titles) == 0 {
		return nil, errors.Wrap(entities.ErrInvalidParam, "titles cannot be empty")
	}

	if err := s.processNotExistingTitles(ctx, titles); err != nil {
		return nil, errors.Wrap(err, "failed to process not existing titles")
	}

	aggregateCoins, err := s.Storage.GetAggregateCoins(ctx, titles, AggTypeMin)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aggregate coins")
	}

	return aggregateCoins, nil
}

func (s *Service) GetAvgRates(ctx context.Context, titles []string) ([]*entities.Coin, error) {
	if len(titles) == 0 {
		return nil, errors.Wrap(entities.ErrInvalidParam, "titles cannot be empty")
	}

	if err := s.processNotExistingTitles(ctx, titles); err != nil {
		return nil, errors.Wrap(err, "failed to process not existing titles")
	}

	aggregateCoins, err := s.Storage.GetAggregateCoins(ctx, titles, AggTypeAvg)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get aggregate coins")
	}

	return aggregateCoins, nil
}

func (s *Service) ActualizeRates(ctx context.Context) error {
	listCoins, err := s.Storage.GetCoinsList(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get coins list")
	}

	actualRatesCoins, err := s.Provider.GetActualRates(ctx, listCoins)
	if err != nil {
		return errors.Wrap(err, "failed to get actual rates")
	}

	if err = s.Storage.Store(ctx, actualRatesCoins); err != nil {
		return errors.Wrap(err, "failed to store coins")
	}

	return nil
}

func (s *Service) processNotExistingTitles(ctx context.Context, titles []string) error {
	storedCoins, err := s.Storage.GetCoinsList(ctx)
	if err != nil {
		return errors.Wrap(err, "failed to get coins list")
	}

	allExistingTitles := make(map[string]struct{}, len(storedCoins))

	for _, title := range storedCoins {
		allExistingTitles[title] = struct{}{}
	}

	notStoredCoins := make([]string, 0)

	for _, title := range titles {
		if _, ok := allExistingTitles[title]; !ok {
			notStoredCoins = append(notStoredCoins, title)
		}
	}

	if len(notStoredCoins) == 0 {
		return nil
	}

	coins, err := s.Provider.GetActualRates(ctx, notStoredCoins)
	if err != nil {
		return errors.Wrap(err, "failed to get actual rates")
	}

	if err = s.Storage.Store(ctx, coins); err != nil {
		return errors.Wrap(err, "failed to store coins")
	}

	return nil
}
