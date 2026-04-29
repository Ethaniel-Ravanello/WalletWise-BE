package categories

import "context"

type Repository interface {
	SearchAll(ctx context.Context) ([]*Categories, error)
}
