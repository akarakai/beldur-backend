package campaign

import "time"

type CreationRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreationResponse struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	MasterID    int       `json:"master_id"`
}
