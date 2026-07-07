package postgres

import (
	"context"
	"database/sql"
	"time"
	"walletwise/internal/domain/saving_goals"
)

type SavingGoalsRepo struct {
	db *sql.DB
}

func NewSavingGoalsRepo(db *sql.DB) *SavingGoalsRepo { return &SavingGoalsRepo{db: db} }

var _ saving_goals.Repository = (*SavingGoalsRepo)(nil)

func (s SavingGoalsRepo) Save(ctx context.Context, sg *saving_goals.SavingGoals) error {
	const q = `INSERT INTO saving_goals(id, user_id, name, target_amount, current_amount, deadline, status, description, created_at, updated_at)
				VALUES($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := s.db.ExecContext(ctx, q,
		sg.ID,
		sg.UserID,
		sg.Name,
		sg.TargetAmount,
		sg.CurrentAmount,
		sg.Deadline,
		sg.Status,
		sg.Description,
		time.Now(),
		time.Now())
	return err
}

func (s SavingGoalsRepo) SearchAll(ctx context.Context, userId saving_goals.UserID) ([]*saving_goals.SavingGoals, error) {
	const q = `SELECT id, user_id, name, target_amount, current_amount, deadline, status, description, created_at, updated_at 
				FROM saving_goals 
				WHERE user_id = $1`
	rows, err := s.db.QueryContext(ctx, q, userId)
	if err != nil {
		return nil, err
	}

	var (
		id, userID                  uint64
		name                        string
		targetAmount, currentAmount int64
		deadline                    time.Time
		status, description         string
		createdAt, updatedAt        time.Time
	)

	defer rows.Close()
	var savingGoals []*saving_goals.SavingGoals
	for rows.Next() {
		if err = rows.Scan(
			&id,
			&userID,
			&name,
			&targetAmount,
			&currentAmount,
			&deadline,
			&status,
			&description,
			&createdAt,
			&updatedAt); err != nil {
			return nil, err
		}
	}
	tempSg := saving_goals.Reconstitute(saving_goals.SavingGoalsID(id), saving_goals.UserID(userID), name, saving_goals.TargetAmount(targetAmount), saving_goals.CurrentAmount(currentAmount), deadline, saving_goals.GoalStatus(status), description, createdAt, updatedAt)

	savingGoals = append(savingGoals, tempSg)
	return savingGoals, nil
}

func (s SavingGoalsRepo) Update(ctx context.Context, sg *saving_goals.SavingGoals) error {
	const q = `UPDATE saving_goals SET user_id=$1, name=$2, target_amount=$3, current_amount=$4, deadline=$5, status=$6, description=$7, created_at=$8, updated_at=$9 
                    WHERE id = $10`

	_, err := s.db.ExecContext(ctx, q, sg.UserID(), sg.Name(), sg.TargetAmount(), sg.CurrentAmount(), sg.Deadline(), sg.Status(), sg.Description(), sg.CreatedAt(), time.Now(), sg.UserID())
	if err != nil {
		return err
	}
	return nil
}

func (s SavingGoalsRepo) Delete(ctx context.Context, id saving_goals.SavingGoalsID) error {
	const q = `DELETE FROM saving_goals WHERE id = $1`
	_, err := s.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	return nil
}

func (s SavingGoalsRepo) SearchByID(ctx context.Context, id saving_goals.SavingGoalsID, userId saving_goals.UserID) (*saving_goals.SavingGoals, error) {
	const q = `SELECT id, user_id, name, target_amount, current_amount, deadline, status, description, created_at, updated_at 
				FROM saving_goals 
				WHERE id = $1 AND user_id = $2`

	var (
		ID, userID                  uint64
		name                        string
		targetAmount, currentAmount int64
		deadline                    time.Time
		status, description         string
		createdAt, updatedAt        time.Time
	)

	err := s.db.QueryRowContext(ctx, q, id, userId).Scan(
		&ID,
		&userID,
		&name,
		&targetAmount,
		&currentAmount,
		&deadline,
		&status,
		&description,
		&createdAt,
		&updatedAt)
	if err != nil {
		return nil, err
	}
	tempSg := saving_goals.Reconstitute(
		saving_goals.SavingGoalsID(ID),
		saving_goals.UserID(userID),
		name,
		saving_goals.TargetAmount(targetAmount),
		saving_goals.CurrentAmount(currentAmount),
		deadline,
		saving_goals.GoalStatus(status),
		description,
		createdAt,
		updatedAt)

	return tempSg, nil
}

func (s SavingGoalsRepo) SearchByStatus(ctx context.Context, userId saving_goals.UserID, status saving_goals.GoalStatus) ([]*saving_goals.SavingGoals, error) {
	const q = `SELECT id, user_id, name, target_amount, current_amount, deadline, status, description, created_at, updated_at 
				FROM saving_goals 
				WHERE status = $1 AND user_id = $2`

	rows, err := s.db.QueryContext(ctx, q, status, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var savingGoals []*saving_goals.SavingGoals
	var (
		ID, userID                  uint64
		name                        string
		targetAmount, currentAmount int64
		deadline                    time.Time
		statuses, description       string
		createdAt, updatedAt        time.Time
	)

	for rows.Next() {
		if err = rows.Scan(
			&ID,
			&userID,
			&name,
			&targetAmount,
			&currentAmount,
			&deadline,
			&statuses,
			&description,
			&createdAt,
			&updatedAt); err != nil {
			return nil, err
		}
	}
	tempSg := saving_goals.Reconstitute(
		saving_goals.SavingGoalsID(ID),
		saving_goals.UserID(userID),
		name,
		saving_goals.TargetAmount(targetAmount),
		saving_goals.CurrentAmount(currentAmount),
		deadline,
		saving_goals.GoalStatus(statuses),
		description,
		createdAt,
		updatedAt)
	savingGoals = append(savingGoals, tempSg)
	return savingGoals, nil
}
