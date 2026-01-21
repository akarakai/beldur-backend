package jwt

import (
	"beldur/internal/auth"
	"context"
	"errors"
	"fmt"
	"time"

	jwtlib "github.com/golang-jwt/jwt/v5"
)

var (
	ErrIssueToken   = errors.New("failed to issue token")
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token expired")
)

type Service struct {
	secret     []byte
	expiration time.Duration
	issuer     string
}

func NewService(secret []byte, expiration time.Duration, issuer string) *Service {
	return &Service{
		secret:     secret,
		expiration: expiration,
		issuer:     issuer,
	}
}

func (s *Service) Issue(ctx context.Context, c auth.Claims) (string, error) {
	now := time.Now()

	claims := jwtlib.MapClaims{
		"iss": s.issuer,
		"sub": c.Subject,
		"aid": c.AccountID,
		"iat": jwtlib.NewNumericDate(now),
		"exp": jwtlib.NewNumericDate(now.Add(s.expiration)),
	}

	tok := jwtlib.NewWithClaims(jwtlib.SigningMethodHS256, claims)
	signed, err := tok.SignedString(s.secret)
	if err != nil {
		return "", errors.Join(ErrIssueToken, err)
	}
	return signed, nil
}

func (s *Service) Verify(ctx context.Context, tokenStr string) (auth.Verified, error) {
	claims := jwtlib.MapClaims{}

	token, err := jwtlib.ParseWithClaims(
		tokenStr,
		claims,
		func(t *jwtlib.Token) (any, error) {
			if t.Method != jwtlib.SigningMethodHS256 {
				return nil, fmt.Errorf("%w: unexpected signing method %v", ErrInvalidToken, t.Header["alg"])
			}
			return s.secret, nil
		},
		jwtlib.WithIssuer(s.issuer),
		jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}),
	)
	if err != nil {
		if errors.Is(err, jwtlib.ErrTokenExpired) {
			return auth.Verified{}, ErrExpiredToken
		}
		return auth.Verified{}, errors.Join(ErrInvalidToken, err)
	}
	if !token.Valid {
		return auth.Verified{}, ErrInvalidToken
	}

	sub, err := claims.GetSubject()
	if err != nil || sub == "" {
		return auth.Verified{}, ErrInvalidToken
	}

	// MapClaims stores JSON numbers as float64
	var accountID int
	if v, ok := claims["aid"]; ok {
		switch n := v.(type) {
		case float64:
			accountID = int(n)
		case int:
			accountID = n
		case int64:
			accountID = int(n)
		}
	}

	return auth.Verified{
		Subject:   sub,
		AccountID: accountID,
	}, nil
}
