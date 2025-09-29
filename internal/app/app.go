package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Arkine2054/l0/internal/controller"
	"github.com/Arkine2054/l0/internal/kafka"
	"github.com/Arkine2054/l0/internal/logic"
	"github.com/Arkine2054/l0/internal/models"
	"github.com/Arkine2054/l0/internal/repo"
	"github.com/Arkine2054/l0/internal/shutdown"
	"github.com/Arkine2054/l0/internal/util"
	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

func getEnv(key string, required bool, def string) string {
	val := os.Getenv(key)
	if val == "" {
		if required {
			log.Fatalf("missing required environment variable: %s", key)
		}
		return def
	}
	return val
}

func Run() error {
	dbHost := getEnv("DB_HOST", true, "")
	dbPort := getEnv("DB_PORT", true, "")
	dbUser := getEnv("DB_USER", true, "")
	dbPassword := getEnv("DB_PASSWORD", true, "")
	dbName := getEnv("DB_NAME", true, "")
	dbSSL := getEnv("DB_SSLMODE", false, "disable")

	kafkaBrokers := getEnv("KAFKA_BROKERS", true, "")
	kafkaTopic := getEnv("KAFKA_TOPIC", true, "")
	kafkaGroup := getEnv("KAFKA_GROUP", false, "l0-consumer-group")
	dlqTopic := getEnv("KAFKA_DLQ_TOPIC", false, "orders_dlq")

	httpPort := getEnv("HTTP_PORT", false, "8080")

	if err := util.TestConnection(kafkaBrokers); err != nil {
		return err
	}

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSL,
	)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return fmt.Errorf("DB connect error: %w", err)
	}

	cacheSizeStr := os.Getenv("CACHE_SIZE")
	cacheSize := 1000
	if cacheSizeStr != "" {
		if v, err := strconv.Atoi(cacheSizeStr); err == nil && v > 0 {
			cacheSize = v
		}
	}

	repository := repo.NewRepo(db, cacheSize)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := repository.WarmUpCache(ctx); err != nil {
		log.Printf("cache warmup error: %v", err)
	} else {
		log.Println("cache warmed up from DB")
	}

	l := logic.NewLogic(repository)

	dlqProducer, err := kafka.NewProducer(kafkaBrokers, dlqTopic)
	if err != nil {
		log.Fatalf("DLQ producer error: %v", err)
	}

	consumer, err := kafka.NewConsumer(kafkaBrokers, kafkaTopic, kafkaGroup, dlqProducer)
	if err != nil {
		log.Fatalf("consumer error: %v", err)
	}

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

	srv := &http.Server{
		Addr:    ":" + httpPort,
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

	log.Printf("HTTP сервер запущен на :%s", httpPort)

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server error: %w", err)
	}

	return nil
}
