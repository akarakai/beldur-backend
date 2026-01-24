package account

import (
	"beldur/internal/auth"
	"beldur/internal/player"
	"beldur/pkg/db/postgres"
	"beldur/pkg/db/tx"
)

type Deps struct {
	Transactor tx.Transactor
	QProvider  postgres.QuerierProvider
	Issuer     auth.TokenIssuer
}

func NewHandlerFromDeps(deps Deps) *HttpHandler {
	accountRepo := NewPostgresRepository(deps.QProvider)
	playerRepo := player.NewPostgresRepository(deps.QProvider)
	uniquePlayerSvc := player.NewUniquePlayerService(playerRepo)

	registerUC := NewAccountRegistration(deps.Transactor, accountRepo, uniquePlayerSvc, deps.Issuer)
	loginUC := NewUsernamePasswordLogin(accountRepo, playerRepo, deps.Issuer)
	manageUC := NewAccountManagement(accountRepo)

	return NewHttpHandler(registerUC, loginUC, manageUC)
}
