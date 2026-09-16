package budget

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"walletwise/internal/domain/budget"
)

type BudgetInput struct {
	UserID     uint64
	CategoryID uint64
	Month      int
	Year       int
	Amount     int64
}

type BudgetUpdateInput struct {
	ID         uint64
	CategoryID uint64
	Month      int
	Year       int
	Amount     int64
}

type BudgetDetailResponse struct {
	UserID        uint64 `json:"user_id"`
	CategoryID    uint64 `json:"category_id"`
	Month         int    `json:"month"`
	Year          int    `json:"year"`
	MaxAmount     int64  `json:"max_amount"`
	CurrentAmount int64  `json:"current_amount"`
	Remaining     int64  `json:"remaining"`
}

type Service struct {
	repo   budget.Repository
	logger *zap.Logger
}

func NewService(repo budget.Repository) *Service {
	return &Service{
		repo:   repo,
		logger: zap.L(),
	}
}

func (s *Service) CreateBudget(ctx context.Context, input BudgetInput) (*BudgetDetailResponse, error) {
	s.logger.Debug("Checking existing budget for user and category",
		zap.Uint64("UserId", input.UserID),
		zap.Uint64("CategoryId", input.CategoryID),
		zap.Int("Month", input.Month),
		zap.Int("Year", input.Year))

	existingBudget, err := s.repo.FindByUserAndCategory(
		ctx,
		budget.UserID(input.UserID),
		budget.CategoryID(input.CategoryID),
		input.Month,
		input.Year,
	)

	if err == nil && existingBudget != nil {
		s.logger.Warn("Budget for category, month, and year already exists",
			zap.Uint64("UserId", input.UserID),
			zap.Uint64("CategoryId", input.CategoryID),
			zap.Int("Month", input.Month),
			zap.Int("Year", input.Year))
		return nil, errors.New("budget for this category in the specified month and year already exists")
	}

	newBudget, err := budget.NewBudget(
		budget.UserID(input.UserID),
		budget.CategoryID(input.CategoryID),
		input.Month,
		input.Year,
		input.Amount,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		s.logger.Error("Failed to create budget entity",
			zap.Error(err),
			zap.Uint64("UserId", input.UserID))
		return nil, fmt.Errorf("create budget entity: %w", err)
	}

	_, err = s.repo.Save(ctx, newBudget)
	if err != nil {
		s.logger.Error("Failed to save budget in repository",
			zap.Error(err),
			zap.Uint64("UserId", input.UserID))
		return nil, fmt.Errorf("save budget: %w", err)
	}

	totalSpent, err := s.repo.CalculateTotalSpent(ctx, budget.UserID(input.UserID), budget.CategoryID(input.CategoryID), input.Month, input.Year)
	if err != nil {
		s.logger.Error("Failed to calculate total spent for budget",
			zap.Error(err),
			zap.Uint64("UserId", input.UserID))
		return nil, fmt.Errorf("calculate total spent: %w", err)
	}

	s.logger.Info("Budget created successfully",
		zap.Uint64("UserId", input.UserID),
		zap.Uint64("CategoryId", input.CategoryID),
		zap.Int("Month", input.Month),
		zap.Int("Year", input.Year),
		zap.Int64("Amount", input.Amount))
	return buildBudgetDetailResponse(newBudget, totalSpent), nil
}

func (s *Service) GetBudgetByID(ctx context.Context, id uint64, userId uint64) (*BudgetDetailResponse, error) {
	s.logger.Debug("Fetching budget by ID",
		zap.Uint64("BudgetId", id),
		zap.Uint64("UserId", userId))

	b, err := s.repo.FindByID(ctx, budget.BudgetID(id), budget.UserID(userId))
	if err != nil {
		s.logger.Warn("Budget not found by ID",
			zap.Error(err),
			zap.Uint64("BudgetId", id),
			zap.Uint64("UserId", userId))
		return nil, fmt.Errorf("find budget by id: %w", err)
	}

	if b.UserID() != budget.UserID(userId) {
		s.logger.Warn("Unauthorized access attempt to budget",
			zap.Uint64("BudgetId", id),
			zap.Uint64("UserId", userId),
			zap.Uint64("OwnerId", uint64(b.UserID())))
		return nil, errors.New("Unauthorized Access")
	}

	totalSpent, err := s.repo.CalculateTotalSpent(ctx, b.UserID(), b.CategoryID(), b.Month(), b.Year())
	if err != nil {
		s.logger.Error("Failed to calculate total spent for budget",
			zap.Error(err),
			zap.Uint64("BudgetId", id))
		return nil, fmt.Errorf("calculate total spent: %w", err)
	}

	s.logger.Debug("Budget retrieved successfully",
		zap.Uint64("BudgetId", id),
		zap.Uint64("UserId", userId))
	return buildBudgetDetailResponse(b, totalSpent), nil
}

func (s *Service) GetBudgetsByMonth(ctx context.Context, userID uint64, month int, year int) ([]*BudgetDetailResponse, error) {
	s.logger.Debug("Fetching budgets by user and month",
		zap.Uint64("UserId", userID),
		zap.Int("Month", month),
		zap.Int("Year", year))

	budgets, err := s.repo.FindByUserAndMonth(ctx, budget.UserID(userID), month, year)
	if err != nil {
		s.logger.Error("Failed to find budgets by user and month",
			zap.Error(err),
			zap.Uint64("UserId", userID),
			zap.Int("Month", month),
			zap.Int("Year", year))
		return nil, fmt.Errorf("find budgets by user and month: %w", err)
	}

	if len(budgets) == 0 {
		s.logger.Warn("No budgets found for specified month",
			zap.Uint64("UserId", userID),
			zap.Int("Month", month),
			zap.Int("Year", year))
		return nil, errors.New("no budgets found for this month")
	}

	responses := make([]*BudgetDetailResponse, 0, len(budgets))
	for _, b := range budgets {
		totalSpent, err := s.repo.CalculateTotalSpent(ctx, b.UserID(), b.CategoryID(), b.Month(), b.Year())
		if err != nil {
			s.logger.Error("Failed to calculate total spent for budget in list",
				zap.Error(err),
				zap.Uint64("UserId", userID))
			return nil, fmt.Errorf("calculate total spent: %w", err)
		}
		responses = append(responses, buildBudgetDetailResponse(b, totalSpent))
	}

	s.logger.Debug("Budgets retrieved successfully for month",
		zap.Uint64("UserId", userID),
		zap.Int("count", len(responses)))
	return responses, nil
}

func (s *Service) UpdateBudget(ctx context.Context, input BudgetUpdateInput, userId uint64) error {
	s.logger.Debug("Updating budget",
		zap.Uint64("BudgetId", input.ID),
		zap.Uint64("UserId", userId))

	existingBudget, err := s.repo.FindByID(ctx, budget.BudgetID(input.ID), budget.UserID(userId))
	if err != nil {
		s.logger.Warn("Budget not found for update",
			zap.Error(err),
			zap.Uint64("BudgetId", input.ID),
			zap.Uint64("UserId", userId))
		return fmt.Errorf("find budget by id: %w", err)
	}

	if existingBudget.CategoryID() != budget.CategoryID(input.CategoryID) ||
		existingBudget.Month() != input.Month ||
		existingBudget.Year() != input.Year {

		checkDuplicate, err := s.repo.FindByUserAndCategory(
			ctx,
			existingBudget.UserID(),
			budget.CategoryID(input.CategoryID),
			input.Month,
			input.Year,
		)

		if err == nil && checkDuplicate != nil && checkDuplicate.ID() != existingBudget.ID() {
			s.logger.Warn("Duplicate budget exists for new category and month",
				zap.Uint64("BudgetId", input.ID),
				zap.Uint64("CategoryId", input.CategoryID),
				zap.Int("Month", input.Month),
				zap.Int("Year", input.Year))
			return errors.New("another budget for this category and month already exists")
		}
	}

	err = existingBudget.UpdateBudget(
		budget.CategoryID(input.CategoryID),
		input.Month,
		input.Year,
		input.Amount,
	)
	if err != nil {
		s.logger.Error("Failed to update budget entity",
			zap.Error(err),
			zap.Uint64("BudgetId", input.ID))
		return fmt.Errorf("update budget entity: %w", err)
	}

	if err := s.repo.Update(ctx, existingBudget); err != nil {
		s.logger.Error("Failed to update budget in repository",
			zap.Error(err),
			zap.Uint64("BudgetId", input.ID))
		return fmt.Errorf("update budget in repo: %w", err)
	}

	s.logger.Info("Budget updated successfully",
		zap.Uint64("BudgetId", input.ID),
		zap.Uint64("UserId", userId))
	return nil
}

func (s *Service) DeleteBudget(ctx context.Context, id uint64, userId uint64) error {
	s.logger.Debug("Deleting budget",
		zap.Uint64("BudgetId", id),
		zap.Uint64("UserId", userId))

	if err := s.repo.Delete(ctx, budget.BudgetID(id), budget.UserID(userId)); err != nil {
		s.logger.Error("Failed to delete budget in repository",
			zap.Error(err),
			zap.Uint64("BudgetId", id),
			zap.Uint64("UserId", userId))
		return fmt.Errorf("delete budget: %w", err)
	}

	s.logger.Info("Budget deleted successfully",
		zap.Uint64("BudgetId", id),
		zap.Uint64("UserId", userId))
	return nil
}

func buildBudgetDetailResponse(b *budget.Budget, totalSpent int64) *BudgetDetailResponse {
	return &BudgetDetailResponse{
		UserID:        uint64(b.UserID()),
		CategoryID:    uint64(b.CategoryID()),
		Month:         b.Month(),
		Year:          b.Year(),
		MaxAmount:     b.Amount(),
		CurrentAmount: totalSpent,
		Remaining:     b.Amount() - totalSpent,
	}
}
