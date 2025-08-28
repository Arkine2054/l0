package kafka

import (
	"context"
	"fmt"
	"github.com/Arkine2054/l0/internal/models"
	"github.com/Arkine2054/l0/internal/util"
	"github.com/confluentinc/confluent-kafka-go/kafka"
	"log"
	"time"
)

// Consumer отвечает за чтение сообщений из Kafka
type Consumer struct {
	consumer    *kafka.Consumer
	topic       string
	group       string
	dlqProducer *Producer
}

func NewConsumer(brokers, topic, group string, dlqProducer *Producer) (*Consumer, error) {
	var (
		c   *kafka.Consumer
		err error
	)

	for i := 0; i < 5; i++ {
		c, err = kafka.NewConsumer(&kafka.ConfigMap{
			"bootstrap.servers":  brokers,
			"group.id":           group,
			"auto.offset.reset":  "earliest", // читаем из начала, если новый consumer group
			"enable.auto.commit": false,      // сами управляем коммитами
		})
		if err == nil {
			break
		}
		fmt.Printf("Ошибка подключения к Kafka (%v), повтор...\n", err)
		time.Sleep(time.Second * 2)
	}
	if err != nil {
		return nil, fmt.Errorf("error cannot create consumer: %w", err)
	}

	err = c.SubscribeTopics([]string{topic}, nil)
	if err != nil {
		return nil, fmt.Errorf("error cannot subscribe: %w", err)
	}

	return &Consumer{
		consumer:    c,
		topic:       topic,
		group:       group,
		dlqProducer: dlqProducer,
	}, nil
}

func (c *Consumer) Listen(ctx context.Context, handler func(order *models.Order) error) error {
	log.Println("Polling Kafka...")
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			ev := c.consumer.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *kafka.Message:
				order, err := util.UnmarshalOrder(e.Value)
				if err != nil {
					log.Printf("[WARN] Ошибка парсинга JSON: %v. Отправляем в DLQ...", err)

					if c.dlqProducer != nil {
						if dlqErr := c.dlqProducer.Send(nil, e.Value); dlqErr != nil {
							log.Printf("[ERROR] Не удалось отправить в DLQ: %v", dlqErr)
						}
					}

					// коммитим offset, чтобы не застревать на битом сообщении
					_, _ = c.consumer.CommitMessage(e)
					continue
				}

				if err := handler(order); err != nil {
					log.Printf("[ERROR] Ошибка обработки заказа %s: %v", order.OrderUID, err)
					continue
				}

				// успех → коммитим offset
				_, err = c.consumer.CommitMessage(e)
				if err != nil {
					log.Printf("[ERROR] Commit offset error: %v", err)
				} else {
					log.Printf("[OK] Заказ %s обработан и offset закоммичен", order.OrderUID)
				}

			case kafka.Error:
				log.Printf("[KAFKA ERROR] %v", e)
				if e.IsFatal() {
					return e
				}
			}
		}
	}
}

func (c *Consumer) Close(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := c.consumer.Close(); err != nil {
			log.Printf("Kafka consumer close error: %v", err)
		}
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
