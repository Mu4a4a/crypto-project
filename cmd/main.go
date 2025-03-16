package main

import (
	"context"
	"crypto-project/config"
	"crypto-project/internal/cases"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/spf13/viper"
	"log"
	"os"

	"crypto-project/internal/adapters/provider/cryptocompare"
	"crypto-project/internal/adapters/storage"
	"crypto-project/internal/adapters/storage/postgres"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func init() {
	if err := godotenv.Load(); err != nil {
		os.Exit(1)
	}

	viper.AddConfigPath("./config")
	viper.SetConfigName("config")

	if err := viper.ReadInConfig(); err != nil {
		os.Exit(1)
	}
}

func main() {
	ctx := context.TODO()

	cfg := config.NewConfig()

	pool, err := postgres.ConnectDB(ctx, cfg)
	if err != nil {
		log.Fatal(err, "failed to connect db")
	}

	storag, err := storage.NewStorage(pool)
	if err != nil {
		log.Fatal(err, "failed to create new storage")
	}

	provider, err := cryptocompare.NewCryptoCompareClient(cfg)
	if err != nil {
		log.Fatal(err, "failed to create new ccc")
	}

	service, err := cases.NewService(provider, storag)
	if err != nil {
		log.Fatal(err, "failed to create new service")
	}

	value, err := service.GetLastRates(ctx, []string{"BTC", "ETH"})
	if err != nil {
		log.Print(err)
	}

	value, err = service.GetLastRates(ctx, []string{"BTC", "ETH"})
	if err != nil {
		log.Print(err)
	}

	for _, coin := range value {
		fmt.Println(coin)
	}

	/*m, err := migrate.New(
		"file://migrations",
		"postgres://postgres:qwerty@172.19.0.1:5432/postgres?sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	if err = m.Up(); err != nil {
		log.Fatal(err)
	}*/
}
