package budget

import (
	"errors"
	"time"
)

type BudgetID uint64
type UserID uint64
type CategoryID uint64

type Budget struct {
	id         BudgetID
	userID     UserID
	categoryID CategoryID
	month      int
	year       int
	amount     int64
	createdAt  time.Time
	updatedAt  time.Time
}

func NewBudget(
	userID UserID,
	categoryID CategoryID,
	month int,
	year int,
	amount int64,
	createdAt time.Time,
	updatedAt time.Time,
) (*Budget, error) {

	if userID == 0 {
		return nil, errors.New("user ID is required")
	}
	if categoryID == 0 {
		return nil, errors.New("category ID is required")
	}
	if month < 1 || month > 12 {
		return nil, errors.New("invalid month")
	}
	if year < 1970 {
		return nil, errors.New("invalid year")
	}
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	return &Budget{
		userID:     userID,
		categoryID: categoryID,
		month:      month,
		year:       year,
		amount:     amount,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}, nil
}

func ReconstituteBudget(
	id BudgetID,
	userID UserID,
	categoryID CategoryID,
	month int,
	year int,
	amount int64,
	createdAt time.Time,
	updatedAt time.Time,
) *Budget {
	return &Budget{
		id:         id,
		userID:     userID,
		categoryID: categoryID,
		month:      month,
		year:       year,
		amount:     amount,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

func (b *Budget) UpdateBudget(
	categoryID CategoryID,
	month int,
	year int,
	amount int64,
) error {
	if categoryID == 0 {
		return errors.New("category ID is required")
	}
	if month < 1 || month > 12 {
		return errors.New("invalid month")
	}
	if year < 1970 {
		return errors.New("invalid year")
	}
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	b.categoryID = categoryID
	b.month = month
	b.year = year
	b.amount = amount
	b.updatedAt = time.Now()

	return nil
}

func (b *Budget) ID() BudgetID           { return b.id }
func (b *Budget) UserID() UserID         { return b.userID }
func (b *Budget) CategoryID() CategoryID { return b.categoryID }
func (b *Budget) Month() int             { return b.month }
func (b *Budget) Year() int              { return b.year }
func (b *Budget) Amount() int64          { return b.amount }
func (b *Budget) CreatedAt() time.Time   { return b.createdAt }
func (b *Budget) UpdatedAt() time.Time   { return b.updatedAt }

