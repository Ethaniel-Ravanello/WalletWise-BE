package saving_goal

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"walletwise/internal/domain/saving_goals"
)

type SgInput struct {
	UserID        saving_goals.UserID
	Name          string
	TargetAmount  saving_goals.TargetAmount
	CurrentAmount saving_goals.CurrentAmount
	Deadline      time.Time
	GoalStatus    saving_goals.GoalStatus
	Description   string
}

type SgUpdate struct {
	GoalID        saving_goals.SavingGoalsID
	UserID        saving_goals.UserID
	Name          string
	TargetAmount  saving_goals.TargetAmount
	CurrentAmount saving_goals.CurrentAmount
	Deadline      time.Time
	GoalStatus    saving_goals.GoalStatus
	Description   string
}

type Service struct {
	repo saving_goals.Repository
}

func NewService(repo saving_goals.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateGoal(ctx context.Context, input *SgInput) (*saving_goals.SavingGoals, error) {
	now := time.Now()

	sg, err := saving_goals.NewSavingGoals(
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
		return nil, err
	}

	if err := s.repo.Save(ctx, sg); err != nil {
		return nil, err
	}
	fmt.Println("Ini Sukses last")
	return sg, nil
}

func (s *Service) GetAllGoals(ctx context.Context, userID saving_goals.UserID) ([]*saving_goals.SavingGoals, error) {
	return s.repo.SearchAll(ctx, userID)
}

func (s *Service) GetGoalByID(ctx context.Context, id saving_goals.SavingGoalsID, userID saving_goals.UserID) (*saving_goals.SavingGoals, error) {
	return s.repo.SearchByID(ctx, id, userID)
}

func (s *Service) GetGoalsByStatus(ctx context.Context, userID saving_goals.UserID, status saving_goals.GoalStatus) ([]*saving_goals.SavingGoals, error) {
	return s.repo.SearchByStatus(ctx, userID, status)
}

func (s *Service) UpdateGoal(ctx context.Context, input *SgUpdate) (*saving_goals.SavingGoals, error) {
	now := time.Now()

	sg := saving_goals.Reconstitute(
		input.GoalID,
		input.UserID,
		input.Name,
		input.TargetAmount,
		input.CurrentAmount,
		input.Deadline,
		input.GoalStatus,
		input.Description,
		time.Time{},
		now,
	)

	if err := s.repo.Update(ctx, sg); err != nil {
		return nil, err
	}

	return sg, nil
}

func (s *Service) DeleteGoal(ctx context.Context, id saving_goals.SavingGoalsID) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) UpdateAmountWithTx(ctx context.Context, tx *sql.Tx, goalID uint64, amount int64) error {
	return s.repo.UpdateAmount(ctx, tx, saving_goals.SavingGoalsID(goalID), amount)
}
