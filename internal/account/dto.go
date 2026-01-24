package account

import "time"

// #### ACCOUNT CREATION

type CreateAccountRequest struct {
	Username string  `json:"username"`
	Password string  `json:"password"`
	Email    *string `json:"email,omitempty"`
}

type CreateAccountResponse struct {
	AccountID   int                  `json:"account_id"`
	AccountName string               `json:"username"`
	Email       *string              `json:"email"`
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

// UpdateAccountRequest 's fields will be nullable when more than one.
// For now only email can be updated
type UpdateAccountRequest struct {
	Email string `json:"email"`
}

type UpdateAccountResponse struct {
	Email string `json:"email"`
}
