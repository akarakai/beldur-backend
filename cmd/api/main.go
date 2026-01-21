package main

import (
	"beldur/internal/account"
	"beldur/internal/auth"
	"beldur/internal/auth/jwt"
	"beldur/internal/player"
	"beldur/pkg/db/postgres"
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
	app := fiber.New()
	app.Use(healthcheck.New())
	app.Use(logger.New())

	accountHandler := buildAccountHandler()

	app.Post("/auth/signup", accountHandler.Register)
	app.Post("/auth/login", accountHandler.Login)

	app.Get("/auth", auth.Middleware(jwtService()), func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/noauth", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	if err := app.Listen(":3000"); err != nil {
		panic(err)
	}
}

func buildAccountHandler() *account.HttpHandler {
	cfg, err := postgres.ConfigFromEnv()
	if err != nil {
		panic(err)
	}

	pgxPool, err := postgres.NewPgxPool(context.Background(), cfg)
	if err != nil {
		panic(err)
	}

	transactor, querier := postgres.NewTransactor(pgxPool)

	accountRepo := account.NewAccountRepository(querier)
	playerRepo := player.NewPlayerRepository(querier)

	jwtService := jwtService()

	registerUC := account.NewAccountRegistration(
		transactor,
		accountRepo,
		player.NewUniquePlayerService(playerRepo),
		jwtService,
	)

	loginUC := account.NewUsernamePasswordLogin(
		accountRepo,
		jwtService,
	)

	return account.NewHttpHandler(registerUC, loginUC)

}

func jwtService() *jwt.Service {
	secret := []byte(os.Getenv("JWT_SECRET"))
	expiration, _ := time.ParseDuration(os.Getenv("JWT_EXPIRATION"))
	issuer := os.Getenv("JWT_ISSUER")
	return jwt.NewService(secret, expiration, issuer)
}
