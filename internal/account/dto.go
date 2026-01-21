package account

import "time"

// #### ACCOUNT CREATION

type CreateAccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

type CreateAccountResponse struct {
	AccountID   int                  `json:"account_id"`
	AccountName string               `json:"username"`
	Email       string               `json:"email,omitempty"`
	CreatedAt   time.Time            `json:"created_at"`
	Player      PlayerCreateResponse `json:"player"`
}

type PlayerCreateResponse struct {
	PlayerID int    `json:"player_id"`
	Name     string `json:"name"`
}

// #### Login

type UsernamePasswordLoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
