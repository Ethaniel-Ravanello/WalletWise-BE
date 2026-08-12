package category

import "context"

type Repository interface {
	SearchAll(ctx context.Context) ([]*Category, error)
}

