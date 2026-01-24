package campaign

import (
	"beldur/pkg/db/postgres"
	"beldur/pkg/db/tx"
)

type Deps struct {
	QProvider  postgres.QuerierProvider
	Transactor tx.Transactor
}

func NewHandlerFromDeps(deps Deps) *HttpHandler {
	campaignRepo := NewPostgresRepository(deps.QProvider)
	newCampaignUC := NewUseCase(campaignRepo, campaignRepo, campaignRepo, deps.Transactor)
	return NewHttpHandler(newCampaignUC)
}
