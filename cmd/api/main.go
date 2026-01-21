package main

import (
	"beldur/internal/account"
	"beldur/internal/auth"
	"beldur/internal/auth/jwt"
	"beldur/pkg/db/postgres"
	"beldur/pkg/db/tx"
	"context"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env.dev"); err != nil {
		panic(err)
	}

	jwtService := buildJwtService()
	transactor, qProvider := buildTransactorQuerierProvider()

	accountHandler := account.NewHandlerFromDeps(account.Deps{
		Transactor: transactor,
		QProvider:  qProvider,
		Issuer:     jwtService,
	})

	app := fiber.New()
	app.Use(healthcheck.New())
	app.Use(logger.New())

	app.Post("/auth/signup", accountHandler.Register)
	app.Post("/auth/login", accountHandler.Login)

	app.Get("/auth", auth.Middleware(jwtService), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/noauth", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	if err := app.Listen(":3000"); err != nil {
		panic(err)
	}
}

func buildJwtService() *jwt.Service {
	secret := []byte(os.Getenv("JWT_SECRET"))
	expiration, _ := time.ParseDuration(os.Getenv("JWT_EXPIRATION"))
	issuer := os.Getenv("JWT_ISSUER")
	return jwt.NewService(secret, expiration, issuer)
}

func buildTransactorQuerierProvider() (tx.Transactor, postgres.QuerierProvider) {
	cfg, err := postgres.ConfigFromEnv()
	if err != nil {
		panic(err)
	}

	pgxPool, err := postgres.NewPgxPool(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	return postgres.NewTransactor(pgxPool)
}
