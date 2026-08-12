package transaction

import (
	"errors"
	"time"
)

type TransactionID uint64
type UserID uint64
type GoalID uint64
type CategoryID uint64
type TransactionType string
type WalletID uint64
type Money int64

const (
	Income  TransactionType = "income"
	Expense TransactionType = "expense"
	Saving  TransactionType = "saving"
)

func (t TransactionType) IsValid() bool {
	return t == Income || t == Expense || t == Saving
}

type Transaction struct {
	id              TransactionID
	userID          UserID
	goalID          *GoalID
	amount          Money
	categoryID      CategoryID
	description     string
	transactionType TransactionType
	walletID        WalletID
	transactionDate time.Time
	createdAt       time.Time
	updatedAt       time.Time
}

func NewTransaction(
	userID UserID,
	goalID *GoalID,
	amount Money,
	categoryID CategoryID,
	description string,
	transactionType TransactionType,
	walletID WalletID,
	transactionDate time.Time,
) (*Transaction, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	if !transactionType.IsValid() {
		return nil, errors.New("invalid transaction type")
	}
	if categoryID == 0 {
		return nil, errors.New("category ID is required")
	}
	if walletID == 0 {
		return nil, errors.New("wallet ID is required")
	}
	if transactionDate.After(time.Now()) {
		return nil, errors.New("transaction date cannot be in the future")
	}

	now := time.Now()
	return &Transaction{
		userID:          userID,
		goalID:          goalID,
		amount:          amount,
		categoryID:      categoryID,
		description:     description,
		transactionType: transactionType,
		walletID:        walletID,
		transactionDate: transactionDate,
		createdAt:       now,
		updatedAt:       now,
	}, nil
}

func (t *Transaction) UpdateDetails(
	goalID *GoalID,
	amount Money,
	categoryID CategoryID,
	description string,
	transactionType TransactionType,
	walletID WalletID,
	transactionDate time.Time,
) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}
	if !transactionType.IsValid() {
		return errors.New("invalid transaction type")
	}
	if categoryID == 0 {
		return errors.New("category ID is required")
	}
	if walletID == 0 {
		return errors.New("wallet ID is required")
	}
	if transactionDate.After(time.Now()) {
		return errors.New("transaction date cannot be in the future")
	}

	t.goalID = goalID
	t.amount = amount
	t.categoryID = categoryID
	t.description = description
	t.transactionType = transactionType
	t.walletID = walletID
	t.transactionDate = transactionDate
	t.updatedAt = time.Now()

	return nil
}

func Reconstitute(
	id TransactionID,
	userID UserID,
	goalID *GoalID,
	amount Money,
	categoryID CategoryID,
	description string,
	transactionType TransactionType,
	walletID WalletID,
	transactionDate, createdAt, updatedAt time.Time,
) *Transaction {
	return &Transaction{
		id:              id,
		userID:          userID,
		goalID:          goalID,
		amount:          amount,
		categoryID:      categoryID,
		description:     description,
		transactionType: transactionType,
		walletID:        walletID,
		transactionDate: transactionDate,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
	}
}

func (t *Transaction) ID() TransactionID {
	return t.id
}
func (t *Transaction) UserID() UserID {
	return t.userID
}
func (t *Transaction) GoalID() *GoalID { return t.goalID }
func (t *Transaction) Amount() Money {
	return t.amount
}
func (t *Transaction) CategoryID() CategoryID {
	return t.categoryID
}
func (t *Transaction) Description() string {
	return t.description
}
func (t *Transaction) TransactionType() TransactionType {
	return t.transactionType
}
func (t *Transaction) WalletID() WalletID { return t.walletID }
func (t *Transaction) TransactionDate() time.Time {
	return t.transactionDate
}

func (t *Transaction) IsIncome() bool {
	return t.transactionType == Income
}
func (t *Transaction) IsExpense() bool {
	return t.transactionType == Expense
}
func (t *Transaction) IsSaving() bool { return t.transactionType == Saving }
func (t *Transaction) SignedAmount() int64 {
	if t.IsIncome() {
		return int64(t.amount)
	}
	return -int64(t.amount)
}
func (t *Transaction) CreatedAt() time.Time { return t.createdAt }
func (t *Transaction) UpdatedAt() time.Time { return t.updatedAt }
