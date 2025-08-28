package shutdown

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func GracefulShutdown(ctx context.Context, cancel context.CancelFunc, cleanupFns ...func(context.Context)) {
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigCh
		log.Println("Завершаем работу...")

		// отменяем контекст приложения
		cancel()

		// создаём общий контекст с таймаутом для graceful shutdown
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		// выполняем все cleanup-функции
		for _, fn := range cleanupFns {
			fn(shutdownCtx)
		}

		log.Println("Все ресурсы освобождены, выходим.")
		os.Exit(0)
	}()
}
