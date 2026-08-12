package saving_goal

import (
	"errors"
	"time"
)

type SavingGoalID uint64
type SavingGoalsID = SavingGoalID
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

type SavingGoal struct {
	id            SavingGoalID
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

type SavingGoals = SavingGoal

func NewSavingGoal(
	userID UserID,
	name string,
	targetAmount TargetAmount,
	currentAmount CurrentAmount,
	deadline time.Time,
	status GoalStatus,
	description string,
	createdAt time.Time,
	updatedAt time.Time,
) (*SavingGoal, error) {
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
		createdAt = time.Now()
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	return &SavingGoal{
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
) (*SavingGoal, error) {
	return NewSavingGoal(userID, name, targetAmount, currentAmount, deadline, status, description, createdAt, updatedAt)
}

func ReconstituteSavingGoal(
	id SavingGoalID,
	userID UserID,
	name string,
	targetAmount TargetAmount,
	currentAmount CurrentAmount,
	deadline time.Time,
	status GoalStatus,
	description string,
	createdAt time.Time,
	updatedAt time.Time,
) *SavingGoal {
	return &SavingGoal{
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

func Reconstitute(
	id SavingGoalID,
	userID UserID,
	name string,
	targetAmount TargetAmount,
	currentAmount CurrentAmount,
	deadline time.Time,
	status GoalStatus,
	description string,
	createdAt time.Time,
	updatedAt time.Time,
) *SavingGoal {
	return ReconstituteSavingGoal(id, userID, name, targetAmount, currentAmount, deadline, status, description, createdAt, updatedAt)
}

func (s *SavingGoal) ID() SavingGoalID              { return s.id }
func (s *SavingGoal) UserID() UserID               { return s.userID }
func (s *SavingGoal) Name() string                 { return s.name }
func (s *SavingGoal) TargetAmount() TargetAmount   { return s.targetAmount }
func (s *SavingGoal) CurrentAmount() CurrentAmount { return s.currentAmount }
func (s *SavingGoal) Deadline() time.Time          { return s.deadline }
func (s *SavingGoal) Status() GoalStatus           { return s.status }
func (s *SavingGoal) Description() string          { return s.description }
func (s *SavingGoal) CreatedAt() time.Time         { return s.createdAt }
func (s *SavingGoal) UpdatedAt() time.Time         { return s.updatedAt }


