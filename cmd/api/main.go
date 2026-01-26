package main

import (
	"beldur/internal/app"
	"beldur/pkg/logger"
	"os"

	"github.com/joho/godotenv"
)

func main() {
	logger.Init()
	defer logger.Sync()
	if err := godotenv.Load(".env.dev"); err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")

	fiber := app.New()
	if err := fiber.Listen(port); err != nil {
		panic(err)
	}
}
