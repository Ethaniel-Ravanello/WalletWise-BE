package transaction

import (
	"context"
	"fmt"
	"time"

	"walletwise/internal/domain/transaction"
)

type TransactionInput struct {
	UserID          uint64
	GoalID          uint64
	Amount          int64
	CategoryID      uint64
	Description     string
	TransactionType string
	WalletID        uint64
	Date            time.Time
}

type UpdateTransactionInput struct {
	ID              uint64
	UserID          uint64
	GoalID          uint64
	Amount          int64
	CategoryID      uint64
	Description     string
	TransactionType string
	WalletID        uint64
	Date            time.Time
}

// Backward compatibility type aliases
type TrxInput = TransactionInput
type TrxUpdate = UpdateTransactionInput

type GetTransactionsInput struct {
	UserID          uint64
	GoalID          *uint64
	Amount          uint64
	TransactionType string
	TransactionDate time.Time
	CategoryID      uint64
	WalletID        uint64
	StartDate       time.Time
	EndDate         time.Time
	Limit           int
	Page            int
}

type Service struct {
	repo transaction.Repository
}

func NewService(repo transaction.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateTransaction(ctx context.Context, input *TransactionInput) (*transaction.Transaction, error) {
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
		return nil, fmt.Errorf("create transaction entity: %w", err)
	}

	if err := s.repo.Save(ctx, tx); err != nil {
		return nil, fmt.Errorf("save transaction: %w", err)
	}
	return tx, nil
}

func (s *Service) GetTransaction(ctx context.Context, input GetTransactionsInput) ([]*transaction.Transaction, int, error) {
	filter := transaction.FilterTrx{
		UserID:          transaction.UserID(input.UserID),
		GoalID:          (*transaction.GoalID)(input.GoalID),
		Amount:          transaction.Money(input.Amount),
		CategoryID:      transaction.CategoryID(input.CategoryID),
		TransactionType: transaction.TransactionType(input.TransactionType),
		WalletID:        transaction.WalletID(input.WalletID),
		StartDate:       input.StartDate,
		EndDate:         input.EndDate,
		Limit:           input.Limit,
		Page:            input.Page,
	}

	if input.GoalID != nil && *input.GoalID != 0 {
		gID := transaction.GoalID(*input.GoalID)
		filter.GoalID = &gID
	}
	if input.Amount != 0 {
		amt := transaction.Money(input.Amount)
		filter.Amount = amt
	}
	if input.CategoryID != 0 {
		catID := transaction.CategoryID(input.CategoryID)
		filter.CategoryID = catID
	}
	if input.TransactionType != "" {
		tType := transaction.TransactionType(input.TransactionType)
		filter.TransactionType = tType
	}
	if input.WalletID != 0 {
		wID := transaction.WalletID(input.WalletID)
		filter.WalletID = wID
	}

	trx, currPage, err := s.repo.Search(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("search transactions: %w", err)
	}
	return trx, currPage, nil
}

func (s *Service) GetTransactionByID(ctx context.Context, trxID transaction.TransactionID, userId transaction.UserID) (*transaction.Transaction, error) {
	trx, err := s.repo.SearchByID(ctx, trxID, userId)
	if err != nil {
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}
	return trx, nil
}

func (s *Service) UpdateTransaction(ctx context.Context, input *UpdateTransactionInput, userId transaction.UserID) error {
	if transaction.UserID(input.UserID) != userId {
		return fmt.Errorf("Unauthorized User")
	}

	input.UserID = uint64(userId)

	existingTrx, err := s.repo.SearchByID(ctx, transaction.TransactionID(input.ID), userId)
	if err != nil {
		return fmt.Errorf("find transaction by id: %w", err)
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
		input.Date,
	)
	if err != nil {
		return fmt.Errorf("update transaction entity: %w", err)
	}
	if err := s.repo.Update(ctx, existingTrx, userId); err != nil {
		return fmt.Errorf("update transaction in repo: %w", err)
	}
	return nil
}

func (s *Service) DeleteTransaction(ctx context.Context, trxID transaction.TransactionID, userId transaction.UserID) error {
	if err := s.repo.Delete(ctx, trxID, userId); err != nil {
		return fmt.Errorf("delete transaction: %w", err)
	}
	return nil
}

func (s *Service) GetUserBalance(ctx context.Context, userID uint64, walletID uint64) (*transaction.Money, error) {
	balance, err := s.repo.GetBalance(ctx, transaction.UserID(userID), transaction.WalletID(walletID))
	if err != nil {
		return nil, fmt.Errorf("get user balance: %w", err)
	}
	return &balance, nil
}

func (s *Service) GetMonthlySummary(ctx context.Context, userID uint64, month int, year int) (*transaction.MonthlySummary, error) {
	summary, err := s.repo.GetMonthlySummary(ctx, transaction.UserID(userID), month, year)
	if err != nil {
		return nil, fmt.Errorf("get monthly summary: %w", err)
	}
	return &summary, nil
}

func (s *Service) GetHighestExpense(ctx context.Context, userID uint64, month int, year int, limit int) (*transaction.Transaction, error) {
	highestExpense, err := s.repo.GetHighestExpense(ctx, transaction.UserID(userID), month, year, limit)
	if err != nil {
		return nil, fmt.Errorf("get highest expense: %w", err)
	}
	return highestExpense, nil
}

func (s *Service) GetMostSpend(ctx context.Context, userID uint64, month int, year int, limit int) ([]*transaction.CategorySpend, error) {
	categorySpend, err := s.repo.GetMostSpend(ctx, transaction.UserID(userID), month, year, limit)
	if err != nil {
		return nil, fmt.Errorf("get most spend: %w", err)
	}
	return categorySpend, nil
}
