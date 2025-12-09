package main

import (
	"log"
	"log/slog"
	"sync"
	"time"

	"outbox-service/internal/config"
	"outbox-service/internal/infra/db"
	"outbox-service/internal/infra/kafka"
	"outbox-service/internal/infra/logger"
	"outbox-service/internal/service"
)

func main() {
	// 1. 설정 로드
	cfg := config.Load()

	// 2. 로그 레벨 설정 (이제 호출 가능)
	logger.SetLogger(cfg.LogLevel)

	// 3. Kafka Producer 초기화
	producer, err := kafka.NewSaramaProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatal("Kafka init failed: ", err)
	}
	defer producer.Close()

	// 4. DB 연결
	accountDB, err := db.NewPostgresProcessor(cfg.AccountDB)
	if err != nil {
		log.Fatal("Account DB init failed: ", err)
	}

	orderDB, err := db.NewPostgresProcessor(cfg.OrderDB)
	if err != nil {
		log.Fatal("Order DB init failed: ", err)
	}

	// 5. 워커 생성
	// - Account Worker: "outbox_events" 테이블 폴링
	accountWorker := service.NewOutboxWorker("AccountWorker", accountDB, producer, "outbox_events")

	// - Order Worker: "outbox" 테이블 폴링
	orderWorker := service.NewOutboxWorker("OrderWorker", orderDB, producer, "outbox")

	// 6. 실행 (고루틴)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		accountWorker.Start(100 * time.Millisecond)
	}()

	go func() {
		defer wg.Done()
		orderWorker.Start(100 * time.Millisecond)
	}()

	slog.Info("🚀 Unified Outbox Worker is Running...", "log_level", cfg.LogLevel)
	wg.Wait()
}
