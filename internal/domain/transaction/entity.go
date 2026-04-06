package transaction

import (
	"errors"
	"time"
)

type TransactionId uint64
type Type string
type Money int64

const (
	Debit  Type = "DEBIT"
	Credit Type = "CREDIT"
)

func (t Type) IsValid() bool {
	return t == Debit || t == Credit
}

type Transaction struct {
	id                TransactionId
	userID            uint64
	goalID            uint64
	amount            Money
	category          string
	description       string
	transactionType   Type
	transactionSource string
	transactionDate   time.Time
	createdAt         time.Time
	updatedAt         time.Time
}

func NewTransaction(
	id TransactionId,
	userID uint64,
	goalID uint64,
	amount Money,
	category string,
	description string,
	transactionType Type,
	transactionSource string,
	transactionDate time.Time,
) (*Transaction, error) {
	if id <= 0 {
		return nil, errors.New("transaction id cannot be empty")
	}

	if userID == 0 {
		return nil, errors.New("user id required")
	}

	if amount <= 0 {
		return nil, errors.New("amount cannot be zero or negative")
	}

	if !transactionType.IsValid() {
		return nil, errors.New("invalid transaction type")
	}

	if category == "" {
		return nil, errors.New("category required")
	}
	if transactionSource == "" {
		return nil, errors.New("source required")
	}

	if transactionDate.After(time.Now()) {
		return nil, errors.New("date cannot be in future")
	}

	return &Transaction{
		id:                id,
		userID:            userID,
		goalID:            goalID,
		amount:            amount,
		category:          category,
		description:       description,
		transactionType:   transactionType,
		transactionSource: transactionSource,
		transactionDate:   transactionDate,
		createdAt:         time.Now(),
		updatedAt:         time.Now(),
	}, nil
}

func Reconstitute(
	id TransactionId,
	userId, goalId uint64,
	amount Money,
	category string,
	description string,
	transactionType Type,
	transactionSource string,
	transactionDate, createdAt, updatedAt time.Time,
) *Transaction {
	return &Transaction{
		id:                id,
		userID:            userId,
		goalID:            goalId,
		amount:            amount,
		category:          category,
		description:       description,
		transactionType:   transactionType,
		transactionSource: transactionSource,
		transactionDate:   transactionDate,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
	}
}

func (t *Transaction) ID() TransactionId {
	return t.id
}

func (t *Transaction) UserID() uint64 {
	return t.userID
}

func (t *Transaction) GoalID() uint64 { return t.goalID }

func (t *Transaction) Amount() Money {
	return t.amount
}

func (t *Transaction) Category() string {
	return t.category
}

func (t *Transaction) Description() string {
	return t.description
}

func (t *Transaction) TransactionType() Type {
	return t.transactionType
}

func (t *Transaction) TransactionSource() string { return t.transactionSource }

func (t *Transaction) TransactionDate() time.Time {
	return t.transactionDate
}

func (t *Transaction) IsDebit() bool {
	return t.transactionType == Debit
}

func (t *Transaction) IsCredit() bool {
	return t.transactionType == Credit
}

func (t *Transaction) SignedAmount() int64 {
	if t.IsDebit() {
		return int64(t.amount)
	}
	return -int64(t.amount)
}
