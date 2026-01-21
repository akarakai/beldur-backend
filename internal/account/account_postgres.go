package account

import (
	"beldur/pkg/db/postgres"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
)

type PostgresRepository struct {
	q postgres.QuerierProvider
}

func (a *PostgresRepository) UpdateLastAccess(ctx context.Context, accountId int) error {
	query := `
		UPDATE accounts
		SET last_access = NOW()
		WHERE account_id = $1;
	`

	cmdTag, err := a.q(ctx).Exec(ctx, query, accountId)
	if err != nil {
		return err
	}

	if cmdTag.RowsAffected() == 0 {
		return postgres.ErrNoRowUpdated
	}

	return nil
}

func NewPostgresRepository(q postgres.QuerierProvider) *PostgresRepository {
	return &PostgresRepository{q: q}
}

func (a *PostgresRepository) Save(ctx context.Context, acc *Account) (*Account, error) {
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
		// IMPORTANT: rely on DB uniqueness constraint; map it to a domain-level repo error
		// so the usecase can errors.Is() it and return a service error.
		if errors.Is(err, postgres.ErrUniqueValueViolation) {
			// If you want to distinguish username vs email, do it by parsing the constraint name
			// in ErrUniqueValueViolation (recommended), and return ErrUsernameAlreadyTaken / ErrEmailAlreadyTaken.
			// For now, return the generic unique-violation sentinel.
			return nil, postgres.ErrUniqueValueViolation
		}
		return nil, err
	}

	// INSERT ... RETURNING should always return a row
	if saved == nil {
		return nil, errors.New("insert account returned no row")
	}
	return saved, nil
}

func (a *PostgresRepository) FindByUsername(ctx context.Context, username string) (*Account, error) {
	query := `
		SELECT account_id, username, password, email, created_at
		FROM accounts
		WHERE username = $1
		LIMIT 1
	`

	row := a.q(ctx).QueryRow(ctx, query, username)
	return a.scanAccount(row)
}

func (a *PostgresRepository) FindById(ctx context.Context, accountId int) (*Account, error) {
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
func (a *PostgresRepository) scanAccount(row pgx.Row) (*Account, error) {
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
		// If your QuerierProvider / pg error mapper already converts pg unique errors
		// into ErrUniqueValueViolation, this will bubble up to Save() and be mapped.
		return nil, err
	}

	var accEmail Email
	if em != nil {
		// Already validated when stored; ignore error defensively
		accEmail, _ = NewEmail(*em)
	}

	return &Account{
		Id:        id,
		Username:  username,
		Password:  password,
		Email:     accEmail,
		CreatedAt: createdAt,
	}, nil
}
