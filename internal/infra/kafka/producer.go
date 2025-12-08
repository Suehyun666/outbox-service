package kafka

import (
	"github.com/IBM/sarama"
)

type SaramaProducer struct {
	producer sarama.SyncProducer
}

func NewSaramaProducer(brokers []string) (*SaramaProducer, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	// 필요 시 파티셔너 설정 등 추가

	p, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}
	return &SaramaProducer{producer: p}, nil
}

func (sp *SaramaProducer) Send(topic string, key string, payload []byte) error {
	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key), // 순서 보장을 위해 Key 사용 (aggregate_id)
		Value: sarama.ByteEncoder(payload),
	}
	_, _, err := sp.producer.SendMessage(msg)
	return err
}

func (sp *SaramaProducer) Close() error {
	return sp.producer.Close()
}
