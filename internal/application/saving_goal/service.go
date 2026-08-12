package saving_goal

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"walletwise/internal/domain/saving_goal"
)

type SavingGoalInput struct {
	UserID        saving_goal.UserID
	Name          string
	TargetAmount  saving_goal.TargetAmount
	CurrentAmount saving_goal.CurrentAmount
	Deadline      time.Time
	GoalStatus    saving_goal.GoalStatus
	Description   string
}

// Backward compatibility type aliases
type SgInput = SavingGoalInput
type SgUpdate = UpdateSavingGoalInput

type UpdateSavingGoalInput struct {
	GoalID        saving_goal.SavingGoalID
	UserID        saving_goal.UserID
	Name          string
	TargetAmount  saving_goal.TargetAmount
	CurrentAmount saving_goal.CurrentAmount
	Deadline      time.Time
	GoalStatus    saving_goal.GoalStatus
	Description   string
}

type Service struct {
	repo saving_goal.Repository
}

func NewService(repo saving_goal.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateGoal(ctx context.Context, input *SavingGoalInput) (*saving_goal.SavingGoal, error) {
	now := time.Now()

	sg, err := saving_goal.NewSavingGoal(
		input.UserID,
		input.Name,
		input.TargetAmount,
		input.CurrentAmount,
		input.Deadline,
		input.GoalStatus,
		input.Description,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("create saving goal entity: %w", err)
	}

	if err := s.repo.Save(ctx, sg); err != nil {
		return nil, fmt.Errorf("save saving goal: %w", err)
	}

	return sg, nil
}

func (s *Service) GetAllGoals(ctx context.Context, userID saving_goal.UserID) ([]*saving_goal.SavingGoal, error) {
	goals, err := s.repo.SearchAll(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get all saving goals: %w", err)
	}
	return goals, nil
}

func (s *Service) GetGoalByID(ctx context.Context, id saving_goal.SavingGoalID, userID saving_goal.UserID) (*saving_goal.SavingGoal, error) {
	goal, err := s.repo.SearchByID(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("get saving goal by id: %w", err)
	}
	return goal, nil
}

func (s *Service) GetGoalsByStatus(ctx context.Context, userID saving_goal.UserID, status saving_goal.GoalStatus) ([]*saving_goal.SavingGoal, error) {
	goals, err := s.repo.SearchByStatus(ctx, userID, status)
	if err != nil {
		return nil, fmt.Errorf("get saving goals by status: %w", err)
	}
	return goals, nil
}

func (s *Service) UpdateGoal(ctx context.Context, input *UpdateSavingGoalInput, userId saving_goal.UserID) (*saving_goal.SavingGoal, error) {
	if input.UserID != userId {
		return nil, fmt.Errorf("invalid user ID")
	}

	input.UserID = userId

	existingGoal, err := s.repo.SearchByID(ctx, input.GoalID, input.UserID)
	createdAt := time.Now()
	if err == nil && existingGoal != nil {
		createdAt = existingGoal.CreatedAt()
	}

	now := time.Now()
	sg := saving_goal.ReconstituteSavingGoal(
		input.GoalID,
		input.UserID,
		input.Name,
		input.TargetAmount,
		input.CurrentAmount,
		input.Deadline,
		input.GoalStatus,
		input.Description,
		createdAt,
		now,
	)

	if err := s.repo.Update(ctx, sg, userId); err != nil {
		return nil, fmt.Errorf("update saving goal: %w", err)
	}

	return sg, nil
}

func (s *Service) DeleteGoal(ctx context.Context, id saving_goal.SavingGoalID, userId saving_goal.UserID) error {
	if err := s.repo.Delete(ctx, id, userId); err != nil {
		return fmt.Errorf("delete saving goal: %w", err)
	}
	return nil
}

func (s *Service) UpdateAmountWithTx(ctx context.Context, tx *sql.Tx, goalID uint64, amount int64, userId saving_goal.UserID) error {
	if err := s.repo.UpdateAmount(ctx, tx, saving_goal.SavingGoalID(goalID), amount, userId); err != nil {
		return fmt.Errorf("update saving goal amount with tx: %w", err)
	}
	return nil
}

