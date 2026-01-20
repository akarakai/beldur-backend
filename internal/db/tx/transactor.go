package tx

import "context"

// https://blog.thibaut-rousseau.com/blog/sql-transactions-in-go-the-good-way/
type Transactor interface {
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}
