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
	for range ticker.C {
		// 콜백 함수: DB에서 읽은 데이터를 Kafka로 쏨
		sendTask := func(entry domain.OutboxEntry) error {
			topic := entry.GetTopic()
			key := entry.GetKey()
			return w.producer.Send(topic, key, entry.Payload)
		}

		count, err := w.processor.ProcessBatch(50, w.tableName, sendTask)
		if err != nil {
			slog.Error("Batch Processing Failed", "worker", w.name, "error", err)
			continue
		}

		if count > 0 {
			slog.Info("Events Processed", "worker", w.name, "count", count)
		}
	}
}
