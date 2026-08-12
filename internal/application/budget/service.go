package budget

import (
	"context"
	"errors"
	"fmt"
	"time"

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
	repo budget.Repository
}

func NewService(repo budget.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) CreateBudget(ctx context.Context, input BudgetInput) (*BudgetDetailResponse, error) {
	existingBudget, err := s.repo.FindByUserAndCategory(
		ctx,
		budget.UserID(input.UserID),
		budget.CategoryID(input.CategoryID),
		input.Month,
		input.Year,
	)

	if err == nil && existingBudget != nil {
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
		return nil, fmt.Errorf("create budget entity: %w", err)
	}

	_, err = s.repo.Save(ctx, newBudget)
	if err != nil {
		return nil, fmt.Errorf("save budget: %w", err)
	}

	totalSpent, err := s.repo.CalculateTotalSpent(ctx, budget.UserID(input.UserID), budget.CategoryID(input.CategoryID), input.Month, input.Year)
	if err != nil {
		return nil, fmt.Errorf("calculate total spent: %w", err)
	}

	return buildBudgetDetailResponse(newBudget, totalSpent), nil
}

func (s *Service) GetBudgetByID(ctx context.Context, id uint64, userId uint64) (*BudgetDetailResponse, error) {
	b, err := s.repo.FindByID(ctx, budget.BudgetID(id), budget.UserID(userId))

	if err != nil {
		return nil, fmt.Errorf("find budget by id: %w", err)
	}

	if b.UserID() != budget.UserID(userId) {
		return nil, errors.New("Unauthorized Access")
	}

	totalSpent, err := s.repo.CalculateTotalSpent(ctx, b.UserID(), b.CategoryID(), b.Month(), b.Year())
	if err != nil {
		return nil, fmt.Errorf("calculate total spent: %w", err)
	}

	return buildBudgetDetailResponse(b, totalSpent), nil
}

func (s *Service) GetBudgetsByMonth(ctx context.Context, userID uint64, month int, year int) ([]*BudgetDetailResponse, error) {
	budgets, err := s.repo.FindByUserAndMonth(ctx, budget.UserID(userID), month, year)
	if err != nil {
		return nil, fmt.Errorf("find budgets by user and month: %w", err)
	}

	if len(budgets) == 0 {
		return nil, errors.New("no budgets found for this month")
	}

	responses := make([]*BudgetDetailResponse, 0, len(budgets))
	for _, b := range budgets {
		totalSpent, err := s.repo.CalculateTotalSpent(ctx, b.UserID(), b.CategoryID(), b.Month(), b.Year())
		if err != nil {
			return nil, fmt.Errorf("calculate total spent: %w", err)
		}
		responses = append(responses, buildBudgetDetailResponse(b, totalSpent))
	}

	return responses, nil
}

func (s *Service) UpdateBudget(ctx context.Context, input BudgetUpdateInput, userId uint64) error {
	existingBudget, err := s.repo.FindByID(ctx, budget.BudgetID(input.ID), budget.UserID(userId))
	if err != nil {
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
		return fmt.Errorf("update budget entity: %w", err)
	}

	if err := s.repo.Update(ctx, existingBudget); err != nil {
		return fmt.Errorf("update budget in repo: %w", err)
	}

	return nil
}

func (s *Service) DeleteBudget(ctx context.Context, id uint64, userId uint64) error {
	if err := s.repo.Delete(ctx, budget.BudgetID(id), budget.UserID(userId)); err != nil {
		return fmt.Errorf("delete budget: %w", err)
	}
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
