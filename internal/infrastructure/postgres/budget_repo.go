package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	// Sesuaikan path import domain budget lu
	"walletwise/internal/domain/budget"
)

type BudgetRepo struct {
	db *sql.DB
}

func NewBudgetRepo(db *sql.DB) *BudgetRepo {
	return &BudgetRepo{db: db}
}

func (r *BudgetRepo) Save(ctx context.Context, b *budget.Budget) error {
	query := `
		INSERT INTO budgets (user_id, category_id, month, year, amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
	`
	_, err := r.db.ExecContext(
		ctx, query,
		b.UserID(),
		b.CategoryID(),
		b.Month(),
		b.Year(),
		b.Amount(),
		b.CreatedAt(),
		b.UpdatedAt(),
	)

	if err != nil {
		return err
	}
	return nil
}

func (r *BudgetRepo) FindByID(ctx context.Context, id budget.BudgetID) (*budget.Budget, error) {
	query := `
		SELECT id, user_id, category_id, month, year, amount, created_at, updated_at
		FROM budgets
		WHERE id = $1
	`
	row := r.db.QueryRowContext(ctx, query, id)

	var (
		dbID         uint64
		dbUserID     uint64
		dbCategoryID uint64
		dbMonth      int
		dbYear       int
		dbAmount     int64
		dbCreatedAt  time.Time
		dbUpdatedAt  time.Time
	)

	err := row.Scan(&dbID, &dbUserID, &dbCategoryID, &dbMonth, &dbYear, &dbAmount, &dbCreatedAt, &dbUpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("budget not found")
		}
		return nil, err
	}

	return budget.ReconstituteBudget(
		budget.BudgetID(dbID),
		budget.UserID(dbUserID),
		budget.CategoryID(dbCategoryID),
		dbMonth,
		dbYear,
		dbAmount,
		dbCreatedAt,
		dbUpdatedAt,
	), nil
}

func (r *BudgetRepo) FindByUserAndMonth(ctx context.Context, userID budget.UserID, month int, year int) ([]*budget.Budget, error) {
	query := `
		SELECT id, user_id, category_id, month, year, amount, created_at, updated_at
		FROM budgets
		WHERE user_id = $1 AND month = $2 AND year = $3
	`
	rows, err := r.db.QueryContext(ctx, query, userID, month, year)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var budgets []*budget.Budget
	for rows.Next() {
		var (
			dbID         uint64
			dbUserID     uint64
			dbCategoryID uint64
			dbMonth      int
			dbYear       int
			dbAmount     int64
			dbCreatedAt  time.Time
			dbUpdatedAt  time.Time
		)

		if err := rows.Scan(&dbID, &dbUserID, &dbCategoryID, &dbMonth, &dbYear, &dbAmount, &dbCreatedAt, &dbUpdatedAt); err != nil {
			return nil, err
		}

		b := budget.ReconstituteBudget(
			budget.BudgetID(dbID),
			budget.UserID(dbUserID),
			budget.CategoryID(dbCategoryID),
			dbMonth,
			dbYear,
			dbAmount,
			dbCreatedAt,
			dbUpdatedAt,
		)
		budgets = append(budgets, b)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return budgets, nil
}

func (r *BudgetRepo) FindByUserAndCategory(ctx context.Context, userID budget.UserID, categoryID budget.CategoryID, month int, year int) (*budget.Budget, error) {
	query := `
		SELECT id, user_id, category_id, month, year, amount, created_at, updated_at
		FROM budgets
		WHERE user_id = $1 AND category_id = $2 AND month = $3 AND year = $4
	`
	row := r.db.QueryRowContext(ctx, query, userID, categoryID, month, year)

	var (
		dbID         uint64
		dbUserID     uint64
		dbCategoryID uint64
		dbMonth      int
		dbYear       int
		dbAmount     int64
		dbCreatedAt  time.Time
		dbUpdatedAt  time.Time
	)

	err := row.Scan(&dbID, &dbUserID, &dbCategoryID, &dbMonth, &dbYear, &dbAmount, &dbCreatedAt, &dbUpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("budget not found for this category and month")
		}
		return nil, err
	}

	return budget.ReconstituteBudget(
		budget.BudgetID(dbID),
		budget.UserID(dbUserID),
		budget.CategoryID(dbCategoryID),
		dbMonth,
		dbYear,
		dbAmount,
		dbCreatedAt,
		dbUpdatedAt,
	), nil
}

func (r *BudgetRepo) Update(ctx context.Context, b *budget.Budget) error {
	query := `
		UPDATE budgets
		SET category_id = $1, month = $2, year = $3, amount = $4, updated_at = $5
		WHERE id = $6
	`
	result, err := r.db.ExecContext(
		ctx, query,
		b.CategoryID(),
		b.Month(),
		b.Year(),
		b.Amount(),
		b.UpdatedAt(),
		b.ID(),
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no budget updated, id might not exist")
	}

	return nil
}

func (r *BudgetRepo) Delete(ctx context.Context, id budget.BudgetID) error {
	query := `
		DELETE FROM budgets
		WHERE id = $1
	`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("no budget deleted, id might not exist")
	}

	return nil
}
