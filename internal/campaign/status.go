package campaign

type StatusCampaign string

const (
	// StatusCreated campaign is created but still not started
	StatusCreated   StatusCampaign = "created"
	StatusStarted   StatusCampaign = "started"
	StatusFinished  StatusCampaign = "finished"
	StatusCancelled StatusCampaign = "cancelled"
)
