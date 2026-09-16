package saving_goal

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"go.uber.org/zap"

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
	repo   saving_goal.Repository
	logger *zap.Logger
}

func NewService(repo saving_goal.Repository) *Service {
	return &Service{
		repo:   repo,
		logger: zap.L(),
	}
}

func (s *Service) CreateGoal(ctx context.Context, input *SavingGoalInput) (*saving_goal.SavingGoal, error) {
	s.logger.Debug("Creating saving goal in application service",
		zap.Uint64("UserId", uint64(input.UserID)),
		zap.String("Name", input.Name))

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
		s.logger.Error("Failed to create saving goal entity",
			zap.Error(err),
			zap.Uint64("UserId", uint64(input.UserID)),
			zap.String("Name", input.Name))
		return nil, fmt.Errorf("create saving goal entity: %w", err)
	}

	if err := s.repo.Save(ctx, sg); err != nil {
		s.logger.Error("Failed to save saving goal in repository",
			zap.Error(err),
			zap.Uint64("UserId", uint64(input.UserID)),
			zap.String("Name", input.Name))
		return nil, fmt.Errorf("save saving goal: %w", err)
	}

	s.logger.Info("Saving goal created successfully in application service",
		zap.Uint64("GoalId", uint64(sg.ID())),
		zap.Uint64("UserId", uint64(input.UserID)),
		zap.String("Name", input.Name))
	return sg, nil
}

func (s *Service) GetAllGoals(ctx context.Context, userID saving_goal.UserID) ([]*saving_goal.SavingGoal, error) {
	s.logger.Debug("Fetching all saving goals for user from repository", zap.Uint64("UserId", uint64(userID)))

	goals, err := s.repo.SearchAll(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get all saving goals from repository",
			zap.Error(err),
			zap.Uint64("UserId", uint64(userID)))
		return nil, fmt.Errorf("get all saving goals: %w", err)
	}

	s.logger.Debug("Saving goals retrieved successfully",
		zap.Uint64("UserId", uint64(userID)),
		zap.Int("count", len(goals)))
	return goals, nil
}

func (s *Service) GetGoalByID(ctx context.Context, id saving_goal.SavingGoalID, userID saving_goal.UserID) (*saving_goal.SavingGoal, error) {
	s.logger.Debug("Fetching saving goal by ID from repository",
		zap.Uint64("GoalId", uint64(id)),
		zap.Uint64("UserId", uint64(userID)))

	goal, err := s.repo.SearchByID(ctx, id, userID)
	if err != nil {
		s.logger.Warn("Saving goal not found by ID",
			zap.Error(err),
			zap.Uint64("GoalId", uint64(id)),
			zap.Uint64("UserId", uint64(userID)))
		return nil, fmt.Errorf("get saving goal by id: %w", err)
	}

	s.logger.Debug("Saving goal retrieved successfully",
		zap.Uint64("GoalId", uint64(id)),
		zap.Uint64("UserId", uint64(userID)))
	return goal, nil
}

func (s *Service) GetGoalsByStatus(ctx context.Context, userID saving_goal.UserID, status saving_goal.GoalStatus) ([]*saving_goal.SavingGoal, error) {
	s.logger.Debug("Fetching saving goals by status",
		zap.Uint64("UserId", uint64(userID)),
		zap.String("Status", string(status)))

	goals, err := s.repo.SearchByStatus(ctx, userID, status)
	if err != nil {
		s.logger.Error("Failed to get saving goals by status",
			zap.Error(err),
			zap.Uint64("UserId", uint64(userID)),
			zap.String("Status", string(status)))
		return nil, fmt.Errorf("get saving goals by status: %w", err)
	}

	s.logger.Debug("Saving goals by status retrieved successfully",
		zap.Uint64("UserId", uint64(userID)),
		zap.String("Status", string(status)),
		zap.Int("count", len(goals)))
	return goals, nil
}

func (s *Service) UpdateGoal(ctx context.Context, input *UpdateSavingGoalInput, userId saving_goal.UserID) (*saving_goal.SavingGoal, error) {
	if input.UserID != userId {
		s.logger.Warn("Unauthorized update saving goal: user ID mismatch",
			zap.Uint64("InputUserId", uint64(input.UserID)),
			zap.Uint64("UserId", uint64(userId)))
		return nil, fmt.Errorf("invalid user ID")
	}

	input.UserID = userId

	s.logger.Debug("Updating saving goal",
		zap.Uint64("GoalId", uint64(input.GoalID)),
		zap.Uint64("UserId", uint64(userId)))

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
		s.logger.Error("Failed to update saving goal in repository",
			zap.Error(err),
			zap.Uint64("GoalId", uint64(input.GoalID)),
			zap.Uint64("UserId", uint64(userId)))
		return nil, fmt.Errorf("update saving goal: %w", err)
	}

	s.logger.Info("Saving goal updated successfully in application service",
		zap.Uint64("GoalId", uint64(input.GoalID)),
		zap.Uint64("UserId", uint64(userId)))
	return sg, nil
}

func (s *Service) DeleteGoal(ctx context.Context, id saving_goal.SavingGoalID, userId saving_goal.UserID) error {
	s.logger.Debug("Deleting saving goal",
		zap.Uint64("GoalId", uint64(id)),
		zap.Uint64("UserId", uint64(userId)))

	if err := s.repo.Delete(ctx, id, userId); err != nil {
		s.logger.Error("Failed to delete saving goal in repository",
			zap.Error(err),
			zap.Uint64("GoalId", uint64(id)),
			zap.Uint64("UserId", uint64(userId)))
		return fmt.Errorf("delete saving goal: %w", err)
	}

	s.logger.Info("Saving goal deleted successfully in application service",
		zap.Uint64("GoalId", uint64(id)),
		zap.Uint64("UserId", uint64(userId)))
	return nil
}

func (s *Service) UpdateAmountWithTx(ctx context.Context, tx *sql.Tx, goalID uint64, amount int64, userId saving_goal.UserID) error {
	s.logger.Debug("Updating saving goal amount with transaction",
		zap.Uint64("GoalId", goalID),
		zap.Int64("Amount", amount),
		zap.Uint64("UserId", uint64(userId)))

	if err := s.repo.UpdateAmount(ctx, tx, saving_goal.SavingGoalID(goalID), amount, userId); err != nil {
		s.logger.Error("Failed to update saving goal amount with tx in repository",
			zap.Error(err),
			zap.Uint64("GoalId", goalID),
			zap.Uint64("UserId", uint64(userId)))
		return fmt.Errorf("update saving goal amount with tx: %w", err)
	}

	s.logger.Info("Saving goal amount updated with tx successfully in application service",
		zap.Uint64("GoalId", goalID),
		zap.Int64("Amount", amount),
		zap.Uint64("UserId", uint64(userId)))
	return nil
}

