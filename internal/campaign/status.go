package campaign

type StatusCampaign string

const (
	// StatusCreated campaign is created but still not started
	StatusCreated   StatusCampaign = "CREATED"
	StatusStarted   StatusCampaign = "STARTED"
	StatusFinished  StatusCampaign = "FINISHED"
	StatusCancelled StatusCampaign = "CANCELLED"
)
