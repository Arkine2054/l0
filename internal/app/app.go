package app

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Arkine2054/l0/internal/controller"
	"github.com/Arkine2054/l0/internal/kafka"
	"github.com/Arkine2054/l0/internal/logic"
	"github.com/Arkine2054/l0/internal/models"
	"github.com/Arkine2054/l0/internal/repo"
	"github.com/Arkine2054/l0/internal/shutdown"
	"github.com/Arkine2054/l0/internal/util"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
	"time"
)

func Run() error {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbName := os.Getenv("DB_NAME")
	kafkaBrokers := os.Getenv("KAFKA_BROKERS")
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	group := "l0-consumer-group"

	if kafkaBrokers == "" || kafkaTopic == "" {
		return fmt.Errorf("KAFKA_BROKERS и KAFKA_TOPIC обязательны")
	}

	// проверка Kafka соединения
	if err := util.TestConnection(kafkaBrokers); err != nil {
		return err
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		dbHost, dbPort, dbUser, dbPassword, dbName,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("DB connect error: %w", err)
	}

	// Repo + Cache
	repository := repo.NewRepo(db)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := repository.WarmUpCache(ctx); err != nil {
		log.Printf("cache warmup error: %v", err)
	} else {
		log.Println("cache warmed up from DB")
	}

	l := logic.NewLogic(repository)

	// Kafka producer для DLQ
	dlqProducer, err := kafka.NewProducer(kafkaBrokers, "orders_dlq")
	if err != nil {
		log.Fatalf("DLQ producer error: %v", err)
	}

	consumer, err := kafka.NewConsumer(kafkaBrokers, kafkaTopic, group, dlqProducer)
	if err != nil {
		log.Fatalf("consumer error: %v", err)
	}

	// Consumer слушает Kafka в фоне
	go func() {
		if err := consumer.Listen(context.Background(), func(order *models.Order) error {
			if err := l.CreateOrder(context.Background(), order); err != nil {
				return err
			}
			log.Printf("order %s saved and cached", order.OrderUID)
			return nil
		}); err != nil {
			log.Printf("consumer stopped: %v", err)
		}
	}()

	c := controller.NewController(l, dlqProducer)
	router := gin.Default()
	router.LoadHTMLGlob("templates/*")

	router.GET("/", c.Index)
	router.GET("/order/:id", c.GetOrder)

	// HTTP сервер с graceful shutdown
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	shutdown.GracefulShutdown(ctx, cancel,
		func(ctx context.Context) { _ = repository.Close(ctx) },
		func(ctx context.Context) { _ = consumer.Close(ctx) },
		func(ctx context.Context) { _ = dlqProducer.Close(ctx) },
		func(ctx context.Context) {
			if err := srv.Shutdown(ctx); err != nil {
				log.Printf("HTTP server shutdown error: %v", err)
			}
		},
	)

	log.Println("HTTP сервер запущен на :8080")

	// блокируем Run, пока сервер работает
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
