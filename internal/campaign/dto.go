package campaign

import "time"

type CreationRequest struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description" validate:"required"`
}

type CreationResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	MasterID    int       `json:"master_id"`
	AccessCode  string    `json:"access_code"`
}

type JoinRequest struct {
	Code string `json:"access_code" validate:"required"`
}

type JoinResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// For now no info request, I think im doing with request paramaters for filtering
// and now I will not filter

type SimpleCampaignInfoResponse struct {
	ID          int        `json:"id"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Status      string     `json:"status"`
	CreatedAt   time.Time  `json:"created_at"`
	StartedAt   *time.Time `json:"started_at"`
	// With master excluded
	NumberPlayers int  `json:"number_players"`
	CanBeJoined   bool `json:"can_be_joined"`
}
