package campaign

import "context"

type Repository interface {
	// Save mutates campaign by adding id and other fields after persistence
	Save(ctx context.Context, campaign *Campaign) error
}
