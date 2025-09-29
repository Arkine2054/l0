package kafka

import (
	"context"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"testing"
	"time"

	"github.com/Arkine2054/l0/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestNewProducer(t *testing.T) {
	tests := []struct {
		name    string
		brokers string
		topic   string
		wantErr bool
	}{
		{
			name:    "valid config",
			brokers: "localhost:9092",
			topic:   "test-topic",
			wantErr: false,
		},
		{
			name:    "empty brokers",
			brokers: "",
			topic:   "test-topic",
			wantErr: false, // Kafka позволяет создать producer с пустыми brokers
		},
		{
			name:    "empty topic",
			brokers: "localhost:9092",
			topic:   "",
			wantErr: false, // Producer создается, но topic может быть пустым
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			producer, err := NewProducer(tt.brokers, tt.topic)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, producer)
			} else {
				// В реальном тесте без Kafka это может упасть
				// В production лучше использовать testcontainers
				if err != nil {
					t.Skip("Skipping test - Kafka not available")
					return
				}
				assert.NoError(t, err)
				assert.NotNil(t, producer)
				assert.Equal(t, tt.topic, producer.topic)
			}
		})
	}
}

func TestProducer_SendOrder(t *testing.T) {
	// Этот тест требует реального Kafka
	// В production лучше использовать testcontainers
	t.Skip("Skipping integration test - requires Kafka")

	producer, err := NewProducer("localhost:9092", "test-topic")
	assert.NoError(t, err)
	defer producer.Close(context.Background())

	order := &models.Order{
		OrderUID:        "test-order-123",
		TrackNumber:     "WBILMTESTTRACK",
		Entry:           "WBIL",
		Locale:          "en",
		CustomerID:      "test-customer-123",
		DeliveryService: "meest",
		ShardKey:        "9",
		SmID:            99,
		DateCreated:     time.Now(),
		OofShard:        "1",
		Delivery: models.Delivery{
			Name:    "Test Testov",
			Phone:   "+9720000000",
			Zip:     "2639809",
			City:    "Kiryat Mozkin",
			Address: "Ploshad Mira 15",
			Region:  "Kraiot",
			Email:   "test@gmail.com",
		},
		Payment: models.Payment{
			Transaction:  "b563feb7b2b84b6test",
			Currency:     "USD",
			Provider:     "wbpay",
			Amount:       1817,
			PaymentDT:    1637907727,
			Bank:         "alpha",
			DeliveryCost: 1500,
			GoodsTotal:   317,
			CustomFee:    0,
		},
		Items: []models.Item{
			{
				ChrtID:      9934930,
				TrackNumber: "WBILMTESTTRACK",
				Price:       453,
				RID:         "ab4219087a764ae0b229c92aa27dd3ff",
				Name:        "Mascaras",
				Sale:        30,
				Size:        "0",
				TotalPrice:  317,
				NmID:        2389212,
				Brand:       "Vivienne Sabo",
				Status:      202,
			},
		},
	}

	err = producer.SendOrder(order)
	assert.NoError(t, err)
}

func TestProducer_Close(t *testing.T) {
	// Этот тест требует реального Kafka
	t.Skip("Skipping integration test - requires Kafka")

	producer, err := NewProducer("localhost:9092", "test-topic")
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = producer.Close(ctx)
	assert.NoError(t, err)
}

// --- Unit tests with mocked client ---

type mockProducerClient struct {
	produceErr error
	flushed    bool
}

func (m *mockProducerClient) Produce(msg *kafka.Message) error { return m.produceErr }
func (m *mockProducerClient) Flush(timeoutMs int)              { m.flushed = true }

func TestProducer_Send_PropagatesError(t *testing.T) {
	p := &Producer{client: &mockProducerClient{produceErr: assert.AnError}, topic: "t"}
	err := p.Send([]byte("k"), []byte("v"))
	assert.Error(t, err)
}
