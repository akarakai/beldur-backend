//go:build integration

package app

import (
	"beldur/internal/account"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCreateAccount_Success(t *testing.T) {
	email := "spatagarru@gmail.com"
	req := account.CreateAccountRequest{
		Username: "username123",
		Password: "password123",
		Email:    &email,
	}
	client := &http.Client{
		Timeout: time.Second,
	}
	resp, body := DoJSONOK[account.CreateAccountResponse](t, client, "POST", "/auth/signup", req)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	assert.Equal(t, *body.Email, email)
	assert.Equal(t, body.AccountName, req.Username)
}
