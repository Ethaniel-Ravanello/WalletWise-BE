package budget

import (
	"context"
)

type Repository interface {
	Save(ctx context.Context, budget *Budget) (BudgetID, error)
	FindByID(ctx context.Context, id BudgetID, userId UserID) (*Budget, error)
	FindByUserAndMonth(ctx context.Context, userID UserID, month int, year int) ([]*Budget, error)
	FindByUserAndCategory(ctx context.Context, userID UserID, categoryID CategoryID, month int, year int) (*Budget, error)
	Update(ctx context.Context, budget *Budget) error
	Delete(ctx context.Context, id BudgetID, userId UserID) error
	CalculateTotalSpent(ctx context.Context, userID UserID, categoryID CategoryID, month int, year int) (int64, error)
}
