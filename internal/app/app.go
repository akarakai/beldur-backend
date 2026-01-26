package app

import (
	"beldur/internal/account"
	"beldur/internal/campaign"
	"beldur/pkg/auth/jwt"
	"beldur/pkg/db/postgres"
	"beldur/pkg/db/tx"
	"beldur/pkg/middleware"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
)

type FiberApp struct {
	app *fiber.App
}

func New() *FiberApp {
	jwtService := buildJwtService()
	transactor, qProvider := buildTransactorQuerierProvider()
	accountHandler := account.NewHandlerFromDeps(account.Deps{
		Transactor: transactor,
		QProvider:  qProvider,
		Issuer:     jwtService,
	})

	campaignHandler := campaign.NewHandlerFromDeps(campaign.Deps{
		QProvider:  qProvider,
		Transactor: transactor,
	})

	app := fiber.New()
	app.Use(healthcheck.New())
	app.Use(fiberlogger.New())

	authMiddleware := middleware.Auth(jwtService)

	app.Post("/auth/signup", middleware.Validation[account.CreateAccountRequest](), accountHandler.Register)
	app.Post("/auth/login", middleware.Validation[account.UsernamePasswordLoginRequest](), accountHandler.Login)
	app.Post("/campaign", authMiddleware, middleware.Validation[campaign.CreationRequest](), campaignHandler.HandleCreateCampaign)
	app.Get("/campaign", campaignHandler.HandleGetCampaign)
	app.Patch("/account", authMiddleware, accountHandler.UpdateAccount)
	app.Post("/campaign/:campaignId", authMiddleware, middleware.Validation[campaign.JoinRequest](), campaignHandler.HandleJoinCampaign)

	return &FiberApp{
		app: app,
	}
}

func (app *FiberApp) Listen(port string) error {
	return app.app.Listen(fmt.Sprintf(":%s", port))
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
