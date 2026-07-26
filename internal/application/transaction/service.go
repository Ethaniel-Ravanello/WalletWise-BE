package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"walletwise/internal/application/saving_goal"
	"walletwise/internal/domain/transaction"
)

type TrxInput struct {
	UserID          uint64
	GoalID          uint64
	Amount          int64
	CategoryID      uint64
	Description     string
	TransactionType string
	WalletID        uint64
	Date            time.Time
}

type TrxUpdate struct {
	ID              uint64
	GoalID          uint64
	Amount          int64
	CategoryID      uint64
	Description     string
	TransactionType string
	WalletID        uint64
	Date            time.Time
}

type GetTransactionsInput struct {
	UserID          uint64
	GoalID          *uint64
	Amount          *uint64
	TransactionType *string
	CategoryID      *uint64
	WalletID        *uint64
	StartDate       *time.Time
	EndDate         *time.Time
	Limit           int
}

type Service struct {
	repo        transaction.Repository
	db          *sql.DB
	goalService saving_goal.Service
}

func NewService(repo transaction.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTransaction(ctx context.Context, input *TrxInput) (*transaction.Transaction, error) {

	var goalIDPtr *transaction.GoalID
	if input.GoalID != 0 {
		gID := transaction.GoalID(input.GoalID)
		goalIDPtr = &gID
	}

	tx, err := transaction.NewTransaction(
		transaction.UserID(input.UserID),
		goalIDPtr,
		transaction.Money(input.Amount),
		transaction.CategoryID(input.CategoryID),
		input.Description,
		transaction.TransactionType(input.TransactionType),
		transaction.WalletID(input.WalletID),
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
	newGoalID := transaction.GoalID(*input.GoalID)
	newAmount := transaction.Money(*input.Amount)
	newCategoryID := transaction.CategoryID(*input.CategoryID)
	newTrasactionType := transaction.TransactionType(*input.TransactionType)
	newWalletID := transaction.WalletID(*input.WalletID)

	trx, err := s.repo.Search(ctx, transaction.FilterTrx{
		UserID:          transaction.UserID(input.UserID),
		GoalID:          &newGoalID,
		Amount:          &newAmount,
		CategoryID:      &newCategoryID,
		TransactionType: &newTrasactionType,
		WalletID:        &newWalletID,
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		Limit:           input.Limit,
	})
	if err != nil {
		return nil, fmt.Errorf("error getting transaction: %w", err)
	}
	return trx, nil
}

func (s *Service) GetTransactionById(ctx context.Context, trxId transaction.TransactionID) (*transaction.Transaction, error) {
	trx, err := s.repo.SearchByID(ctx, trxId)

	if err != nil {
		return nil, fmt.Errorf("error getting transaction: %w", err)
	}
	return trx, nil
}

func (s *Service) UpdateTransaction(ctx context.Context, input *TrxUpdate) error {
	existingTrx, err := s.repo.SearchByID(ctx, transaction.TransactionID(input.ID))
	if err != nil {
		return fmt.Errorf("error getting transaction: %w", err)
	}

	var newGoalID *transaction.GoalID
	if input.GoalID != 0 {
		gID := transaction.GoalID(input.GoalID)
		newGoalID = &gID
	}

	err = existingTrx.UpdateDetails(
		newGoalID,
		transaction.Money(input.Amount),
		transaction.CategoryID(input.CategoryID),
		input.Description,
		transaction.TransactionType(input.TransactionType),
		transaction.WalletID(input.WalletID),
		input.Date)
	if err != nil {
		return fmt.Errorf("error updating transaction: %w", err)
	}
	return s.repo.Update(ctx, existingTrx)
}

func (s *Service) DeleteTransaction(ctx context.Context, trxId transaction.TransactionID) error {
	err := s.repo.Delete(ctx, trxId)
	if err != nil {
		return fmt.Errorf("error deleting transaction: %w", err)
	}
	return nil
}

func (s *Service) GetUserBalance(ctx context.Context, userId uint64, walletId uint64) (*transaction.Money, error) {
	balance, err := s.repo.GetBalance(ctx, transaction.UserID(userId), transaction.WalletID(walletId))
	if err != nil {
		return nil, fmt.Errorf("error getting users balance: %w", err)
	}
	return &balance, nil
}

func (s *Service) GetMonthlySummary(ctx context.Context, userId uint64, month int, year int) (*transaction.MonthlySummary, error) {
	summary, err := s.repo.GetMonthlySummary(ctx, transaction.UserID(userId), month, year)
	if err != nil {
		return nil, fmt.Errorf("error getting monthly summary: %w", err)
	}
	return &summary, nil
}

func (s *Service) GetHighestExpense(ctx context.Context, userId uint64, month int, year int, limit int) (*transaction.Transaction, error) {
	hiExpense, err := s.repo.GetHighestExpense(ctx, transaction.UserID(userId), month, year, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting highest expense: %w", err)
	}
	return hiExpense, nil
}

func (s *Service) GetMostSpend(ctx context.Context, userId uint64, month int, year int, limit int) ([]*transaction.CategorySpend, error) {
	catSpend, err := s.repo.GetMostSpend(ctx, transaction.UserID(userId), month, year, limit)
	if err != nil {
		return nil, fmt.Errorf("error getting most spend: %w", err)
	}
	return catSpend, nil
}
