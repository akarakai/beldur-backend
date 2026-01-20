package main

import (
	"beldur/internal/db/postgres"
	"beldur/internal/domain/account"
	"beldur/internal/domain/player"
	"beldur/internal/usecase"
	"context"
	"log/slog"

	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(".env.dev"); err != nil {
		panic(err.Error())
	}

	cfg, err := postgres.ConfigFromEnv()
	if err != nil {
		panic(err.Error())
	}

	pool, err := postgres.NewPgxPool(context.Background(), cfg)
	if err != nil {
		panic(err.Error())
	}

	defer pool.Close()

	t, qp := postgres.NewTransactor(pool)

	accountRepo := postgres.NewAccountRepository(qp)
	playerRepo := postgres.NewPlayerRepository(qp)

	accountSvc := account.NewService(accountRepo)
	playerSvc := player.NewService(playerRepo)

	useCase := usecase.NewRegisterAccount(t, accountSvc, playerSvc)

	req := usecase.CreateAccountRequest{
		Username: "user1234",
		Password: "pass1234",
		Email:    "spatagarru.laspezia2@gmail.com",
	}

	_, err = useCase.Execute(context.Background(), req)
	if err != nil {
		panic(err.Error())
	}
	slog.Info("All success")
}
