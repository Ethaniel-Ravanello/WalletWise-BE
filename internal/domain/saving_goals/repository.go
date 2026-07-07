package saving_goals

import (
	"context"
)

type Repository interface {
	Save(ctx context.Context, sg *SavingGoals) error
	SearchAll(ctx context.Context, userId UserID) ([]*SavingGoals, error)
	Update(ctx context.Context, sg *SavingGoals) error
	Delete(ctx context.Context, id SavingGoalsID) error

	SearchByID(ctx context.Context, id SavingGoalsID, userId UserID) (*SavingGoals, error)
	SearchByStatus(ctx context.Context, userId UserID, status GoalStatus) ([]*SavingGoals, error)
}
