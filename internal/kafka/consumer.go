package kafka

import (
	"context"
	"fmt"
	"github.com/go-playground/validator/v10"
	"log"
	"time"

	"github.com/Arkine2054/l0/internal/models"
	"github.com/Arkine2054/l0/internal/util"
	ckafka "github.com/confluentinc/confluent-kafka-go/kafka"
)

type Consumer struct {
	client      ConsumerClient
	topic       string
	group       string
	dlqProducer ProducerIF
}

var validate = validator.New()

type ConsumerIF interface {
	Listen(ctx context.Context, handler func(order *models.Order) error) error
	Close(ctx context.Context) error
}

type ConsumerClient interface {
	SubscribeTopics(topics []string) error
	Poll(timeoutMs int) interface{}
	CommitMessage(msg *ckafka.Message) ([]ckafka.TopicPartition, error)
	Close() error
}

type confluentConsumerClient struct{ c *ckafka.Consumer }

func (a *confluentConsumerClient) SubscribeTopics(topics []string) error {
	return a.c.SubscribeTopics(topics, nil)
}
func (a *confluentConsumerClient) Poll(timeoutMs int) interface{} { return a.c.Poll(timeoutMs) }
func (a *confluentConsumerClient) CommitMessage(msg *ckafka.Message) ([]ckafka.TopicPartition, error) {
	return a.c.CommitMessage(msg)
}

func (a *confluentConsumerClient) Close() error { return a.c.Close() }

func NewConsumer(brokers, topic, group string, dlqProducer *Producer) (*Consumer, error) {
	var (
		c   *ckafka.Consumer
		err error
	)

	for i := 0; i < 5; i++ {
		c, err = ckafka.NewConsumer(&ckafka.ConfigMap{
			"bootstrap.servers":  brokers,
			"group.id":           group,
			"auto.offset.reset":  "earliest",
			"enable.auto.commit": false,
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

	client := &confluentConsumerClient{c: c}
	err = client.SubscribeTopics([]string{topic})
	if err != nil {
		return nil, fmt.Errorf("error cannot subscribe: %w", err)
	}

	return &Consumer{
		client:      client,
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
			ev := c.client.Poll(100)
			if ev == nil {
				continue
			}

			switch e := ev.(type) {
			case *ckafka.Message:
				order, err := util.UnmarshalOrder(e.Value)
				if err != nil {
					log.Printf("[WARN] Ошибка парсинга JSON: %v. Отправляем в DLQ...", err)
					c.sendToDLQ(e.Value)
					_, _ = c.client.CommitMessage(e)
					continue
				}

				if err := validate.Struct(order); err != nil {
					log.Printf("[WARN] Ошибка валидации заказа: %v. Отправляем в DLQ...", err)
					c.sendToDLQ(e.Value)
					_, _ = c.client.CommitMessage(e)
					continue
				}

				if err := handler(order); err != nil {
					log.Printf("[ERROR] Ошибка обработки заказа %s: %v", order.OrderUID, err)
					c.sendToDLQ(e.Value)
					continue
				}

				if _, err = c.client.CommitMessage(e); err != nil {
					log.Printf("[ERROR] Commit offset error: %v", err)
				} else {
					log.Printf("[OK] Заказ %s обработан и offset закоммичен", order.OrderUID)
				}

			case ckafka.Error:
				log.Printf("[KAFKA ERROR] %v", e)
				if e.IsFatal() {
					return e
				}
			}
		}
	}
}

func (c *Consumer) sendToDLQ(value []byte) {
	if c.dlqProducer != nil {
		if err := c.dlqProducer.Send(nil, value); err != nil {
			log.Printf("[ERROR] Не удалось отправить в DLQ: %v", err)
		}
	}
}

func (c *Consumer) Close(ctx context.Context) error {
	done := make(chan struct{})
	go func() {
		defer close(done)
		if err := c.client.Close(); err != nil {
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
