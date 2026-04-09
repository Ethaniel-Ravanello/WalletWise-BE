package transaction

import (
	"context"
	"time"
)

type FilterTrx struct {
	UserID    uint64
	GoalID    *uint64 // Optional: Filter by specific goal
	Amount    *Money
	Types     *Type      // Optional: Filter by DEBIT/CREDIT
	Category  *string    // Optional: Filter by category
	StartDate *time.Time // Optional: Range start
	EndDate   *time.Time // Optional: Range end
	Limit     int        // Pagination: items per page
}

type CategorySpend struct {
	Category string
	Total    int64
}

type MonthlySummary struct {
	TotalIncome  Money
	TotalExpense Money
}

type Repository interface {
	SearchByID(ctx context.Context, trxId TransactionId) (*Transaction, error)
	GetBalance(ctx context.Context, userId uint64) (*Money, error)
	GetMonthlySummary(ctx context.Context, userId uint64, month int, year int) (*MonthlySummary, error)
	GetHighestExpense(ctx context.Context, userId uint64, month int, year int, limit int) (*Transaction, error)
	GetMostSpend(ctx context.Context, userId uint64, month int, year int, limit int) ([]*CategorySpend, error)

	Save(ctx context.Context, tx *Transaction) error
	Search(ctx context.Context, filter FilterTrx) ([]*Transaction, error)
	Update(ctx context.Context, tx *Transaction) error
	Delete(ctx context.Context, id TransactionId) error
}
