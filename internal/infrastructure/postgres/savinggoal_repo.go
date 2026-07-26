package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"walletwise/internal/domain/saving_goals"
)

type SavingGoalsRepo struct {
	db *sql.DB
}

func NewSavingGoalsRepo(db *sql.DB) *SavingGoalsRepo {
	return &SavingGoalsRepo{db: db}
}

var _ saving_goals.Repository = (*SavingGoalsRepo)(nil)

func (r *SavingGoalsRepo) Save(ctx context.Context, sg *saving_goals.SavingGoals) error {
	query := `INSERT INTO saving_goals (user_id, name, target_amount, current_amount, deadline, status, description, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := r.db.ExecContext(ctx, query,
		sg.UserID(),
		sg.Name(),
		sg.TargetAmount(),
		sg.CurrentAmount(),
		sg.Deadline(),
		sg.Status(),
		sg.Description(),
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to save saving goal: %w", err)
	}
	return nil
}

func (r *SavingGoalsRepo) SearchAll(ctx context.Context, userID saving_goals.UserID) ([]*saving_goals.SavingGoals, error) {
	query := `SELECT id, user_id, name, target_amount, current_amount, deadline, status, description, created_at, updated_at 
	          FROM saving_goals 
	          WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query saving goals: %w", err)
	}
	defer rows.Close()

	var savingGoals []*saving_goals.SavingGoals
	for rows.Next() {
		var (
			id            uint64
			dbUserID      uint64
			name          string
			targetAmount  int64
			currentAmount int64
			deadline      time.Time
			status        string
			description   string
			createdAt     time.Time
			updatedAt     time.Time
		)

		if err := rows.Scan(
			&id,
			&dbUserID,
			&name,
			&targetAmount,
			&currentAmount,
			&deadline,
			&status,
			&description,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan saving goal row: %w", err)
		}

		sg := saving_goals.Reconstitute(
			saving_goals.SavingGoalsID(id),
			saving_goals.UserID(dbUserID),
			name,
			saving_goals.TargetAmount(targetAmount),
			saving_goals.CurrentAmount(currentAmount),
			deadline,
			saving_goals.GoalStatus(status),
			description,
			createdAt,
			updatedAt,
		)
		savingGoals = append(savingGoals, sg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating saving goal rows: %w", err)
	}

	return savingGoals, nil
}

func (r *SavingGoalsRepo) Update(ctx context.Context, sg *saving_goals.SavingGoals) error {
	query := `UPDATE saving_goals 
	          SET user_id = $1, name = $2, target_amount = $3, current_amount = $4, deadline = $5, status = $6, description = $7, created_at = $8, updated_at = $9 
	          WHERE id = $10`

	_, err := r.db.ExecContext(ctx, query,
		sg.UserID(),
		sg.Name(),
		sg.TargetAmount(),
		sg.CurrentAmount(),
		sg.Deadline(),
		sg.Status(),
		sg.Description(),
		sg.CreatedAt(),
		time.Now(),
		sg.ID(),
	)
	if err != nil {
		return fmt.Errorf("failed to update saving goal: %w", err)
	}
	return nil
}

func (r *SavingGoalsRepo) Delete(ctx context.Context, id saving_goals.SavingGoalsID) error {
	query := `DELETE FROM saving_goals WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete saving goal: %w", err)
	}
	return nil
}

func (r *SavingGoalsRepo) SearchByID(ctx context.Context, id saving_goals.SavingGoalsID, userID saving_goals.UserID) (*saving_goals.SavingGoals, error) {
	query := `SELECT id, user_id, name, target_amount, current_amount, deadline, status, description, created_at, updated_at 
	          FROM saving_goals 
	          WHERE id = $1 AND user_id = $2`

	var (
		goalID        uint64
		dbUserID      uint64
		name          string
		targetAmount  int64
		currentAmount int64
		deadline      time.Time
		status        string
		description   string
		createdAt     time.Time
		updatedAt     time.Time
	)

	err := r.db.QueryRowContext(ctx, query, id, userID).Scan(
		&goalID,
		&dbUserID,
		&name,
		&targetAmount,
		&currentAmount,
		&deadline,
		&status,
		&description,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("saving goal not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find saving goal by id: %w", err)
	}

	sg := saving_goals.Reconstitute(
		saving_goals.SavingGoalsID(goalID),
		saving_goals.UserID(dbUserID),
		name,
		saving_goals.TargetAmount(targetAmount),
		saving_goals.CurrentAmount(currentAmount),
		deadline,
		saving_goals.GoalStatus(status),
		description,
		createdAt,
		updatedAt,
	)
	return sg, nil
}

func (r *SavingGoalsRepo) SearchByStatus(ctx context.Context, userID saving_goals.UserID, status saving_goals.GoalStatus) ([]*saving_goals.SavingGoals, error) {
	query := `SELECT id, user_id, name, target_amount, current_amount, deadline, status, description, created_at, updated_at 
	          FROM saving_goals 
	          WHERE status = $1 AND user_id = $2`

	rows, err := r.db.QueryContext(ctx, query, status, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query saving goals by status: %w", err)
	}
	defer rows.Close()

	var savingGoals []*saving_goals.SavingGoals
	for rows.Next() {
		var (
			goalID        uint64
			dbUserID      uint64
			name          string
			targetAmount  int64
			currentAmount int64
			deadline      time.Time
			statusStr     string
			description   string
			createdAt     time.Time
			updatedAt     time.Time
		)

		if err := rows.Scan(
			&goalID,
			&dbUserID,
			&name,
			&targetAmount,
			&currentAmount,
			&deadline,
			&statusStr,
			&description,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan saving goal row: %w", err)
		}

		sg := saving_goals.Reconstitute(
			saving_goals.SavingGoalsID(goalID),
			saving_goals.UserID(dbUserID),
			name,
			saving_goals.TargetAmount(targetAmount),
			saving_goals.CurrentAmount(currentAmount),
			deadline,
			saving_goals.GoalStatus(statusStr),
			description,
			createdAt,
			updatedAt,
		)
		savingGoals = append(savingGoals, sg)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating saving goal rows: %w", err)
	}

	return savingGoals, nil
}

func (r *SavingGoalsRepo) UpdateAmount(ctx context.Context, tx *sql.Tx, id saving_goals.SavingGoalsID, amount int64) error {
	query := `
        UPDATE saving_goals 
        SET current_amount = current_amount + $1, updated_at = NOW() 
        WHERE id = $2
    `

	_, err := tx.ExecContext(ctx, query, amount, id)
	if err != nil {
		return fmt.Errorf("failed to update saving goal amount: %w", err)
	}
	return nil
}
