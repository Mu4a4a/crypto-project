package cases

import (
	"context"

	"github.com/pkg/errors"
	"go.uber.org/zap"

	"crypto-project/internal/entities"
)

const (
	AggTypeMax = "max"
	AggTypeMin = "min"
	AggTypeAvg = "avg"
)

type Service struct {
	provider CryptoProvider
	storage  Storage
	logger   Logger
}

func NewService(provider CryptoProvider, storage Storage, logger Logger) (*Service, error) {
	if provider == nil || provider == CryptoProvider(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "crypto provider not set")
	}

	if storage == nil || storage == Storage(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "storage not set")
	}

	if logger == nil || logger == Logger(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "logger not set")
	}

	return &Service{
		provider: provider,
		storage:  storage,
		logger:   logger,
	}, nil
}

func (s *Service) GetLastRates(ctx context.Context, titles []string) ([]*entities.Coin, error) {
	if len(titles) == 0 {
		s.logger.Warn("empty titles param",
			zap.String("method", "get last rates"))

		return nil, errors.Wrap(entities.ErrInvalidParam, "titles cannot be empty")
	}

	if err := s.ActualizeRates(ctx); err != nil {
		s.logger.Error("failed to actualize rates",
			zap.String("method", "get last rates"),
			zap.Error(err))

		return nil, errors.Wrap(err, "failed to actualize rates")
	}

	if err := s.processNotExistingTitles(ctx, titles); err != nil {
		s.logger.Error("failed to process not existing titles",
			zap.String("method", "get last rates"),
			zap.Strings("titles", titles),
			zap.Error(err))

		return nil, errors.Wrap(err, "failed to process not existing titles")
	}

	actualCoins, err := s.storage.GetActualCoin(ctx, titles)
	if err != nil {
		s.logger.Error("failed to get actual coin",
			zap.String("method", "get last rates"),
			zap.Strings("titles", titles),
			zap.Error(err))

		return nil, errors.Wrap(err, "failed to get actual coin")
	}

	return actualCoins, nil
}

func (s *Service) GetAggregateRates(ctx context.Context, titles []string, aggFunc string) ([]*entities.Coin, error) {
	if len(titles) == 0 {
		s.logger.Warn("empty titles param",
			zap.String("method", "get aggregate rates"))

		return nil, errors.Wrap(entities.ErrInvalidParam, "titles cannot be empty")
	}

	if err := s.ActualizeRates(ctx); err != nil {
		s.logger.Error("failed to actualize rates",
			zap.String("method", "get aggregate rates"),
			zap.Error(err))

		return nil, errors.Wrap(err, "failed to actualize rates")
	}

	if err := s.processNotExistingTitles(ctx, titles); err != nil {
		s.logger.Error("failed to process not existing titles",
			zap.String("method", "get aggregate rates"),
			zap.Strings("titles", titles),
			zap.Error(err))

		return nil, errors.Wrap(err, "failed to process not existing titles")
	}

	aggregateCoins, err := s.storage.GetAggregateCoins(ctx, titles, aggFunc)
	if err != nil {
		s.logger.Error("failed to get aggregate coins",
			zap.String("method", "get aggregate rates"),
			zap.Strings("titles", titles),
			zap.Error(err))

		return nil, errors.Wrap(err, "failed to get aggregate coins")
	}

	return aggregateCoins, nil
}

func (s *Service) ActualizeRates(ctx context.Context) error {
	listCoins, err := s.storage.GetCoinsList(ctx)
	if err != nil {
		s.logger.Error("failed to get coins list",
			zap.String("method", "get actualize rates"),
			zap.Error(err))

		return errors.Wrap(err, "failed to get coins list")
	}

	if len(listCoins) == 0 {
		s.logger.Warn("coins not found",
			zap.String("method", "actualize rates"))

		return nil
	}

	actualRatesCoins, err := s.provider.GetActualRates(ctx, listCoins)
	if err != nil {
		s.logger.Error("failed to get actual rates",
			zap.String("method", "get actualize rates"),
			zap.Strings("list coins", listCoins),
			zap.Error(err))

		return errors.Wrap(err, "failed to get actual rates")
	}

	if err = s.storage.Store(ctx, actualRatesCoins); err != nil {
		s.logger.Error("failed to store coins",
			zap.String("method", "get actualize rates"),
			zap.Any("actual rates coins", actualRatesCoins),
			zap.Error(err))

		return errors.Wrap(err, "failed to store coins")
	}

	return nil
}

func (s *Service) processNotExistingTitles(ctx context.Context, titles []string) error {
	storedCoins, err := s.storage.GetCoinsList(ctx)
	if err != nil {
		s.logger.Error("failed to get coins list",
			zap.String("method", "processNotExistingTitles"),
			zap.Error(err))

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
		s.logger.Warn("all coins stored",
			zap.String("method", "processNotExistingTitles"))

		return nil
	}

	actualCoins, err := s.provider.GetActualRates(ctx, notStoredCoins)
	if err != nil {
		s.logger.Error("failed to get actual rates",
			zap.String("method", "processNotExistingTitles"),
			zap.Any("not stored coins", notStoredCoins),
			zap.Error(err))

		return errors.Wrap(err, "failed to get actual rates")
	}

	if err = s.storage.Store(ctx, actualCoins); err != nil {
		s.logger.Error("failed to store",
			zap.String("method", "processNotExistingTitles"),
			zap.Any("actual coins", actualCoins),
			zap.Error(err))

		return errors.Wrap(err, "failed to store coins")
	}

	return nil
}
