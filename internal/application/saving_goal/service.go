package saving_goal

import (
	"context"
	"time"
	"walletwise/internal/domain/saving_goals"
)

// DTO untuk Input Create
type SgInput struct {
	UserID        saving_goals.UserID
	Name          string
	TargetAmount  saving_goals.TargetAmount
	CurrentAmount saving_goals.CurrentAmount
	Deadline      time.Time
	GoalStatus    saving_goals.GoalStatus
	Description   string
}

// DTO untuk Input Update
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

// 1. Create Transaction
func (s *Service) CreateTransaction(ctx context.Context, input *SgInput) (*saving_goals.SavingGoals, error) {
	now := time.Now()

	// Panggil validasi dari Entity
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

	return sg, nil
}

// 2. Get All Goals
func (s *Service) GetAllGoals(ctx context.Context, userID saving_goals.UserID) ([]*saving_goals.SavingGoals, error) {
	return s.repo.SearchAll(ctx, userID)
}

// 3. Get Goal By ID
func (s *Service) GetGoalByID(ctx context.Context, id saving_goals.SavingGoalsID, userID saving_goals.UserID) (*saving_goals.SavingGoals, error) {
	return s.repo.SearchByID(ctx, id, userID)
}

// 4. Get Goals By Status
func (s *Service) GetGoalsByStatus(ctx context.Context, userID saving_goals.UserID, status saving_goals.GoalStatus) ([]*saving_goals.SavingGoals, error) {
	return s.repo.SearchByStatus(ctx, userID, status)
}

// 5. Update Goal
func (s *Service) UpdateGoal(ctx context.Context, input *SgUpdate) (*saving_goals.SavingGoals, error) {
	now := time.Now()

	// Gunakan Reconstitute karena kita sudah punya ID dan ini bukan pembuatan data dari nol
	sg := saving_goals.Reconstitute(
		input.GoalID,
		input.UserID,
		input.Name,
		input.TargetAmount,
		input.CurrentAmount,
		input.Deadline,
		input.GoalStatus,
		input.Description,
		time.Time{}, // CreatedAt biasanya tidak diupdate, biarkan kosong atau query dulu kalau butuh
		now,         // UpdatedAt diperbarui
	)

	if err := s.repo.Update(ctx, sg); err != nil {
		return nil, err
	}

	return sg, nil
}

// 6. Delete Goal
func (s *Service) DeleteGoal(ctx context.Context, id saving_goals.SavingGoalsID) error {
	return s.repo.Delete(ctx, id)
}
