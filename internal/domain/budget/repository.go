package budget

import (
	"context"
)

type Repository interface {
	Save(ctx context.Context, budget *Budget) error
	FindByID(ctx context.Context, id BudgetID) (*Budget, error)
	FindByUserAndMonth(ctx context.Context, userID UserID, month int, year int) ([]*Budget, error)
	FindByUserAndCategory(ctx context.Context, userID UserID, categoryID CategoryID, month int, year int) (*Budget, error)
	Update(ctx context.Context, budget *Budget) error
	Delete(ctx context.Context, id BudgetID) error
}
