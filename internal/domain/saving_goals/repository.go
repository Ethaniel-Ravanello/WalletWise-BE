package saving_goals

import (
	"context"
	"database/sql"
)

type Repository interface {
	Save(ctx context.Context, sg *SavingGoals) error
	SearchAll(ctx context.Context, userID UserID) ([]*SavingGoals, error)
	Update(ctx context.Context, sg *SavingGoals) error
	Delete(ctx context.Context, id SavingGoalsID) error

	SearchByID(ctx context.Context, id SavingGoalsID, userID UserID) (*SavingGoals, error)
	SearchByStatus(ctx context.Context, userID UserID, status GoalStatus) ([]*SavingGoals, error)
	UpdateAmount(ctx context.Context, tx *sql.Tx, id SavingGoalsID, amount int64) error
}

