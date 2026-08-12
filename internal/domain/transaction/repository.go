package transaction

import (
	"context"
	"time"
)

type FilterTrx struct {
	UserID          UserID
	GoalID          *GoalID
	Amount          Money
	CategoryID      CategoryID
	TransactionType TransactionType
	WalletID        WalletID
	StartDate       time.Time
	EndDate         time.Time
	Limit           int
	Page            int
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
	SearchByID(ctx context.Context, trxID TransactionID, userId UserID) (*Transaction, error)
	GetBalance(ctx context.Context, userID UserID, walletID WalletID) (Money, error)
	GetMonthlySummary(ctx context.Context, userID UserID, month int, year int) (MonthlySummary, error)
	GetHighestExpense(ctx context.Context, userID UserID, month int, year int, limit int) (*Transaction, error)
	GetMostSpend(ctx context.Context, userID UserID, month int, year int, limit int) ([]*CategorySpend, error)

	Save(ctx context.Context, tx *Transaction) error
	Search(ctx context.Context, filter FilterTrx) ([]*Transaction, int, error)
	Update(ctx context.Context, tx *Transaction, userId UserID) error
	Delete(ctx context.Context, id TransactionID, userId UserID) error
}
