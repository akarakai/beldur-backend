package main

import (
	"beldur/internal/app"
	"beldur/pkg/auth/jwt"
	"beldur/pkg/db/postgres"
	"beldur/pkg/db/tx"
	"beldur/pkg/logger"
	"context"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	logger.Init()
	defer logger.Sync()
	if err := godotenv.Load(".env.dev"); err != nil {
		panic(err)
	}

	port := os.Getenv("PORT")

	jwtService := buildJwtService()
	transactor, querier := buildTransactorQuerierProvider()
	deps := app.Deps{
		JwtService: jwtService,
		Transactor: transactor,
		QProvider:  querier,
	}

	fiber := app.NewDev(deps)
	if err := fiber.Listen(port); err != nil {
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
