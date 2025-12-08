package config

import (
	"os"
	"strings"
)

type Config struct {
	KafkaBrokers []string
	AccountDB    string // DSN
	OrderDB      string // DSN
}

func Load() *Config {
	return &Config{
		KafkaBrokers: strings.Split(getEnv("KAFKA_BROKERS", "localhost:9092"), ","),
		AccountDB:    getEnv("ACCOUNT_DB_DSN", "postgres://hts:hts@localhost:5432/hts_account?sslmode=disable"),
		OrderDB:      getEnv("ORDER_DB_DSN", "postgres://hts:hts@localhost:5432/hts_order?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
