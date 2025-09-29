package kafka

import (
	"context"
	"encoding/json"

	"github.com/Arkine2054/l0/internal/models"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type Producer struct {
	client ProducerClient
	topic  string
}

// ProducerIF is a minimal interface used by other components (e.g., DLQ sender).
type ProducerIF interface {
	Send(key, value []byte) error
}

// ProducerClient abstracts the low-level Kafka producer we use.
type ProducerClient interface {
	Produce(msg *ckafka.Message) error
	Flush(timeoutMs int)
}

// confluentProducerClient adapts confluent-kafka-go Producer to ProducerClient.
type confluentProducerClient struct {
	p *ckafka.Producer
}

func (c *confluentProducerClient) Produce(msg *ckafka.Message) error { return c.p.Produce(msg, nil) }
func (c *confluentProducerClient) Flush(timeoutMs int)               { c.p.Flush(timeoutMs) }

func NewProducer(brokers, topic string) (*Producer, error) {
	p, err := ckafka.NewProducer(&ckafka.ConfigMap{"bootstrap.servers": brokers})
	if err != nil {
		return nil, err
	}
	return &Producer{client: &confluentProducerClient{p: p}, topic: topic}, nil
}

func (p *Producer) Send(key, value []byte) error {
	return p.client.Produce(&ckafka.Message{
		TopicPartition: ckafka.TopicPartition{Topic: &p.topic, Partition: ckafka.PartitionAny},
		Key:            key,
		Value:          value,
	})
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
		p.client.Flush(5000)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
