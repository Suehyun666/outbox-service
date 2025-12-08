package domain

import (
	"fmt"
	"time"
)

// OutboxEntry : DB outbox_events 테이블과 매핑되는 구조체
type OutboxEntry struct {
	ID             int64
	AggregateType  string
	AggregateID    int64
	EventType      string // 이걸로 Kafka 토픽 결정
	Payload        []byte // BYTEA
	IdempotencyKey string
	Status         string
	CreatedAt      time.Time
	AvailableAt    time.Time
}

// GetTopic : event_type을 Kafka 토픽으로 매핑
func (e *OutboxEntry) GetTopic() string {
	// event_type → topic 매핑
	switch e.EventType {
	case "ACCOUNT_RESERVED", "ACCOUNT_FILLED", "ACCOUNT_RELEASED", "BALANCE_UPDATED":
		return "account-events"
	case "ORDER_PLACED", "ORDER_CREATED", "ORDER_CANCEL_REQUESTED":
		return "order.created"
	default:
		return "unknown-events"
	}
}

// GetKey : aggregate_id를 파티셔닝 키로 사용
func (e *OutboxEntry) GetKey() string {
	return fmt.Sprintf("%d", e.AggregateID)
}

// EventProducer : 메시지 발송 인터페이스 (Kafka가 될지 뭐일지 모름)
type EventProducer interface {
	Send(topic string, key string, payload []byte) error
	Close() error
}

// OutboxProcessor : DB 트랜잭션 내에서 처리 로직을 정의하는 인터페이스
type OutboxProcessor interface {
	// ProcessBatch : 트랜잭션을 열고 -> 읽고(Lock) -> 콜백(Send) -> 업데이트 -> 커밋
	ProcessBatch(batchSize int, tableName string, sendFunc func(entry OutboxEntry) error) (int, error)
}
