package usecase

import "time"

type CreateAccountRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Email    string `json:"email,omitempty"`
}

type CreateAccountResponse struct {
	AccountID   int                `json:"account_id"`
	AccountName string             `json:"accountName"`
	Email       string             `json:"email,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	User        SimpleUserResponse `json:"user"`
}

type SimpleUserResponse struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
}
