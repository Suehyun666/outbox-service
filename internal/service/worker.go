package service

import (
	"log/slog"
	"outbox-service/internal/domain"
	"time"
)

type OutboxWorker struct {
	name      string
	processor domain.OutboxProcessor
	producer  domain.EventProducer
	tableName string
}

func NewOutboxWorker(name string, proc domain.OutboxProcessor, prod domain.EventProducer, table string) *OutboxWorker {
	return &OutboxWorker{
		name:      name,
		processor: proc,
		producer:  prod,
		tableName: table,
	}
}

// Start : 고루틴으로 무한 루프 실행
func (w *OutboxWorker) Start(interval time.Duration) {
	slog.Info("Worker Started", "worker", w.name, "table", w.tableName)

	ticker := time.NewTicker(interval)
	defer ticker.Stop() // 리소스 해제 습관

	for { // 무한 루프
		select {
		case <-ticker.C:
			// 티커가 울리면 실행 (기본)
			w.processLoop()
		}
	}
}

func (w *OutboxWorker) processLoop() {
	for {
		// 콜백 함수 정의 (그대로 유지)
		sendTask := func(entry domain.OutboxEntry) error {
			return w.producer.Send(entry.GetTopic(), entry.GetKey(), entry.Payload)
		}

		// 1. 배치 처리 실행
		count, err := w.processor.ProcessBatch(50, w.tableName, sendTask)
		if err != nil {
			slog.Error("Batch Processing Failed", "worker", w.name, "error", err)
			break // 에러 나면 이번 틱은 포기하고 다음 틱 기다림
		}

		if count > 0 {
			slog.Info("Events Processed", "worker", w.name, "count", count)
		}

		// 2. 핵심 로직: 가져온 게 50개보다 적으면(0~49) DB가 비었다는 뜻 -> 루프 탈출하고 쉼
		if count < 50 {
			break
		}

		// 3. 만약 50개를 꽉 채워 가져왔다면? (count == 50)
		// -> "밀린 게 더 있다"는 뜻이므로 break 없이 즉시 다시 ProcessBatch 실행 (Zero-Sleep)
	}
}
