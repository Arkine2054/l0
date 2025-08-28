package main

import (
	"log"
	"os"

	"github.com/Arkine2054/l0/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		log.Fatalf("Service failed: %v", err)
		os.Exit(1)
	}
}
