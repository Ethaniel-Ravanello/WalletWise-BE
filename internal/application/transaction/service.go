package transaction

import (
	"context"
	"time"
	"walletwise/internal/domain/transaction"
)

type CreateTrxInput struct {
	UserID      uint64
	GoalID      uint64
	Amount      int64
	Category    string
	Description string
	Type        string
	Source      string
	Date        time.Time
}

type TransactionDTO struct {
	ID       string    `json:"id"`
	UserID   uint64    `json:"user_id"`
	Amount   int64     `json:"amount"`
	Type     string    `json:"type"`
	Category string    `json:"category"`
	Date     time.Time `json:"date"`
}

type Service struct {
	repo transaction.Repository
}

func NewService(repo transaction.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTransaction(ctx context.Context, input *CreateTrxInput) (*transaction.Transaction, error) {

	newId := transaction.TransactionId(0)

	tx, err := transaction.NewTransaction(
		newId,
		input.UserID,
		input.GoalID,
		transaction.Money(input.Amount),
		input.Category,
		input.Description,
		transaction.Type(input.Type),
		input.Source,
		input.Date,
	)
	if err != nil {
		return nil, err
	}

	if err := s.repo.Save(ctx, tx); err != nil {
		return nil, err
	}
	return tx, nil
}
