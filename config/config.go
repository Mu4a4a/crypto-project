package config

import (
	"github.com/spf13/viper"
	"os"
)

type PostgresConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type CryptoCompareConfig struct {
	ApiKey   string
	BaseURL  string
	Currency string
}

type Config struct {
	Postgres      PostgresConfig
	CryptoCompare CryptoCompareConfig
}

func NewConfig() *Config {
	return &Config{
		Postgres: PostgresConfig{
			Host:     viper.GetString("db.host"),
			Port:     viper.GetString("db.port"),
			User:     getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			DBName:   getEnv("DB_NAME", ""),
			SSLMode:  viper.GetString("db.sslMode"),
		},
		CryptoCompare: CryptoCompareConfig{
			ApiKey:   getEnv("CRYPTO_COMPARE_API_KEY", ""),
			BaseURL:  viper.GetString("provider.baseURL"),
			Currency: viper.GetString("provider.currency"),
		},
	}
}

func getEnv(key string, defaultValue string) string {
	if value, exist := os.LookupEnv(key); exist {
		return value
	}

	return defaultValue
}
