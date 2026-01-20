package postgres

import (
	"context"
	"errors"
	"time"

	"beldur/internal/domain/account"
	emailpkg "beldur/internal/domain/account/email"

	"github.com/jackc/pgx/v5"
)

type AccountRepository struct {
	q QuerierProvider
}

func NewAccountRepository(q QuerierProvider) *AccountRepository {
	return &AccountRepository{q: q}
}

func (a *AccountRepository) Save(ctx context.Context, acc *account.Account) (*account.Account, error) {
	query := `
		INSERT INTO accounts (username, password, email)
		VALUES ($1, $2, $3)
		RETURNING account_id, username, password, email, created_at
	`

	// Map domain Email (value object) -> SQL NULL
	var email any
	if acc.Email.IsNull() {
		email = nil
	} else {
		email = acc.Email.String()
	}

	row := a.q(ctx).QueryRow(ctx, query, acc.Username, acc.Password, email)

	saved, err := a.scanAccount(row)
	if err != nil {
		return nil, err
	}
	// INSERT ... RETURNING should always return a row
	if saved == nil {
		return nil, errors.New("insert account returned no row")
	}
	return saved, nil
}

func (a *AccountRepository) FindByUsername(ctx context.Context, username string) (*account.Account, error) {
	query := `
		SELECT account_id, username, password, email, created_at
		FROM accounts
		WHERE username = $1
		LIMIT 1
	`

	row := a.q(ctx).QueryRow(ctx, query, username)
	return a.scanAccount(row)
}

func (a *AccountRepository) FindById(ctx context.Context, accountId int) (*account.Account, error) {
	query := `
		SELECT account_id, username, password, email, created_at
		FROM accounts
		WHERE account_id = $1
		LIMIT 1
	`

	row := a.q(ctx).QueryRow(ctx, query, accountId)
	return a.scanAccount(row)
}

// scanAccount translates DB row -> domain model.
// Returns (nil, nil) when no row is found.
func (a *AccountRepository) scanAccount(row pgx.Row) (*account.Account, error) {
	var (
		id        int
		username  string
		password  string
		em        *string
		createdAt time.Time
	)

	err := row.Scan(&id, &username, &password, &em, &createdAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	var accEmail emailpkg.Email
	if em != nil {
		// Already validated when stored; ignore error defensively
		accEmail, _ = emailpkg.New(*em)
	}

	return &account.Account{
		Id:        id,
		Username:  username,
		Password:  password,
		Email:     accEmail,
		CreatedAt: createdAt,
	}, nil
}
