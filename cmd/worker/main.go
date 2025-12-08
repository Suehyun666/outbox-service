package main

import (
	"log"
	"sync"
	"time"

	"outbox-service/internal/config"
	"outbox-service/internal/infra/db"
	"outbox-service/internal/infra/kafka"
	"outbox-service/internal/service"
)

func main() {
	// 1. 설정 로드
	cfg := config.Load()

	// 2. Kafka Producer 초기화 (하나를 공유해서 씀)
	producer, err := kafka.NewSaramaProducer(cfg.KafkaBrokers)
	if err != nil {
		log.Fatal("Kafka init failed:", err)
	}
	defer producer.Close()

	// 3. DB 연결 (주문용, 계좌용 각각 생성)
	accountDB, err := db.NewPostgresProcessor(cfg.AccountDB)
	if err != nil {
		log.Fatal("Account DB init failed:", err)
	}

	orderDB, err := db.NewPostgresProcessor(cfg.OrderDB)
	if err != nil {
		log.Fatal("Order DB init failed:", err)
	}

	// 4. 워커 생성 (서비스 주입)
	// - Account Worker: "outbox_events" 테이블 폴링
	accountWorker := service.NewOutboxWorker("AccountWorker", accountDB, producer, "outbox_events")

	// - Order Worker: "outbox" 테이블 폴링 (order는 outbox 테이블명 다를 수 있음)
	orderWorker := service.NewOutboxWorker("OrderWorker", orderDB, producer, "outbox")

	// 5. 실행 (고루틴)
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		accountWorker.Start(100 * time.Millisecond) // 0.1초 주기
	}()

	go func() {
		defer wg.Done()
		orderWorker.Start(100 * time.Millisecond) // 0.1초 주기
	}()

	log.Println("🚀 Unified Outbox Worker is Running...")
	wg.Wait()
}
