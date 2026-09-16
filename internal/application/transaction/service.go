package transaction

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

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
	repo   transaction.Repository
	logger *zap.Logger
}

func NewService(repo transaction.Repository) *Service {
	return &Service{
		repo:   repo,
		logger: zap.L(),
	}
}

func (s *Service) CreateTransaction(ctx context.Context, input *TransactionInput) (*transaction.Transaction, error) {
	s.logger.Debug("Creating transaction in application service",
		zap.Uint64("UserId", input.UserID),
		zap.Uint64("WalletId", input.WalletID),
		zap.Int64("Amount", input.Amount),
		zap.String("Type", input.TransactionType))

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
		s.logger.Error("Failed to create transaction domain entity",
			zap.Error(err),
			zap.Uint64("UserId", input.UserID),
			zap.Uint64("WalletId", input.WalletID))
		return nil, fmt.Errorf("create transaction entity: %w", err)
	}

	if err := s.repo.Save(ctx, tx); err != nil {
		s.logger.Error("Failed to save transaction in repository",
			zap.Error(err),
			zap.Uint64("UserId", input.UserID),
			zap.Uint64("WalletId", input.WalletID))
		return nil, fmt.Errorf("save transaction: %w", err)
	}

	s.logger.Info("Transaction created successfully in application service",
		zap.Uint64("TransactionId", uint64(tx.ID())),
		zap.Uint64("UserId", input.UserID),
		zap.Int64("Amount", input.Amount))
	return tx, nil
}

func (s *Service) GetTransaction(ctx context.Context, input GetTransactionsInput) ([]*transaction.Transaction, int, error) {
	s.logger.Debug("Searching transactions with filters",
		zap.Uint64("UserId", input.UserID),
		zap.Int("limit", input.Limit),
		zap.Int("page", input.Page))

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
		s.logger.Error("Failed to search transactions in repository",
			zap.Error(err),
			zap.Uint64("UserId", input.UserID))
		return nil, 0, fmt.Errorf("search transactions: %w", err)
	}

	s.logger.Debug("Transactions retrieved successfully",
		zap.Uint64("UserId", input.UserID),
		zap.Int("count", len(trx)),
		zap.Int("page", currPage))
	return trx, currPage, nil
}

func (s *Service) GetTransactionByID(ctx context.Context, trxID transaction.TransactionID, userId transaction.UserID) (*transaction.Transaction, error) {
	s.logger.Debug("Fetching transaction by ID from repository",
		zap.Uint64("TransactionId", uint64(trxID)),
		zap.Uint64("UserId", uint64(userId)))

	trx, err := s.repo.SearchByID(ctx, trxID, userId)
	if err != nil {
		s.logger.Warn("Transaction not found by ID",
			zap.Error(err),
			zap.Uint64("TransactionId", uint64(trxID)),
			zap.Uint64("UserId", uint64(userId)))
		return nil, fmt.Errorf("get transaction by id: %w", err)
	}

	s.logger.Debug("Transaction retrieved successfully",
		zap.Uint64("TransactionId", uint64(trxID)),
		zap.Uint64("UserId", uint64(userId)))
	return trx, nil
}

func (s *Service) UpdateTransaction(ctx context.Context, input *UpdateTransactionInput, userId transaction.UserID) error {
	if transaction.UserID(input.UserID) != userId {
		s.logger.Warn("Unauthorized transaction update: user mismatch",
			zap.Uint64("InputUserId", input.UserID),
			zap.Uint64("UserId", uint64(userId)))
		return fmt.Errorf("Unauthorized User")
	}

	input.UserID = uint64(userId)

	s.logger.Debug("Updating transaction",
		zap.Uint64("TransactionId", input.ID),
		zap.Uint64("UserId", uint64(userId)))

	existingTrx, err := s.repo.SearchByID(ctx, transaction.TransactionID(input.ID), userId)
	if err != nil {
		s.logger.Warn("Transaction not found for update",
			zap.Error(err),
			zap.Uint64("TransactionId", input.ID),
			zap.Uint64("UserId", uint64(userId)))
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
		s.logger.Error("Failed to update transaction domain entity",
			zap.Error(err),
			zap.Uint64("TransactionId", input.ID))
		return fmt.Errorf("update transaction entity: %w", err)
	}

	if err := s.repo.Update(ctx, existingTrx, userId); err != nil {
		s.logger.Error("Failed to update transaction in repository",
			zap.Error(err),
			zap.Uint64("TransactionId", input.ID),
			zap.Uint64("UserId", uint64(userId)))
		return fmt.Errorf("update transaction in repo: %w", err)
	}

	s.logger.Info("Transaction updated successfully in application service",
		zap.Uint64("TransactionId", input.ID),
		zap.Uint64("UserId", uint64(userId)))
	return nil
}

func (s *Service) DeleteTransaction(ctx context.Context, trxID transaction.TransactionID, userId transaction.UserID) error {
	s.logger.Debug("Deleting transaction",
		zap.Uint64("TransactionId", uint64(trxID)),
		zap.Uint64("UserId", uint64(userId)))

	if err := s.repo.Delete(ctx, trxID, userId); err != nil {
		s.logger.Error("Failed to delete transaction in repository",
			zap.Error(err),
			zap.Uint64("TransactionId", uint64(trxID)),
			zap.Uint64("UserId", uint64(userId)))
		return fmt.Errorf("delete transaction: %w", err)
	}

	s.logger.Info("Transaction deleted successfully in application service",
		zap.Uint64("TransactionId", uint64(trxID)),
		zap.Uint64("UserId", uint64(userId)))
	return nil
}

func (s *Service) GetUserBalance(ctx context.Context, userID uint64, walletID uint64) (*transaction.Money, error) {
	s.logger.Debug("Fetching user balance from repository",
		zap.Uint64("UserId", userID),
		zap.Uint64("WalletId", walletID))

	balance, err := s.repo.GetBalance(ctx, transaction.UserID(userID), transaction.WalletID(walletID))
	if err != nil {
		s.logger.Error("Failed to get user balance from repository",
			zap.Error(err),
			zap.Uint64("UserId", userID),
			zap.Uint64("WalletId", walletID))
		return nil, fmt.Errorf("get user balance: %w", err)
	}

	s.logger.Debug("User balance retrieved successfully",
		zap.Uint64("UserId", userID),
		zap.Uint64("WalletId", walletID))
	return &balance, nil
}

func (s *Service) GetMonthlySummary(ctx context.Context, userID uint64, month int, year int) (*transaction.MonthlySummary, error) {
	s.logger.Debug("Fetching monthly summary from repository",
		zap.Uint64("UserId", userID),
		zap.Int("Month", month),
		zap.Int("Year", year))

	summary, err := s.repo.GetMonthlySummary(ctx, transaction.UserID(userID), month, year)
	if err != nil {
		s.logger.Error("Failed to get monthly summary from repository",
			zap.Error(err),
			zap.Uint64("UserId", userID),
			zap.Int("Month", month),
			zap.Int("Year", year))
		return nil, fmt.Errorf("get monthly summary: %w", err)
	}

	s.logger.Debug("Monthly summary retrieved successfully",
		zap.Uint64("UserId", userID),
		zap.Int("Month", month),
		zap.Int("Year", year))
	return &summary, nil
}

func (s *Service) GetHighestExpense(ctx context.Context, userID uint64, month int, year int, limit int) (*transaction.Transaction, error) {
	s.logger.Debug("Fetching highest expense from repository",
		zap.Uint64("UserId", userID),
		zap.Int("Month", month),
		zap.Int("Year", year),
		zap.Int("Limit", limit))

	highestExpense, err := s.repo.GetHighestExpense(ctx, transaction.UserID(userID), month, year, limit)
	if err != nil {
		s.logger.Error("Failed to get highest expense from repository",
			zap.Error(err),
			zap.Uint64("UserId", userID),
			zap.Int("Month", month),
			zap.Int("Year", year))
		return nil, fmt.Errorf("get highest expense: %w", err)
	}

	s.logger.Debug("Highest expense retrieved successfully",
		zap.Uint64("UserId", userID),
		zap.Int("Month", month),
		zap.Int("Year", year))
	return highestExpense, nil
}

func (s *Service) GetMostSpend(ctx context.Context, userID uint64, month int, year int, limit int) ([]*transaction.CategorySpend, error) {
	s.logger.Debug("Fetching most spend categories from repository",
		zap.Uint64("UserId", userID),
		zap.Int("Month", month),
		zap.Int("Year", year),
		zap.Int("Limit", limit))

	categorySpend, err := s.repo.GetMostSpend(ctx, transaction.UserID(userID), month, year, limit)
	if err != nil {
		s.logger.Error("Failed to get most spend categories from repository",
			zap.Error(err),
			zap.Uint64("UserId", userID),
			zap.Int("Month", month),
			zap.Int("Year", year))
		return nil, fmt.Errorf("get most spend: %w", err)
	}

	s.logger.Debug("Most spend categories retrieved successfully",
		zap.Uint64("UserId", userID),
		zap.Int("count", len(categorySpend)))
	return categorySpend, nil
}
