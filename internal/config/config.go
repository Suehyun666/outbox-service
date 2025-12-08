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
		// 기본값은 도련님의 요청하신 IP와 포트 반영
		AccountDB: getEnv("ACCOUNT_DB_DSN", "postgres://user:pass@10.0.4.2:5432/account_db?sslmode=disable"),
		OrderDB:   getEnv("ORDER_DB_DSN", "postgres://user:pass@10.0.4.2:5434/order_db?sslmode=disable"),
	}
}

func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
