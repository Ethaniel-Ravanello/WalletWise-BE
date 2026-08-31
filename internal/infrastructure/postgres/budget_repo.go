package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"walletwise/internal/domain/budget"

	"go.uber.org/zap"
)

type BudgetRepo struct {
	db     *sql.DB
	logger *zap.Logger
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
		r.logger.Error("Failed To Execute Query",
			zap.Error(err),
			zap.Int("UserId", int(userID)))
		return -0, err
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
		r.logger.Error("Failed To Execute Query",
			zap.Error(err),
			zap.Int("UserId", int(b.UserID())))
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
			r.logger.Error("Budget Not Found",
				zap.Error(err),
				zap.Int("UserId", int(userId)))
			return nil, errors.New("budget not found")
		}
		r.logger.Error("Budget Not Found By Id",
			zap.Error(err),
			zap.Int("UserId", int(userId)))
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
		r.logger.Error("Failed To Execute Query",
			zap.Error(err),
			zap.Int("UserId", int(userID)))
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
			r.logger.Error("Failed To Scan Budget Row",
				zap.Error(err),
				zap.Int("UserId", int(userID)))
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
		r.logger.Error("Error Iterating Budget Row",
			zap.Error(err),
			zap.Int("UserId", int(userID)))
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
			r.logger.Error("Budget Not Found In This Category and Month",
				zap.Error(err),
				zap.Int("UserId", int(userID)))
			return nil, errors.New("budget not found for this category and month")
		}
		r.logger.Error("Failed To Find Budget By User and Category",
			zap.Error(err),
			zap.Int("UserId", int(userID)))
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
		r.logger.Error("Failed To Execute Query",
			zap.Error(err),
			zap.Int("UserId", int(b.UserID())))
		return fmt.Errorf("failed to update budget: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("Failed To Check Row Affected",
			zap.Error(err),
			zap.Int("UserId", int(b.UserID())))
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		r.logger.Error("No Budget Updated",
			zap.Error(err),
			zap.Int("UserId", int(b.UserID())))
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
		r.logger.Error("Failed To Delete Budget",
			zap.Error(err),
			zap.Int("UserId", int(userId)),
			zap.Int("ID", int(id)))
		return fmt.Errorf("failed to delete budget: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		r.logger.Error("failed to check rows affected",
			zap.Error(err),
			zap.Int("UserId", int(userId)),
			zap.Int("ID", int(id)))
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rowsAffected == 0 {
		r.logger.Error("no budget deleted",
			zap.Error(err),
			zap.Int("UserId", int(userId)),
			zap.Int("ID", int(id)))
		return errors.New("no budget deleted, id might not exist")
	}

	return nil
}
