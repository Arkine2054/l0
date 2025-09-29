package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/Arkine2054/l0/internal/models"
	"github.com/brianvoe/gofakeit/v7"
	"github.com/confluentinc/confluent-kafka-go/kafka"
)

func main() {
	broker := os.Getenv("KAFKA_BROKERS")
	topic := os.Getenv("KAFKA_TOPIC")

	producer, err := kafka.NewProducer(&kafka.ConfigMap{"bootstrap.servers": broker})
	if err != nil {
		log.Fatal("cannot create producer:", err)
	}
	defer producer.Close()

	for i := 1; i <= 5; i++ {
		order := models.Order{
			OrderUID:        fmt.Sprintf("test-%d", time.Now().UnixNano()),
			TrackNumber:     gofakeit.UUID(),
			Entry:           "web",
			Locale:          gofakeit.Language(),
			CustomerID:      gofakeit.Username(),
			DeliveryService: "DHL",
			ShardKey:        fmt.Sprintf("%d", gofakeit.Number(1, 10)),
			SmID:            i,
			DateCreated:     time.Now(),
			OofShard:        "1",
			Delivery: models.Delivery{
				Name:    gofakeit.Name(),
				Phone:   gofakeit.Phone(),
				Zip:     gofakeit.Zip(),
				City:    gofakeit.City(),
				Address: gofakeit.Street(),
				Region:  gofakeit.State(),
				Email:   gofakeit.Email(),
			},
			Payment: models.Payment{
				Transaction:  gofakeit.UUID(),
				Currency:     "RUB",
				Provider:     "visa",
				Amount:       int(gofakeit.Price(500, 5000)),
				PaymentDT:    time.Now().Unix(),
				Bank:         gofakeit.Company(),
				DeliveryCost: 250,
				GoodsTotal:   gofakeit.Number(1000, 5000),
				CustomFee:    0,
			},
			Items: []models.Item{
				{
					ChrtID:      gofakeit.Number(100, 200),
					TrackNumber: gofakeit.UUID(),
					Price:       gofakeit.Number(100, 2000),
					RID:         gofakeit.UUID(),
					Name:        gofakeit.ProductName(),
					Sale:        gofakeit.Number(0, 50),
					Size:        gofakeit.RandomString([]string{"S", "M", "L", "XL"}),
					TotalPrice:  gofakeit.Number(100, 2000),
					NmID:        gofakeit.Number(1000, 2000),
					Brand:       gofakeit.ProductName(),
					Status:      202,
				},
			},
		}

		data, _ := json.Marshal(order)
		err := producer.Produce(&kafka.Message{
			TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
			Value:          data,
		}, nil)
		if err != nil {
			return
		}

		log.Printf("Produced fake order %s\n", order.OrderUID)
	}

	producer.Flush(15 * 1000)
	fmt.Println("Filler finished.")
}
