package application

import (
	"crypto-project/internal/adapters/provider/cryptocompare"
	"crypto-project/internal/entities"
	"crypto-project/internal/ports/http"
	"github.com/joho/godotenv"
	"github.com/pkg/errors"
	"go.uber.org/zap/zapcore"
	"os"

	"github.com/spf13/viper"
	"go.uber.org/zap"

	"crypto-project/internal/adapters/storage/postgres"
	"crypto-project/internal/cases"
)

type App struct {
	logger   cases.Logger
	provider cases.CryptoProvider
	storage  cases.Storage
	service  *cases.Service
	server   *http.Server
}

func New() (*App, error) {
	app := &App{}

	if err := app.initConfig(); err != nil {
		return nil, errors.Wrap(err, "failed to init config")
	}

	if err := app.initLogger(); err != nil {
		return nil, errors.Wrap(err, "failed to init logger")
	}

	if err := app.initStorage(); err != nil {
		return nil, errors.Wrap(err, "failed to init storage")
	}

	if err := app.initProvider(); err != nil {
		return nil, errors.Wrap(err, "failed to init provider")
	}

	if err := app.initService(); err != nil {
		return nil, errors.Wrap(err, "failed to init service")
	}

	if err := app.initServer(); err != nil {
		return nil, errors.Wrap(err, "failed to init server")
	}

	return app, nil
}

func (a *App) Run() error {
	a.logger.Info("Starting application")

	if err := a.server.Run(); err != nil {
		return errors.Wrap(err, "failed to starting server")
	}

	return nil
}

func (a *App) initConfig() error {
	if err := godotenv.Load(); err != nil {
		return errors.Wrapf(entities.ErrInternal, "failed to load env: %v", err)
	}

	viper.AddConfigPath("/app")
	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err != nil {
		return errors.Wrapf(entities.ErrInternal, "failed to read in config: %v", err)
	}

	return nil
}

func (a *App) initLogger() error {
	logLevel := viper.GetString("logger.level")

	if len(logLevel) == 0 {
		return errors.Wrap(entities.ErrInvalidParam, "log level not set")
	}

	var level zapcore.Level
	if err := level.UnmarshalText([]byte(logLevel)); err != nil {
		return errors.Wrapf(entities.ErrInternal, "failed to unmarshal log level: %v", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = zap.NewAtomicLevelAt(level)

	logger, err := cfg.Build()
	if err != nil {
		return errors.Wrapf(entities.ErrInternal, "failed to build logger cfg: %v", err)
	}

	a.logger = logger
	return nil
}

func (a *App) initStorage() error {
	storage, err := postgres.NewStorage(
		a.logger,
		postgres.WithHost(viper.GetString("db.host")),
		postgres.WithPort(viper.GetString("db.port")),
		postgres.WithDatabase(os.Getenv("DB_NAME")),
		postgres.WithUser(os.Getenv("DB_USER")),
		postgres.WithPassword(os.Getenv("DB_PASSWORD")),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create new storage")
	}

	a.storage = storage
	return nil
}

func (a *App) initProvider() error {
	provider, err := cryptocompare.NewClient(
		cryptocompare.WithLogger(a.logger),
		cryptocompare.WithApiKey(os.Getenv("CRYPTO_COMPARE_API_KEY")),
		cryptocompare.WithBaseURL(viper.GetString("provider.base_url")),
		cryptocompare.WithCurrency(viper.GetString("provider.currency")),
	)
	if err != nil {
		return errors.Wrap(err, "failed to create new provider")
	}

	a.provider = provider
	return nil
}

func (a *App) initService() error {
	service, err := cases.NewService(a.provider, a.storage, a.logger)
	if err != nil {
		return errors.Wrap(err, "failed to create new service")
	}

	a.service = service
	return nil
}

func (a *App) initServer() error {
	port := viper.GetString("http_server.port")

	server, err := http.NewServer(a.service, port, a.logger)
	if err != nil {
		return errors.Wrap(err, "failed to create new server")
	}

	a.server = server
	return nil
}

//TODO вынести в апликэйшин
//m, err := migrate.New(
//	"file://migrations",
//	"postgres://postgres:postgres@db:5432/postgres?sslmode=disable")
//if err != nil {
//	log.Fatal(err)
//}
//
//if err = m.Up(); err != nil {
//	log.Fatal(err)
//}
