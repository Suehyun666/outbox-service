package kafka

import (
	"log/slog"
	"time"

	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaProducer(brokers []string) (*SaramaProducer, error) {
	config := sarama.NewConfig()

	// [중요] HTS/금융 시스템을 위한 안정성 설정
	// 1. 모든 ISR(In-Sync Replicas)이 데이터를 받을 때까지 대기 (데이터 유실 방지)
	config.Producer.RequiredAcks = sarama.WaitForAll

	// 2. 전송 실패 시 재시도 횟수 설정
	config.Producer.Retry.Max = 5
	config.Producer.Retry.Backoff = 100 * time.Millisecond

	// 3. 멱등성 프로듀서 활성화 (중복 전송 방지 & 순서 보장 핵심)
	// 네트워크 오류로 인한 재전송 시 중복 생기는 것을 Kafka 레벨에서 막아줍니다.
	config.Producer.Idempotent = true
	config.Net.MaxOpenRequests = 1 // 순서 엄격 보장을 위해 1로 설정 (Idempotent 켤 때 필수)

	// 4. SyncProducer는 Success 리턴이 필수
	config.Producer.Return.Successes = true

	// 5. 타임아웃 (너무 오래 매달려 있지 않도록)
	config.Producer.Timeout = 5 * time.Second

	p, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	slog.Info("Kafka Producer Connected", "brokers", brokers)
	return &SaramaProducer{producer: p}, nil
}

// Send : 메시지 전송 (동기 방식)
func (sp *SaramaProducer) Send(topic string, key string, payload []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key), // 파티셔닝 키 (순서 보장용)
		Value: sarama.ByteEncoder(payload),
		// 타임스탬프도 찍어주면 디버깅에 좋습니다
		Timestamp: time.Now(),
	}

	// 파티션, 오프셋 정보는 지금은 필요 없어서 버림 (_, _)
	_, _, err := sp.producer.SendMessage(msg)
	return err
}

func (sp *SaramaProducer) Close() error {
	slog.Info("Closing Kafka Producer...")
	return sp.producer.Close()
}
