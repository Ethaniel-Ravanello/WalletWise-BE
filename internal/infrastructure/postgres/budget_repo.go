package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"walletwise/internal/domain/budget"
)

type BudgetRepo struct {
	db *sql.DB
}

func NewBudgetRepo(db *sql.DB) *BudgetRepo {
	return &BudgetRepo{db: db}
}

var _ budget.Repository = (*BudgetRepo)(nil)

func (r *BudgetRepo) CalculateTotalSpent(ctx context.Context, userID budget.UserID, categoryID budget.CategoryID, month int, year int) (int64, error) {
	query := `
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE user_id = $1 
		  AND category_id = $2 
		  AND transaction_type = 'expense'
		  AND EXTRACT(MONTH FROM transaction_date) = $3
		  AND EXTRACT(YEAR FROM transaction_date) = $4
	`

	var totalSpent int64
	err := r.db.QueryRowContext(ctx, query, userID, categoryID, month, year).Scan(&totalSpent)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total spent: %w", err)
	}

	return totalSpent, nil
}

func (r *BudgetRepo) Save(ctx context.Context, b *budget.Budget) (budget.BudgetID, error) {
	query := `
		INSERT INTO budgets (user_id, category_id, month, year, amount, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`

	var budgetID uint64
	err := r.db.QueryRowContext(ctx, query,
		b.UserID(),
		b.CategoryID(),
		b.Month(),
		b.Year(),
		b.Amount(),
		b.CreatedAt(),
		b.UpdatedAt(),
	).Scan(&budgetID)

	if err != nil {
		return 0, fmt.Errorf("failed to save budget: %w", err)
	}
	return budget.BudgetID(budgetID), nil
}

func (r *BudgetRepo) FindByID(ctx context.Context, id budget.BudgetID, userId budget.UserID) (*budget.Budget, error) {
	query := `
		SELECT id, user_id, category_id, month, year, amount, created_at, updated_at
		FROM budgets
		WHERE id = $1 AND user_id = $2
	`

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

	row := r.db.QueryRowContext(ctx, query, id, userId)
	err := row.Scan(&dbID, &dbUserID, &dbCategoryID, &dbMonth, &dbYear, &dbAmount, &dbCreatedAt, &dbUpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("budget not found")
		}
		return nil, fmt.Errorf("failed to find budget by id: %w", err)
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
		return nil, fmt.Errorf("failed to query budgets by month: %w", err)
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
			return nil, fmt.Errorf("failed to scan budget row: %w", err)
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
		return nil, fmt.Errorf("error iterating budget rows: %w", err)
	}

	return budgets, nil
}

func (r *BudgetRepo) FindByUserAndCategory(ctx context.Context, userID budget.UserID, categoryID budget.CategoryID, month int, year int) (*budget.Budget, error) {
	query := `
		SELECT id, user_id, category_id, month, year, amount, created_at, updated_at
		FROM budgets
		WHERE user_id = $1 AND category_id = $2 AND month = $3 AND year = $4
	`

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

	row := r.db.QueryRowContext(ctx, query, userID, categoryID, month, year)
	err := row.Scan(&dbID, &dbUserID, &dbCategoryID, &dbMonth, &dbYear, &dbAmount, &dbCreatedAt, &dbUpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("budget not found for this category and month")
		}
		return nil, fmt.Errorf("failed to find budget by user and category: %w", err)
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
		WHERE id = $6 AND user_id = $7
	`

	result, err := r.db.ExecContext(
		ctx, query,
		b.CategoryID(),
		b.Month(),
		b.Year(),
		b.Amount(),
		b.UpdatedAt(),
		b.ID(),
		b.UserID(),
	)
	if err != nil {
		return fmt.Errorf("failed to update budget: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("no budget updated, id might not exist")
	}

	return nil
}

func (r *BudgetRepo) Delete(ctx context.Context, id budget.BudgetID, userId budget.UserID) error {
	query := `
		DELETE FROM budgets
		WHERE id = $1 AND user_id = $2
	`

	result, err := r.db.ExecContext(ctx, query, id, userId)
	if err != nil {
		return fmt.Errorf("failed to delete budget: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("no budget deleted, id might not exist")
	}

	return nil
}
