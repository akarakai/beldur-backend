package auth

import (
	"beldur/internal/id"
	"context"
)

type Claims struct {
	Subject  id.AccountId
	PlayerID id.PlayerId
}

type Principal struct {
	AccountID id.AccountId
	PlayerID  id.PlayerId
}

type Verified struct {
	Subject  id.AccountId
	PlayerId id.PlayerId
}

type TokenIssuer interface {
	Issue(ctx context.Context, claims Claims) (string, error)
}

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (Verified, error)
}
