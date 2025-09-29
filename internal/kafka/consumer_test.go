package kafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Arkine2054/l0/internal/models"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
)

func TestNewConsumer(t *testing.T) {
	tests := []struct {
		name        string
		brokers     string
		topic       string
		group       string
		dlqProducer *Producer
		wantErr     bool
	}{
		{
			name:        "valid config",
			brokers:     "localhost:9092",
			topic:       "test-topic",
			group:       "test-group",
			dlqProducer: nil,
			wantErr:     false,
		},
		{
			name:        "empty brokers",
			brokers:     "",
			topic:       "test-topic",
			group:       "test-group",
			dlqProducer: nil,
			wantErr:     false, // Kafka позволяет создать consumer с пустыми brokers
		},
		{
			name:        "with dlq producer",
			brokers:     "localhost:9092",
			topic:       "test-topic",
			group:       "test-group",
			dlqProducer: &Producer{}, // В реальном тесте это будет создан через NewProducer
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer, err := NewConsumer(tt.brokers, tt.topic, tt.group, tt.dlqProducer)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, consumer)
			} else {
				// В реальном тесте без Kafka это может упасть
				// В production лучше использовать testcontainers
				if err != nil {
					t.Skip("Skipping test - Kafka not available")
					return
				}
				assert.NoError(t, err)
				assert.NotNil(t, consumer)
				assert.Equal(t, tt.topic, consumer.topic)
				assert.Equal(t, tt.group, consumer.group)
			}
		})
	}
}

func TestConsumer_Listen(t *testing.T) {
	// Этот тест требует реального Kafka
	t.Skip("Skipping integration test - requires Kafka")

	consumer, err := NewConsumer("localhost:9092", "test-topic", "test-group", nil)
	assert.NoError(t, err)
	defer consumer.Close(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	handler := func(order *models.Order) error {
		// Тестовая обработка заказа
		return nil
	}

	err = consumer.Listen(ctx, handler)
	assert.NoError(t, err)
}

func TestConsumer_Close(t *testing.T) {
	// Этот тест требует реального Kafka
	t.Skip("Skipping integration test - requires Kafka")

	consumer, err := NewConsumer("localhost:9092", "test-topic", "test-group", nil)
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = consumer.Close(ctx)
	assert.NoError(t, err)
}

// Тест для функции sendToDLQ
func TestConsumer_sendToDLQ(t *testing.T) {
	tests := []struct {
		name        string
		dlqProducer *Producer
		value       []byte
	}{
		{
			name:        "with dlq producer",
			dlqProducer: &Producer{},
			value:       []byte("test message"),
		},
		{
			name:        "without dlq producer",
			dlqProducer: nil,
			value:       []byte("test message"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			consumer := &Consumer{
				dlqProducer: tt.dlqProducer,
			}

			// Функция sendToDLQ не возвращает ошибку, просто логирует
			// В реальном тесте можно проверить логи
			consumer.sendToDLQ(tt.value)
		})
	}
}

// --- Unit tests with mocked client ---

type mockConsumerClient struct {
	events    []interface{}
	commitErr error
	closed    bool
}

func (m *mockConsumerClient) SubscribeTopics(topics []string) error { return nil }
func (m *mockConsumerClient) Poll(timeoutMs int) interface{} {
	if len(m.events) == 0 {
		return nil
	}
	ev := m.events[0]
	m.events = m.events[1:]
	return ev
}
func (m *mockConsumerClient) CommitMessage(msg *ckafka.Message) (*ckafka.TopicPartition, error) {
	return nil, m.commitErr
}
func (m *mockConsumerClient) Close() error { m.closed = true; return nil }

type mockProducerIF struct{ sent [][]byte }

func (m *mockProducerIF) Send(key, value []byte) error { m.sent = append(m.sent, value); return nil }

func TestConsumer_Listen_ParseError_GoesToDLQ(t *testing.T) {
	msg := &ckafka.Message{Value: []byte("{ invalid json ")}
	mc := &mockConsumerClient{events: []interface{}{msg}}
	mp := &mockProducerIF{}
	c := &Consumer{client: mc, dlqProducer: mp}

	// handler should not be called; provide no-op
	_ = c.Listen(context.Background(), func(order *models.Order) error { return nil })

	// After processing, DLQ should have one entry
	assert.Equal(t, 1, len(mp.sent))
}

func TestConsumer_Listen_ValidationError_GoesToDLQ(t *testing.T) {
	// Minimal invalid order (missing required fields)
	bad := []byte(`{"order_uid": ""}`)
	msg := &ckafka.Message{Value: bad}
	mc := &mockConsumerClient{events: []interface{}{msg}}
	mp := &mockProducerIF{}
	c := &Consumer{client: mc, dlqProducer: mp}

	_ = c.Listen(context.Background(), func(order *models.Order) error { return nil })
	assert.Equal(t, 1, len(mp.sent))
}

func TestConsumer_Listen_HandlerError_GoesToDLQ_NoCommitBlock(t *testing.T) {
	// Valid minimal order
	good := []byte(`{"order_uid":"ok","track_number":"t","entry":"e","locale":"en","customer_id":"c","delivery_service":"d","shardkey":"s","sm_id":0,"date_created":"2021-01-01T00:00:00Z","oof_shard":"o","delivery":{"name":"n","phone":"p","zip":"z","city":"c","address":"a","region":"r","email":"e@e.com"},"payment":{"transaction":"t","currency":"USD","provider":"p","amount":1,"payment_dt":1,"bank":"b","delivery_cost":0,"goods_total":0,"custom_fee":0},"items":[{"chrt_id":1,"track_number":"t","price":1,"rid":"r","name":"n","sale":0,"size":"s","total_price":1,"nm_id":1,"brand":"b","status":1}]}`)
	msg := &ckafka.Message{Value: good}
	mc := &mockConsumerClient{events: []interface{}{msg}}
	mp := &mockProducerIF{}
	c := &Consumer{client: mc, dlqProducer: mp}

	_ = c.Listen(context.Background(), func(order *models.Order) error { return errors.New("boom") })

	// Error from handler should send to DLQ
	assert.Equal(t, 1, len(mp.sent))
}
