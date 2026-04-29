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
	id              TransactionID   `json:"id"`
	userID          UserID          `json:"user_id"`
	goalID          *GoalID         `json:"goal_id,omitempty"`
	amount          Money           `json:"amount"`
	categoryID      CategoryID      `json:"category_id"`
	description     string          `json:"description"`
	transactionType TransactionType `json:"transaction_type"`
	walletID        WalletID        `json:"wallet_id"`
	transactionDate time.Time       `json:"transaction_date"`
	createdAt       time.Time       `json:"created_at"`
	updatedAt       time.Time       `json:"updated_at"`
}

func NewTransaction(
	userID UserID,
	goalID *GoalID,
	amount Money,
	categoryId CategoryID,
	description string,
	transactionType TransactionType,
	walletID WalletID,
	transactionDate time.Time,
) (*Transaction, error) {
	if userID == 0 {
		return nil, errors.New("user id required")
	}
	if amount <= 0 {
		return nil, errors.New("amount cannot be zero or negative")
	}
	if !transactionType.IsValid() {
		return nil, errors.New("invalid transaction type")
	}
	if categoryId <= 0 {
		return nil, errors.New("categories required")
	}
	if walletID <= 0 {
		return nil, errors.New("wallet id required")
	}

	if transactionDate.After(time.Now()) {
		return nil, errors.New("date cannot be in future")
	}

	return &Transaction{
		userID:          userID,
		goalID:          goalID,
		amount:          amount,
		categoryID:      categoryId,
		description:     description,
		transactionType: transactionType,
		walletID:        walletID,
		transactionDate: transactionDate,
		createdAt:       time.Now(),
		updatedAt:       time.Now(),
	}, nil
}

func (t *Transaction) UpdateDetails(
	goalID *GoalID,
	amount Money,
	categoryId CategoryID,
	description string,
	transactionType TransactionType,
	walletID WalletID,
	transactionDate time.Time) error {

	if amount <= 0 {
		return errors.New("amount cannot be zero or negative")
	}
	if !transactionType.IsValid() {
		return errors.New("invalid transaction type")
	}
	if categoryId <= 0 {
		return errors.New("categories required")
	}
	if walletID <= 0 {
		return errors.New("wallet id  required")
	}
	if transactionDate.After(time.Now()) {
		return errors.New("date cannot be in future")
	}
	t.goalID = goalID
	t.amount = amount
	t.categoryID = categoryId
	t.description = description
	t.transactionType = transactionType
	t.walletID = walletID
	t.transactionDate = transactionDate
	t.updatedAt = time.Now()

	return nil
}

func Reconstitute(
	id TransactionID,
	userId UserID,
	goalId *GoalID,
	amount Money,
	categoryid CategoryID,
	description string,
	transactionType TransactionType,
	walletID WalletID,
	transactionDate, createdAt, updatedAt time.Time,
) *Transaction {
	return &Transaction{
		id:              id,
		userID:          userId,
		goalID:          goalId,
		amount:          amount,
		categoryID:      categoryid,
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

func (t *Transaction) WalletID() string { return t.WalletID() }

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
