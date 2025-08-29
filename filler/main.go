package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Arkine2054/l0/internal/models"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	broker := os.Getenv("KAFKA_BROKERS")
	topic := os.Getenv("KAFKA_TOPIC")

	producer, err := kafka.NewProducer(&kafka.ConfigMap{
		"bootstrap.servers": broker,
	})
	if err != nil {
		log.Fatal("cannot create producer:", err)
	}
	defer producer.Close()

	for i := 1; i <= 5; i++ {
		order := models.Order{
			OrderUID:          fmt.Sprintf("test-%d", time.Now().UnixNano()),
			TrackNumber:       fmt.Sprintf("track-%d", i),
			Entry:             "web",
			Locale:            "ru",
			InternalSignature: "",
			CustomerID:        fmt.Sprintf("customer-%d", i),
			DeliveryService:   "DHL",
			ShardKey:          "9",
			SmID:              i,
			DateCreated:       time.Now(),
			OofShard:          "1",
			Delivery: models.Delivery{
				Name:    "Иван Иванов",
				Phone:   "+79998887766",
				Zip:     "123456",
				City:    "Москва",
				Address: "ул. Пушкина, д. Колотушкина",
				Region:  "Московская область",
				Email:   "ivan@example.com",
			},
			Payment: models.Payment{
				Transaction:  fmt.Sprintf("tx-%d", i),
				RequestID:    "",
				Currency:     "RUB",
				Provider:     "visa",
				Amount:       1000 * i,
				PaymentDT:    time.Now().Unix(),
				Bank:         "Сбербанк",
				DeliveryCost: 250,
				GoodsTotal:   1000 * i,
				CustomFee:    0,
			},
			Items: []models.Item{
				{
					ChrtID:      100 + i,
					TrackNumber: fmt.Sprintf("track-%d", i),
					Price:       500 * i,
					RID:         "rid-1",
					Name:        "Кроссовки",
					Sale:        10,
					Size:        "42",
					TotalPrice:  450 * i,
					NmID:        555,
					Brand:       "Nike",
					Status:      202,
				},
				{
					ChrtID:      200 + i,
					TrackNumber: fmt.Sprintf("track-%d", i),
					Price:       700 * i,
					RID:         "rid-2",
					Name:        "Футболка",
					Sale:        5,
					Size:        "L",
					TotalPrice:  665 * i,
					NmID:        777,
					Brand:       "Adidas",
					Status:      202,
				},
			},
		}

		data, err := json.Marshal(order)
		if err != nil {
			log.Println("Error marshal test order:", err)
			continue
		}

		err = producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          data,
		}, nil)

		if err != nil {
			log.Println("produce error:", err)
		} else {
			log.Printf("Produced order %s\n", order.OrderUID)
		}

		time.Sleep(1 * time.Second)
	}

	producer.Flush(15 * 1000)
	fmt.Println("Filler finished.")
}
