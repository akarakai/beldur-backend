package auth

import "context"

type Claims struct {
	Subject   string
	AccountID int
}

type Principal struct {
	AccountName string
	AccountID   int
}

type Verified struct {
	Subject   string
	AccountID int
}

type TokenIssuer interface {
	Issue(ctx context.Context, claims Claims) (string, error)
}

type TokenVerifier interface {
	Verify(ctx context.Context, token string) (Verified, error)
}
