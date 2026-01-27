package character

import (
	"beldur/internal/campaign"
	"beldur/pkg/db/postgres"
)

type Deps struct {
	QProvider postgres.QuerierProvider
}

func NewHandlerFromDeps(deps Deps) *HttpHandler {
	charRepo := NewPostgresRepository(deps.QProvider)

	// TODO interesting, I think its better to ask te repository to be dependency
	// and not the query provider like here. In particular here we have in the wiring
	// imported the package campaign
	// if We put the repository interface as dependency then its better
	// but then I have to change also other handlers deps (easy)
	campaignRepo := campaign.NewPostgresRepository(deps.QProvider)
	creationUseCase := NewCreateUseCase(campaignRepo, charRepo)
	return NewHttpHandler(creationUseCase)
}
