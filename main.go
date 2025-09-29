package main

import (
	"log"

	"github.com/Arkine2054/l0/internal/app"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Printf("%v\n", err)
		log.Println("No .env file found, using system environment")
	}

	if err := app.Run(); err != nil {
		log.Fatal(err)
	}
}
