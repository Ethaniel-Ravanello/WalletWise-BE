package saving_goals

import (
	"errors"
	"time"
)

type SavingGoalsID uint64
type UserID uint64
type TargetAmount int64
type CurrentAmount int64
type GoalStatus string

const (
	StatusActive    GoalStatus = "active"
	StatusCompleted GoalStatus = "completed"
	StatusCancelled GoalStatus = "cancelled"
)

func (s GoalStatus) IsValid() bool {
	return s == StatusActive || s == StatusCompleted || s == StatusCancelled
}

type SavingGoals struct {
	id            SavingGoalsID
	userID        UserID
	name          string
	targetAmount  TargetAmount
	currentAmount CurrentAmount
	deadline      time.Time
	status        GoalStatus
	description   string
	createdAt     time.Time
	updatedAt     time.Time
}

func NewSavingGoals(
	userID UserID,
	name string,
	targetAmount TargetAmount,
	currentAmount CurrentAmount,
	deadline time.Time,
	status GoalStatus,
	description string,
	createdAt time.Time,
	updatedAt time.Time,
) (*SavingGoals, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}
	if name == "" {
		return nil, errors.New("name is required")
	}
	if targetAmount <= 0 {
		return nil, errors.New("target amount must be greater than zero")
	}
	if deadline.IsZero() {
		return nil, errors.New("deadline is required")
	}
	if !status.IsValid() {
		return nil, errors.New("invalid goal status")
	}
	if createdAt.IsZero() {
		return nil, errors.New("created_at time is required")
	}
	if updatedAt.IsZero() {
		return nil, errors.New("updated_at time is required")
	}

	return &SavingGoals{
		userID:        userID,
		name:          name,
		targetAmount:  targetAmount,
		currentAmount: currentAmount,
		deadline:      deadline,
		status:        status,
		description:   description,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}, nil
}

func Reconstitute(
	id SavingGoalsID,
	userID UserID,
	name string,
	targetAmount TargetAmount,
	currentAmount CurrentAmount,
	deadline time.Time,
	status GoalStatus,
	description string,
	createdAt time.Time,
	updatedAt time.Time,
) *SavingGoals {
	return &SavingGoals{
		id:            id,
		userID:        userID,
		name:          name,
		targetAmount:  targetAmount,
		currentAmount: currentAmount,
		deadline:      deadline,
		status:        status,
		description:   description,
		createdAt:     createdAt,
		updatedAt:     updatedAt,
	}
}

func (s *SavingGoals) ID() SavingGoalsID            { return s.id }
func (s *SavingGoals) UserID() UserID               { return s.userID }
func (s *SavingGoals) Name() string                 { return s.name }
func (s *SavingGoals) TargetAmount() TargetAmount   { return s.targetAmount }
func (s *SavingGoals) CurrentAmount() CurrentAmount { return s.currentAmount }
func (s *SavingGoals) Deadline() time.Time          { return s.deadline }
func (s *SavingGoals) Status() GoalStatus           { return s.status }
func (s *SavingGoals) Description() string          { return s.description }
func (s *SavingGoals) CreatedAt() time.Time         { return s.createdAt }
func (s *SavingGoals) UpdatedAt() time.Time         { return s.updatedAt }

