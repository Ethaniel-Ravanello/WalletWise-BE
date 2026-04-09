package service

import (
	"context"
	"fmt"
	"time"
	"walletwise/internal/domain/transaction"
)

type TrxInput struct {
	UserID      uint64
	GoalID      uint64
	Amount      int64
	Category    string
	Description string
	Type        string
	Source      string
	Date        time.Time
}

type TrxUpdate struct {
	ID          uint64
	GoalID      uint64
	Amount      int64
	Category    string
	Description string
	Type        string
	Source      string
	Date        time.Time
}

type GetTransactionsInput struct {
	UserID    uint64
	GoalID    *uint64 // pakai pointer = optional, boleh kosong
	Amount    *transaction.Money
	Type      *transaction.Type // optional
	Category  *string           // optional
	StartDate *time.Time        // optional
	EndDate   *time.Time        // optional
	Limit     int               // optional, default 10
}

type Service struct {
	repo transaction.Repository
}

func NewService(repo transaction.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTransaction(ctx context.Context, input *TrxInput) (*transaction.Transaction, error) {

	tx, err := transaction.NewTransaction(
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
		return nil, fmt.Errorf("error creating transaction: %w", err)
	}

	if err := s.repo.Save(ctx, tx); err != nil {
		return nil, fmt.Errorf("failed to save transaction: %w", err)
	}
	return tx, nil
}

func (s *Service) GetTransaction(ctx context.Context, input *GetTransactionsInput) ([]*transaction.Transaction, error) {
	trx, err := s.repo.Search(ctx, transaction.FilterTrx{
		UserID:    input.UserID,
		GoalID:    input.GoalID,
		Amount:    input.Amount,
		Category:  input.Category,
		Types:     input.Type,
		StartDate: input.StartDate,
		EndDate:   input.EndDate,
		Limit:     input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting transaction: %w", err)
	}
	return trx, nil
}

func (s *Service) GetTransactionById(ctx context.Context, trxId transaction.TransactionId) (*transaction.Transaction, error) {
	trx, err := s.repo.SearchByID(ctx, trxId)

	if err != nil {
		return nil, fmt.Errorf("error getting transaction: %w", err)
	}
	return trx, nil
}

func (s *Service) UpdateTransaction(ctx context.Context, input *TrxUpdate) error {
	existingTrx, err := s.repo.SearchByID(ctx, transaction.TransactionId(input.ID))
	if err != nil {
		return fmt.Errorf("error getting transaction: %w", err)
	}

	err = existingTrx.UpdateDetails(
		input.GoalID,
		transaction.Money(input.Amount),
		input.Category,
		input.Description,
		transaction.Type(input.Type),
		input.Source,
		input.Date)
	return err
}

func (s *Service) DeleteTransaction(ctx context.Context, trxId transaction.TransactionId) error {
	err := s.repo.Delete(ctx, trxId)
	if err != nil {
		return fmt.Errorf("error deleting transaction: %w", err)
	}
	return nil
}

func (s *Service) GetUserBalance(ctx context.Context, userId uint64) (*transaction.Money, error) {
	balance, err := s.repo.GetBalance(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error getting user balance: %w", err)
	}
	return balance, nil
}

func (s *Service) GetMonthlySummary(ctx context.Context, userId uint64, month int, year int) (*transaction.MonthlySummary, error) {
	summary, err := s.repo.GetMonthlySummary(ctx, userId, month, year)
	if err != nil {
		return nil, fmt.Errorf("error getting monthly summary: %w", err)
	}
	return summary, nil
}

func (s *Service) GetHighestExpense(ctx context.Context, userId uint64, month int, year int, limit int) (*transaction.Transaction, error) {
	hiExpense, err := s.repo.GetHighestExpense(ctx, userId, month, year, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting highest expense: %w", err)
	}
	return hiExpense, nil
}

func (s *Service) GetMostSpend(ctx context.Context, userId uint64, month int, year int, limit int) ([]*transaction.CategorySpend, error) {
	catSpend, err := s.repo.GetMostSpend(ctx, userId, month, year, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting most spend: %w", err)
	}
	return catSpend, nil
}
