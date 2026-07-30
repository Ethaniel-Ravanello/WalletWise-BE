package saving_goals

import (
	"context"
	"database/sql"
)

type Repository interface {
	Save(ctx context.Context, sg *SavingGoals) error
	SearchAll(ctx context.Context, userID UserID) ([]*SavingGoals, error)
	Update(ctx context.Context, sg *SavingGoals, userId UserID) error
	Delete(ctx context.Context, id SavingGoalsID, userId UserID) error

	SearchByID(ctx context.Context, id SavingGoalsID, userID UserID) (*SavingGoals, error)
	SearchByStatus(ctx context.Context, userID UserID, status GoalStatus) ([]*SavingGoals, error)
	UpdateAmount(ctx context.Context, tx *sql.Tx, id SavingGoalsID, amount int64, userId UserID) error
}
