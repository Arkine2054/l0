package util

import (
	"fmt"

	confluent "github.com/confluentinc/confluent-kafka-go/kafka"
)

func TestConnection(brokers string) error {
	testConsumer, err := confluent.NewConsumer(&confluent.ConfigMap{
		"bootstrap.servers": brokers,
		"group.id":          "test-connection",
		"auto.offset.reset": "earliest",
	})
	if err != nil {
		return fmt.Errorf("Kafka connection error: %w", err)
	}
	defer testConsumer.Close()
	fmt.Println("Connected to Kafka (test connection)")
	return nil
}
