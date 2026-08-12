package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"walletwise/internal/domain/transaction"
)

type TransactionRepo struct {
	db *sql.DB
}

func NewTransactionRepo(db *sql.DB) *TransactionRepo {
	return &TransactionRepo{db: db}
}

var _ transaction.Repository = (*TransactionRepo)(nil)

func (r *TransactionRepo) Search(ctx context.Context, filter transaction.FilterTrx) ([]*transaction.Transaction, int, error) {
	counter := 1
	query := fmt.Sprintf("SELECT id, user_id, goal_id, category_id, amount, description, transaction_type, wallet_id, transaction_date, created_at, updated_at FROM transactions WHERE user_id = $%d", counter)
	args := []interface{}{filter.UserID}
	counter++

	if filter.GoalID != nil {
		query += fmt.Sprintf(" AND goal_id = $%d", counter)
		args = append(args, *filter.GoalID)
		counter++
	}
	if filter.Amount != 0 {
		query += fmt.Sprintf(" AND amount = $%d", counter)
		args = append(args, filter.Amount)
		counter++
	}
	if filter.CategoryID != 0 {
		query += fmt.Sprintf(" AND category_id = $%d", counter)
		args = append(args, filter.CategoryID)
		counter++
	}
	if filter.TransactionType != "" {
		query += fmt.Sprintf(" AND transaction_type = $%d", counter)
		args = append(args, filter.TransactionType)
		counter++
	}
	if !filter.StartDate.IsZero() {
		query += fmt.Sprintf(" AND transaction_date >= $%d", counter)
		args = append(args, filter.StartDate)
		counter++
	}
	if !filter.EndDate.IsZero() {
		query += fmt.Sprintf(" AND transaction_date <= $%d", counter)
		args = append(args, filter.EndDate)
		counter++
	}
	if filter.WalletID != 0 {
		query += fmt.Sprintf(" AND wallet_id = $%d", counter)
		args = append(args, filter.WalletID)
		counter++
	}

	countQuery := strings.Replace(query, "SELECT id, user_id, goal_id, category_id, amount, description, transaction_type, wallet_id, transaction_date, created_at, updated_at", "SELECT COUNT(id)", 1)

	var totalData int
	// Eksekusi khusus untuk hitung jumlah
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalData); err != nil {
		return nil, 0, fmt.Errorf("failed to count total transactions: %w", err)
	}

	// ==========================================
	// 3. Lanjut Pasang Sorting & Pagination
	// ==========================================
	query += " ORDER BY transaction_date DESC"

	if filter.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", counter)
		args = append(args, filter.Limit)
		counter++

		// Pasang OFFSET dari field Page yang baru ditambahin
		page := filter.Page
		if page < 1 {
			page = 1 // Biar kalau user gak ngirim page, default ke halaman 1
		}
		offset := (page - 1) * filter.Limit
		query += fmt.Sprintf(" OFFSET $%d", counter)
		args = append(args, offset)
		counter++
	}

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*transaction.Transaction
	for rows.Next() {
		var (
			id              uint64
			userID          uint64
			goalIDNull      sql.NullInt64
			categoryID      uint64
			amount          int64
			description     string
			transactionType string
			walletID        uint64
			transactionDate time.Time
			createdAt       time.Time
			updatedAt       time.Time
		)

		if err := rows.Scan(
			&id,
			&userID,
			&goalIDNull,
			&categoryID,
			&amount,
			&description,
			&transactionType,
			&walletID,
			&transactionDate,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan transaction row: %w", err)
		}

		var finalGoalID *transaction.GoalID
		if goalIDNull.Valid {
			val := transaction.GoalID(goalIDNull.Int64)
			finalGoalID = &val
		}
		trx := transaction.Reconstitute(
			transaction.TransactionID(id),
			transaction.UserID(userID),
			finalGoalID,
			transaction.Money(amount),
			transaction.CategoryID(categoryID),
			description,
			transaction.TransactionType(transactionType),
			transaction.WalletID(walletID),
			transactionDate,
			createdAt,
			updatedAt,
		)
		transactions = append(transactions, trx)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating transaction rows: %w", err)
	}

	return transactions, totalData, nil
}

func (r *TransactionRepo) SearchByID(ctx context.Context, trxID transaction.TransactionID, userId transaction.UserID) (*transaction.Transaction, error) {
	query := `SELECT id, user_id, goal_id, category_id, amount, description, transaction_type, wallet_id, transaction_date, created_at, updated_at 
	          FROM transactions WHERE id = $1 AND user_id = $2`

	var (
		id              uint64
		userID          uint64
		goalIDNull      sql.NullInt64
		categoryID      uint64
		amount          int64
		description     string
		transactionType string
		walletID        uint64
		transactionDate time.Time
		createdAt       time.Time
		updatedAt       time.Time
	)

	row := r.db.QueryRowContext(ctx, query, trxID, userId)
	err := row.Scan(
		&id,
		&userID,
		&goalIDNull,
		&categoryID,
		&amount,
		&description,
		&transactionType,
		&walletID,
		&transactionDate,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find transaction by id: %w", err)
	}

	var finalGoalID *transaction.GoalID
	if goalIDNull.Valid {
		val := transaction.GoalID(goalIDNull.Int64)
		finalGoalID = &val
	}

	trx := transaction.Reconstitute(
		transaction.TransactionID(id),
		transaction.UserID(userID),
		finalGoalID,
		transaction.Money(amount),
		transaction.CategoryID(categoryID),
		description,
		transaction.TransactionType(transactionType),
		transaction.WalletID(walletID),
		transactionDate,
		createdAt,
		updatedAt,
	)
	return trx, nil
}

func (r *TransactionRepo) GetBalance(ctx context.Context, userID transaction.UserID, walletID transaction.WalletID) (transaction.Money, error) {
	query := `SELECT COALESCE(SUM(CASE WHEN transaction_type = 'income' THEN amount
	                                  WHEN transaction_type = 'expense' THEN -amount ELSE 0 END), 0) 
	          FROM transactions
	          WHERE user_id = $1 AND wallet_id = $2`

	var amount transaction.Money
	err := r.db.QueryRowContext(ctx, query, userID, walletID).Scan(&amount)
	if err != nil {
		return 0, fmt.Errorf("failed to get balance: %w", err)
	}
	return amount, nil
}

func (r *TransactionRepo) GetMonthlySummary(ctx context.Context, userID transaction.UserID, month int, year int) (transaction.MonthlySummary, error) {
	query := `SELECT
				COALESCE(SUM(CASE WHEN transaction_type = 'income' THEN amount ELSE 0 END), 0),
				COALESCE(SUM(CASE WHEN transaction_type = 'expense' THEN amount ELSE 0 END), 0)
	          FROM transactions
	          WHERE user_id = $1 
	            AND EXTRACT(MONTH FROM transaction_date) = $2 
	            AND EXTRACT(YEAR FROM transaction_date) = $3`

	var totalIncome, totalExpense transaction.Money
	err := r.db.QueryRowContext(ctx, query, userID, month, year).Scan(&totalIncome, &totalExpense)
	if err != nil {
		return transaction.MonthlySummary{}, fmt.Errorf("failed to get monthly summary: %w", err)
	}
	return transaction.MonthlySummary{
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
	}, nil
}

func (r *TransactionRepo) GetHighestExpense(ctx context.Context, userID transaction.UserID, month int, year int, limit int) (*transaction.Transaction, error) {
	query := `SELECT id, user_id, goal_id, category_id, amount, description, transaction_type, wallet_id, transaction_date, created_at, updated_at
	          FROM transactions
	          WHERE user_id = $1 
	            AND EXTRACT(MONTH FROM transaction_date) = $2 
	            AND EXTRACT(YEAR FROM transaction_date) = $3 
	            AND transaction_type = 'expense'
	          ORDER BY amount DESC
	          LIMIT $4`

	var (
		id              uint64
		dbUserID        uint64
		goalIDNull      sql.NullInt64
		categoryID      uint64
		amount          int64
		description     string
		transactionType string
		walletID        uint64
		transactionDate time.Time
		createdAt       time.Time
		updatedAt       time.Time
	)

	row := r.db.QueryRowContext(ctx, query, userID, month, year, limit)
	err := row.Scan(
		&id,
		&dbUserID,
		&goalIDNull,
		&categoryID,
		&amount,
		&description,
		&transactionType,
		&walletID,
		&transactionDate,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		return nil, fmt.Errorf("failed to get highest expense: %w", err)
	}

	var finalGoalID *transaction.GoalID
	if goalIDNull.Valid {
		val := transaction.GoalID(goalIDNull.Int64)
		finalGoalID = &val
	}

	trx := transaction.Reconstitute(
		transaction.TransactionID(id),
		transaction.UserID(dbUserID),
		finalGoalID,
		transaction.Money(amount),
		transaction.CategoryID(categoryID),
		description,
		transaction.TransactionType(transactionType),
		transaction.WalletID(walletID),
		transactionDate,
		createdAt,
		updatedAt,
	)
	return trx, nil
}

func (r *TransactionRepo) GetMostSpend(ctx context.Context, userID transaction.UserID, month int, year int, limit int) ([]*transaction.CategorySpend, error) {
	query := `SELECT c.name, COALESCE(SUM(t.amount), 0) AS total
	          FROM transactions t
	          JOIN categories c ON t.category_id = c.id
	          WHERE t.user_id = $1 
	            AND EXTRACT(YEAR FROM t.transaction_date) = $2 
	            AND EXTRACT(MONTH FROM t.transaction_date) = $3
	          GROUP BY c.name
	          ORDER BY total DESC
	          LIMIT $4`

	rows, err := r.db.QueryContext(ctx, query, userID, year, month, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get most spend categories: %w", err)
	}
	defer rows.Close()

	var categorySpends []*transaction.CategorySpend
	for rows.Next() {
		var cs transaction.CategorySpend
		if err := rows.Scan(&cs.Category, &cs.Total); err != nil {
			return nil, fmt.Errorf("failed to scan category spend row: %w", err)
		}
		categorySpends = append(categorySpends, &cs)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category spend rows: %w", err)
	}

	return categorySpends, nil
}

func (r *TransactionRepo) Save(ctx context.Context, trx *transaction.Transaction) error {
	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sqlTx.Rollback()

	query := `INSERT INTO transactions (user_id, goal_id, amount, category_id, description, transaction_type, wallet_id, transaction_date, created_at, updated_at)
	          VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`

	_, err = sqlTx.ExecContext(ctx, query,
		trx.UserID(),
		trx.GoalID(),
		trx.Amount(),
		trx.CategoryID(),
		trx.Description(),
		trx.TransactionType(),
		trx.WalletID(),
		trx.TransactionDate(),
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to insert transaction: %w", err)
	}

	if trx.GoalID() != nil {
		var delta int64
		if trx.TransactionType() == transaction.Expense {
			delta = -int64(trx.Amount())
		} else {
			delta = int64(trx.Amount())
		}

		updateGoalQuery := `
			UPDATE saving_goals
			SET current_amount = current_amount + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3
		`
		res, err := sqlTx.ExecContext(ctx, updateGoalQuery, delta, *trx.GoalID(), trx.UserID())
		if err != nil {
			return fmt.Errorf("failed to update saving goal amount: %w", err)
		}

		//BREADCRUMS = NEXT DAY NGE REFACTOR SI CREATE TRX BUAT NGE AFFECT BUDGET DAN WALLETS, KARENA INI BARU SAVING GOALS DOANG YANG KE AFFECT, INI TERMASUK UPDATE DAN DELETE

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check affected rows for saving goal: %w", err)
		}
		if rowsAffected == 0 {
			return errors.New("saving goal not found or does not belong to user")
		}
	}

	if trx.CategoryID() != 0 {
		var delta int64
		if trx.TransactionType() == transaction.Expense {
			delta = -int64(trx.Amount())
		} else {
			delta = int64(trx.Amount())
		}

		month := int(trx.UpdatedAt().Month())
		year := trx.UpdatedAt().Year()

		updateBudgetQuery := `
			UPDATE budgets
			SET amount = amount + $1, updated_at = NOW()
			WHERE user_id = $2 AND category_id = $3 AND month = $4 AND year = $5
		`
		_, err := sqlTx.ExecContext(ctx, updateBudgetQuery, delta, trx.UserID(), trx.CategoryID(), month, year)
		if err != nil {
			return fmt.Errorf("failed to update budget amount: %w", err)
		}

		//BREADCRUMS = NEXT DAY NGE REFACTOR SI CREATE TRX BUAT NGE AFFECT BUDGET DAN WALLETS, KARENA INI BARU SAVING GOALS DOANG YANG KE AFFECT, INI TERMASUK UPDATE DAN DELET
	}

	if trx.WalletID() != 0 {
		var delta int64
		if trx.TransactionType() == transaction.Expense {
			delta = -int64(trx.Amount())
		} else {
			delta = int64(trx.Amount())
		}

		updateWalletQuery := `
			UPDATE wallets
			SET balance = balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3 
		`
		res, err := sqlTx.ExecContext(ctx, updateWalletQuery, delta, trx.WalletID(), trx.UserID())
		if err != nil {
			return fmt.Errorf("failed to update wallet amount: %w", err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check affected rows for wallet: %w", err)
		}
		if rowsAffected == 0 {
			return errors.New("wallet not found or does not belong to user")
		}
	}

	return sqlTx.Commit()
}

func (r *TransactionRepo) Update(ctx context.Context, trx *transaction.Transaction, userId transaction.UserID) error {
	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sqlTx.Rollback()

	getOldQuery := `SELECT user_id, goal_id, wallet_id, category_id, amount, transaction_type, updated_at FROM transactions WHERE id = $1 AND user_id = $2 FOR UPDATE`
	var (
		oldUserID     uint64
		oldGoalIDNull sql.NullInt64
		oldWalletID   uint64
		oldCategoryID uint64
		oldAmount     int64
		oldType       string
		oldUpdatedAt  time.Time
	)
	err = sqlTx.QueryRowContext(ctx, getOldQuery, trx.ID(), userId).Scan(
		&oldUserID,
		&oldGoalIDNull,
		&oldWalletID,
		&oldCategoryID,
		&oldAmount,
		&oldType,
		&oldUpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("transaction not found: %w", err)
		}
		return fmt.Errorf("failed to fetch old transaction: %w", err)
	}

	// Revert old saving goal
	if oldGoalIDNull.Valid {
		var reverseDelta int64
		if transaction.TransactionType(oldType) == transaction.Expense {
			reverseDelta = oldAmount
		} else {
			reverseDelta = -oldAmount
		}

		reverseGoalQuery := `
			UPDATE saving_goals
			SET current_amount = current_amount + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3
		`
		res, err := sqlTx.ExecContext(ctx, reverseGoalQuery, reverseDelta, oldGoalIDNull.Int64, oldUserID)
		if err != nil {
			return fmt.Errorf("failed to update saving goal amount during reversal: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check affected rows for saving goal reversal: %w", err)
		}
		if rowsAffected == 0 {
			return errors.New("saving goal not found or does not belong to user during reversal")
		}
	}

	// Revert old wallet
	if oldWalletID != 0 {
		var reverseDelta int64
		if transaction.TransactionType(oldType) == transaction.Expense {
			reverseDelta = oldAmount
		} else {
			reverseDelta = -oldAmount
		}

		updateWalletQuery := `
			UPDATE wallets
			SET balance = balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3
		`
		res, err := sqlTx.ExecContext(ctx, updateWalletQuery, reverseDelta, oldWalletID, oldUserID)
		if err != nil {
			return fmt.Errorf("failed to update wallet balance during reversal: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check affected rows for wallet reversal: %w", err)
		}
		if rowsAffected == 0 {
			return errors.New("wallet not found or does not belong to user during reversal")
		}
	}

	// Revert old budget
	if oldCategoryID != 0 {
		var reverseDelta int64
		if transaction.TransactionType(oldType) == transaction.Expense {
			reverseDelta = oldAmount
		} else {
			reverseDelta = -oldAmount
		}

		month := int(oldUpdatedAt.Month())
		year := oldUpdatedAt.Year()

		updateBudgetQuery := `
			UPDATE budgets
			SET amount = amount + $1, updated_at = NOW()
			WHERE user_id = $2 AND category_id = $3 AND month = $4 AND year = $5
		`
		_, err := sqlTx.ExecContext(ctx, updateBudgetQuery, reverseDelta, oldUserID, oldCategoryID, month, year)
		if err != nil {
			return fmt.Errorf("failed to update budget amount during reversal: %w", err)
		}
	}

	// Update transaction record
	now := time.Now()
	query := `UPDATE transactions SET user_id = $1, goal_id = $2, amount = $3, category_id = $4, description = $5, transaction_type = $6, wallet_id = $7, transaction_date = $8, updated_at = $9 WHERE id = $10 AND user_id = $11`

	res, err := sqlTx.ExecContext(ctx, query,
		trx.UserID(),
		trx.GoalID(),
		trx.Amount(),
		trx.CategoryID(),
		trx.Description(),
		trx.TransactionType(),
		trx.WalletID(),
		trx.TransactionDate(),
		now,
		trx.ID(),
		userId,
	)
	if err != nil {
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows for transaction update: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("transaction not found or does not belong to user")
	}

	// Apply new saving goal
	if trx.GoalID() != nil {
		var delta int64
		if trx.TransactionType() == transaction.Expense {
			delta = -int64(trx.Amount())
		} else {
			delta = int64(trx.Amount())
		}

		updateGoalQuery := `
			UPDATE saving_goals
			SET current_amount = current_amount + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3
		`
		res, err := sqlTx.ExecContext(ctx, updateGoalQuery, delta, *trx.GoalID(), trx.UserID())
		if err != nil {
			return fmt.Errorf("failed to update saving goal amount: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check affected rows for saving goal: %w", err)
		}
		if rowsAffected == 0 {
			return errors.New("saving goal not found or does not belong to user")
		}
	}

	// Apply new budget
	if trx.CategoryID() != 0 {
		var delta int64
		if trx.TransactionType() == transaction.Expense {
			delta = -int64(trx.Amount())
		} else {
			delta = int64(trx.Amount())
		}

		month := int(now.Month())
		year := now.Year()

		updateBudgetQuery := `
			UPDATE budgets
			SET amount = amount + $1, updated_at = NOW()
			WHERE user_id = $2 AND category_id = $3 AND month = $4 AND year = $5
		`
		_, err := sqlTx.ExecContext(ctx, updateBudgetQuery, delta, trx.UserID(), trx.CategoryID(), month, year)
		if err != nil {
			return fmt.Errorf("failed to update budget amount: %w", err)
		}
	}

	// Apply new wallet
	if trx.WalletID() != 0 {
		var delta int64
		if trx.TransactionType() == transaction.Expense {
			delta = -int64(trx.Amount())
		} else {
			delta = int64(trx.Amount())
		}

		updateWalletQuery := `
			UPDATE wallets
			SET balance = balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3
		`
		res, err := sqlTx.ExecContext(ctx, updateWalletQuery, delta, trx.WalletID(), trx.UserID())
		if err != nil {
			return fmt.Errorf("failed to update wallet amount: %w", err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check affected rows for wallet: %w", err)
		}
		if rowsAffected == 0 {
			return errors.New("wallet not found or does not belong to user")
		}
	}
	return sqlTx.Commit()
}

func (r *TransactionRepo) Delete(ctx context.Context, trxID transaction.TransactionID, userId transaction.UserID) error {
	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sqlTx.Rollback()

	getOldQuery := `SELECT user_id, goal_id, wallet_id, category_id, amount, transaction_type, updated_at FROM transactions WHERE id = $1 AND user_id = $2 FOR UPDATE`
	var (
		oldUserID     uint64
		oldGoalIDNull sql.NullInt64
		oldWalletID   uint64
		oldCategoryID uint64
		oldAmount     int64
		oldType       string
		oldUpdatedAt  time.Time
	)
	err = sqlTx.QueryRowContext(ctx, getOldQuery, trxID, userId).Scan(
		&oldUserID,
		&oldGoalIDNull,
		&oldWalletID,
		&oldCategoryID,
		&oldAmount,
		&oldType,
		&oldUpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fmt.Errorf("transaction not found: %w", err)
		}
		return fmt.Errorf("failed to fetch transaction to delete: %w", err)
	}

	// GoalId reversal
	if oldGoalIDNull.Valid {
		var reverseDelta int64
		if transaction.TransactionType(oldType) == transaction.Expense {
			reverseDelta = oldAmount
		} else {
			reverseDelta = -oldAmount
		}

		reverseGoalQuery := `
			UPDATE saving_goals
			SET current_amount = current_amount + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3
		`
		res, err := sqlTx.ExecContext(ctx, reverseGoalQuery, reverseDelta, oldGoalIDNull.Int64, oldUserID)
		if err != nil {
			return fmt.Errorf("failed to update saving goal amount: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check affected rows for saving goal: %w", err)
		}
		if rowsAffected == 0 {
			return errors.New("saving goal not found or does not belong to user")
		}
	}

	// Wallet reversal
	if oldWalletID != 0 {
		var reverseDelta int64
		if transaction.TransactionType(oldType) == transaction.Expense {
			reverseDelta = oldAmount
		} else {
			reverseDelta = -oldAmount
		}

		updateWalletQuery := `
			UPDATE wallets
			SET balance = balance + $1, updated_at = NOW()
			WHERE id = $2 AND user_id = $3
		`
		res, err := sqlTx.ExecContext(ctx, updateWalletQuery, reverseDelta, oldWalletID, oldUserID)
		if err != nil {
			return fmt.Errorf("failed to update wallet balance: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			return fmt.Errorf("failed to check affected rows for wallet: %w", err)
		}
		if rowsAffected == 0 {
			return errors.New("wallet not found or does not belong to user")
		}
	}

	// Budget reversal
	if oldCategoryID != 0 {
		var reverseDelta int64
		if transaction.TransactionType(oldType) == transaction.Expense {
			reverseDelta = oldAmount
		} else {
			reverseDelta = -oldAmount
		}

		month := int(oldUpdatedAt.Month())
		year := oldUpdatedAt.Year()

		updateBudgetQuery := `
			UPDATE budgets
			SET amount = amount + $1, updated_at = NOW()
			WHERE user_id = $1 AND category_id = $2 AND month = $3 AND year = $4
		`
		_, err := sqlTx.ExecContext(ctx, updateBudgetQuery, reverseDelta, oldUserID, oldCategoryID, month, year)
		if err != nil {
			return fmt.Errorf("failed to update budget amount: %w", err)
		}
	}

	// Delete transaction
	deleteQuery := `DELETE FROM transactions WHERE id = $1 AND user_id = $2`
	res, err := sqlTx.ExecContext(ctx, deleteQuery, trxID, userId)
	if err != nil {
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows for transaction deletion: %w", err)
	}
	if rowsAffected == 0 {
		return errors.New("transaction not found or does not belong to user")
	}

	return sqlTx.Commit()
}
