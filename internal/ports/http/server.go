package http

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"crypto-project/internal/cases"
	"crypto-project/internal/entities"
)

const (
	keyTitles  = "titles"
	keyAggFunc = "aggFunc"

	RouteLastRates      = "/coins/rates/:titles"
	RouteAggregateRates = "/coins/:aggFunc/:titles"
)

type Server struct {
	service Service
	srv     *fiber.App
	port    string
	logger  cases.Logger
}

func NewServer(service Service, port string, logger cases.Logger) (*Server, error) {
	if service == nil || service == Service(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "service not set")
	}

	if logger == nil || logger == cases.Logger(nil) {
		return nil, errors.Wrap(entities.ErrInvalidParam, "logger not set")
	}

	if port == "" {
		return nil, errors.Wrap(entities.ErrInvalidParam, "port cannot be empty")
	}

	app := fiber.New()

	return &Server{
		service: service,
		srv:     app,
		port:    port,
		logger:  logger,
	}, nil
}

func (s *Server) Run() error {

	api := s.srv.Group("/api")
	{
		v1 := api.Group("/v1")
		{
			v1.Post(RouteLastRates, s.lastRatesHandler)

			v1.Post(RouteAggregateRates, s.aggregateRatesHandler)
		}
	}

	s.logger.Info("Starting the server",
		zap.String("port", s.port))

	if err := s.srv.Listen(s.port); err != nil {
		s.logger.Error("failed to start http server",
			zap.String("port", s.port),
			zap.Error(err))

		return errors.Wrapf(entities.ErrInternal, "failed to start http server: %v", err)
	}

	return nil
}

func (s *Server) lastRatesHandler(c *fiber.Ctx) error {
	titles := strings.Split(c.Params(keyTitles), ",")

	s.logger.Info("input request",
		zap.String("endpoint", "lastRatesHandler"),
		zap.Strings("titles", titles))

	if len(titles) == 0 {
		s.logger.Warn("empty titles param",
			zap.String("endpoint", "lastRatesHandler"))

		return s.processErr(c, errors.Wrap(entities.ErrInvalidParam, "titles cannot be empty"))
	}

	coins, err := s.service.GetLastRates(c.Context(), titles)
	if err != nil {
		s.logger.Error("failed to get last rates",
			zap.String("endpoint", "lastRatesHandler"),
			zap.Strings("titles", titles),
			zap.Error(err))

		return s.processErr(c, err)
	}

	response := &Response{
		Coins: make([]*ResponseCoin, 0, len(coins)),
	}

	for _, coin := range coins {
		dto := &ResponseCoin{
			Title: coin.Title,
			Cost:  coin.Cost,
		}

		response.Coins = append(response.Coins, dto)
	}

	s.logger.Info("successfully processed request",
		zap.String("endpoint", "lastRatesHandler"),
		zap.Int("coins_count", len(coins)))

	return c.Status(fiber.StatusOK).JSON(response)
}

func (s *Server) aggregateRatesHandler(c *fiber.Ctx) error {
	aggFunc := c.Params(keyAggFunc)

	if aggFunc != cases.AggTypeMax && aggFunc != cases.AggTypeMin && aggFunc != cases.AggTypeAvg {
		return s.processErr(c, errors.Wrapf(entities.ErrInvalidParam, "wrong agg func: %s", aggFunc))
	}

	titles := strings.Split(c.Params(keyTitles), ",")

	if len(titles) == 0 {
		s.logger.Warn("empty titles param",
			zap.String("endpoint", "aggregateRatesHandler"))

		return s.processErr(c, errors.Wrap(entities.ErrInvalidParam, "titles cannot be empty"))
	}

	s.logger.Info("input request",
		zap.String("endpoint", "aggregateRatesHandler"),
		zap.String("aggregate function", aggFunc),
		zap.Strings("titles", titles))

	coins, err := s.service.GetAggregateRates(c.Context(), titles, aggFunc)
	if err != nil {
		s.logger.Error("failed to get aggregate rates",
			zap.String("endpoint", "aggregateRatesHandler"),
			zap.Strings("titles", titles),
			zap.String("aggregate func", aggFunc),
			zap.Error(err))

		return s.processErr(c, err)
	}

	response := &Response{
		Coins: make([]*ResponseCoin, 0, len(coins)),
	}

	for _, coin := range coins {
		dto := &ResponseCoin{
			AggregationType: aggFunc,
			Title:           coin.Title,
			Cost:            coin.Cost,
		}

		response.Coins = append(response.Coins, dto)
	}

	s.logger.Info("successfully processed request",
		zap.String("endpoint", "aggregateRatesHandler"),
		zap.Int("coins_count", len(coins)))

	return c.Status(fiber.StatusOK).JSON(response)
}

func (s *Server) processErr(c *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entities.ErrInternal):
		return c.Status(fiber.StatusInternalServerError).JSON(DtoErrResponse{
			Code:    fiber.StatusInternalServerError,
			Message: err.Error(),
		})
	case errors.Is(err, entities.ErrInvalidParam):
		return c.Status(fiber.StatusBadRequest).JSON(DtoErrResponse{
			Code:    fiber.StatusBadRequest,
			Message: err.Error(),
		})
	}

	return c.Status(fiber.StatusInternalServerError).JSON(DtoErrResponse{
		Code:    fiber.StatusInternalServerError,
		Message: err.Error(),
	})
}
