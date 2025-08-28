package kafka

import (
	"context"
	"encoding/json"
	"github.com/Arkine2054/l0/internal/models"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	producer *kafka.Producer
	topic    string
}

func NewProducer(brokers, topic string) (*Producer, error) {
	p, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return nil, err
	}
	return &Producer{producer: p, topic: topic}, nil
}

func (p *Producer) Send(key, value []byte) error {
	return p.producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &p.topic, Partition: kafka.PartitionAny},
		Key:            key,
		Value:          value,
	}, nil)
}

func (p *Producer) SendOrder(order *models.Order) error {
	data, err := json.Marshal(order)
	if err != nil {
		return err
	}
	return p.Send([]byte(order.OrderUID), data)
}

func (p *Producer) Close(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		p.producer.Flush(5000)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
