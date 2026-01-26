package app

import (
	"beldur/internal/account"
	"beldur/internal/campaign"
	"beldur/pkg/auth/jwt"
	"beldur/pkg/db/postgres"
	"beldur/pkg/db/tx"
	"beldur/pkg/middleware"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/healthcheck"
	fiberlogger "github.com/gofiber/fiber/v2/middleware/logger"
)

type Deps struct {
	JwtService *jwt.Service
	Transactor tx.Transactor
	QProvider  postgres.QuerierProvider
}

type FiberApp struct {
	app *fiber.App
}

func NewDev(deps Deps) *FiberApp {
	return build(deps, false)
}

func NewTest(deps Deps) *FiberApp {
	return build(deps, true)
}

func build(deps Deps, test bool) *FiberApp {
	cfg := fiber.Config{}
	if test {
		cfg.DisableStartupMessage = true
	}

	app := fiber.New(cfg)

	app.Use(healthcheck.New())

	if !test {
		app.Use(fiberlogger.New())
	}

	// handlers
	accountHandler := account.NewHandlerFromDeps(account.Deps{
		Transactor: deps.Transactor,
		QProvider:  deps.QProvider,
		Issuer:     deps.JwtService,
	})
	campaignHandler := campaign.NewHandlerFromDeps(campaign.Deps{
		QProvider:  deps.QProvider,
		Transactor: deps.Transactor,
	})

	authMiddleware := middleware.Auth(deps.JwtService)

	// routes
	app.Post("/auth/signup", middleware.Validation[account.CreateAccountRequest](), accountHandler.Register)
	app.Post("/auth/login", middleware.Validation[account.UsernamePasswordLoginRequest](), accountHandler.Login)
	app.Post("/campaign", authMiddleware, middleware.Validation[campaign.CreationRequest](), campaignHandler.HandleCreateCampaign)
	app.Get("/campaign", campaignHandler.HandleGetCampaign)
	app.Patch("/account", authMiddleware, accountHandler.UpdateAccount)
	app.Post("/campaign/:campaignId", authMiddleware, middleware.Validation[campaign.JoinRequest](), campaignHandler.HandleJoinCampaign)

	return &FiberApp{app: app}
}

func (app *FiberApp) Listen(port string) error {
	return app.app.Listen(fmt.Sprintf(":%s", port))
}
