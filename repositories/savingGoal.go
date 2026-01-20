package repositories

import (
	"context"
	"database/sql"
	"walletwise/models"
)

type SavingGoalRepository struct {
	db *sql.DB
}

func (r *SavingGoalRepository) GetAllSavingGoals(ctx context.Context, userId int) (*[]models.SavingGoal, error) {
	const query = `SELECT id, userId, name, amount, deadline, status, description, created_at, updated_at 
				   FROM savingGoals WHERE user_id = $1`

	var savingGoals []models.SavingGoal

	res, err := r.db.QueryContext(ctx, query, userId)
	if err != nil {
		return nil, err
	}
	defer res.Close()

	for res.Next() {
		var savingGoal models.SavingGoal
		err = res.Scan(
			&savingGoal.ID,
			&savingGoal.UserID,
			&savingGoal.Name,
			&savingGoal.Amount,
			&savingGoal.Deadline,
			&savingGoal.Status,
			&savingGoal.Description,
			&savingGoal.CreatedAt,
			&savingGoal.UpdatedAt,
		)
		savingGoals = append(savingGoals, savingGoal)
	}
	if err := res.Err(); err != nil {
		return nil, err
	}
	return &savingGoals, nil
}

func (r *SavingGoalRepository) GetSavingGoalById(ctx context.Context, savingGoalId int, userId int) (*models.SavingGoal, error) {
	const query = `SELECT id, userId, name, amount, deadline, status, description, created_at, updated_at
					WHERE id = $1 AND user_id = $2`

	var savingGoal models.SavingGoal

	err := r.db.QueryRowContext(ctx, query, savingGoalId, userId).Scan(
		&savingGoal.ID,
		&savingGoal.UserID,
		&savingGoal.Name,
		&savingGoal.Amount,
		&savingGoal.Deadline,
		&savingGoal.Status,
		&savingGoal.Description,
		&savingGoal.CreatedAt,
		&savingGoal.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &savingGoal, nil
}

//func (r *SavingGoalRepository) CreatSavingGoal(savingGoal *models.SavingGoal) error {
//	return r.db.Create(savingGoal).Error
//}
//
//func (r *SavingGoalRepository) UpdateSavingGoal(savingGoal *models.SavingGoal) error {
//	return r.db.Save(savingGoal).Error
//}
//
//func (r *SavingGoalRepository) DeleteSavingGoal(savingGoalId int) error {
//	return r.db.Delete(&models.SavingGoal{}, savingGoalId).Error
//}
