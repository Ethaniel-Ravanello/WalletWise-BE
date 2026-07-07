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

func (s GoalStatus) isValid() bool {
	return s == StatusActive || s == StatusCompleted || s == StatusCancelled
}

type SavingGoals struct {
	id            SavingGoalsID `json:"id"`
	userID        UserID        `json:"user_id"`
	name          string        `json:"name"`
	targetAmount  TargetAmount  `json:"target_amount"`
	currentAmount CurrentAmount `json:"current_amount"`
	deadLine      time.Time     `json:"dead_line"`
	status        GoalStatus    `json:"status"`
	description   string        `json:"description"`
	createdAt     time.Time     `json:"created_at"`
	updatedAt     time.Time     `json:"updated_at"`
}

func NewSavingGoals(
	userID UserID,
	name string,
	targetAmount TargetAmount,
	currentAmount CurrentAmount,
	deadLine time.Time,
	status GoalStatus,
	description string,
	createdAt time.Time,
	updatedAt time.Time,
) (*SavingGoals, error) {
	if userID == 0 {
		return nil, errors.New("must provide UserID")
	}
	if name == "" {
		return nil, errors.New("must provide Name")
	}
	if targetAmount == 0 {

	}
	if currentAmount == 0 {
		return nil, errors.New("must provide TargetAmount")
	}
	if deadLine == (time.Time{}) {
		return nil, errors.New("must provide Deadline")
	}
	if status == "" {
		return nil, errors.New("must provide Status")
	}
	if description == "" {
		return nil, errors.New("must provide Description")
	}
	return &SavingGoals{
		userID:        userID,
		name:          name,
		targetAmount:  targetAmount,
		currentAmount: currentAmount,
		deadLine:      deadLine,
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
	deadLine time.Time,
	status GoalStatus,
	description string,
	createdAt time.Time,
	updatedAt time.Time) *SavingGoals {
	return &SavingGoals{
		id:            id,
		userID:        userID,
		name:          name,
		targetAmount:  targetAmount,
		currentAmount: currentAmount,
		deadLine:      deadLine,
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
func (s *SavingGoals) Deadline() time.Time          { return s.deadLine }
func (s *SavingGoals) Status() GoalStatus           { return s.status }
func (s *SavingGoals) Description() string          { return s.description }
func (s *SavingGoals) CreatedAt() time.Time         { return s.createdAt }
func (s *SavingGoals) UpdatedAt() time.Time         { return s.updatedAt }
