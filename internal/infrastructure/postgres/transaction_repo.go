package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

	"walletwise/internal/domain/transaction"
)

type TransactionRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewTransactionRepo(db *sql.DB) *TransactionRepo {
	return &TransactionRepo{
		db:     db,
		logger: zap.L(),
	}
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

	r.logger.Debug("Search count query", zap.String("CountQuery", countQuery))
	r.logger.Debug("Search query", zap.String("Query", query))

	var totalData int
	// Eksekusi khusus untuk hitung jumlah
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&totalData); err != nil {
		r.logger.Error("Failed to execute count query",
			zap.Error(err),
			zap.Int("UserId", int(filter.UserID)),
			zap.String("TransactionType", string(filter.TransactionType)),
			zap.String("StartDate", filter.StartDate.Format(time.RFC3339)))
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
		r.logger.Error("Failed to execute query",
			zap.Error(err),
			zap.Int("UserId", int(filter.UserID)),
			zap.String("TransactionType", string(filter.TransactionType)),
			zap.String("StartDate", filter.StartDate.Format(time.RFC3339)))
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
			r.logger.Error("Failed to scan transaction row",
				zap.Error(err),
				zap.Int("UserId", int(filter.UserID)),
				zap.String("TransactionType", string(filter.TransactionType)),
				zap.String("StartDate", filter.StartDate.Format(time.RFC3339)))
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
		r.logger.Error("Failed to iterate transaction rows",
			zap.Error(err),
			zap.Int("UserId", int(filter.UserID)),
			zap.String("TransactionType", string(filter.TransactionType)),
			zap.String("StartDate", filter.StartDate.Format(time.RFC3339)))
		return nil, 0, fmt.Errorf("error iterating transaction rows: %w", err)
	}

	return transactions, totalData, nil
}

func (r *TransactionRepo) SearchByID(ctx context.Context, trxID transaction.TransactionID, userId transaction.UserID) (*transaction.Transaction, error) {
	query := `SELECT id, user_id, goal_id, category_id, amount, description, transaction_type, wallet_id, transaction_date, created_at, updated_at 
	          FROM transactions WHERE id = $1 AND user_id = $2`
	r.logger.Debug("SearchByID query", zap.String("Query", query))
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
			r.logger.Warn("Transaction not found",
				zap.Int("TransactionId", int(trxID)),
				zap.Int("UserId", int(userId)))
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		r.logger.Error("Failed to find transaction by id",
			zap.Error(err),
			zap.Int("UserId", int(userId)),
			zap.Int("TransactionId", int(trxID)))
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
	r.logger.Debug("GetBalance query", zap.String("Query", query))

	var amount transaction.Money
	err := r.db.QueryRowContext(ctx, query, userID, walletID).Scan(&amount)
	if err != nil {
		r.logger.Error("Failed to get balance",
			zap.Error(err),
			zap.Int("UserId", int(userID)),
			zap.Int("WalletId", int(walletID)))
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
	r.logger.Debug("GetMonthlySummary query", zap.String("Query", query))

	var totalIncome, totalExpense transaction.Money
	err := r.db.QueryRowContext(ctx, query, userID, month, year).Scan(&totalIncome, &totalExpense)
	if err != nil {
		r.logger.Error("Failed to get monthly summary",
			zap.Error(err),
			zap.Int("UserId", int(userID)),
			zap.Int("Month", month),
			zap.Int("Year", year))
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
	r.logger.Debug("GetHighestExpense query", zap.String("Query", query))

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
			r.logger.Warn("Highest expense transaction not found",
				zap.Int("UserId", int(userID)),
				zap.Int("Month", month),
				zap.Int("Year", year),
				zap.Int("Limit", limit))
			return nil, fmt.Errorf("transaction not found: %w", err)
		}
		r.logger.Error("Failed to get highest expense",
			zap.Error(err),
			zap.Int("UserId", int(userID)),
			zap.Int("Month", month),
			zap.Int("Year", year),
			zap.Int("Limit", limit))
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
	            AND t.transaction_type = 'expense'
	          GROUP BY c.name
	          ORDER BY total DESC
	          LIMIT $4`
	r.logger.Debug("GetMostSpend query", zap.String("Query", query))

	rows, err := r.db.QueryContext(ctx, query, userID, year, month, limit)
	if err != nil {
		r.logger.Error("Failed to query most spend categories",
			zap.Error(err),
			zap.Int("UserId", int(userID)),
			zap.Int("Month", month),
			zap.Int("Year", year),
			zap.Int("Limit", limit))
		return nil, fmt.Errorf("failed to get most spend categories: %w", err)
	}
	defer rows.Close()

	var categorySpends []*transaction.CategorySpend
	for rows.Next() {
		var cs transaction.CategorySpend
		if err := rows.Scan(&cs.Category, &cs.Total); err != nil {
			r.logger.Error("Failed to scan category spend row",
				zap.Error(err),
				zap.Int("UserId", int(userID)),
				zap.Int("Month", month),
				zap.Int("Year", year),
				zap.Int("Limit", limit))
			return nil, fmt.Errorf("failed to scan category spend row: %w", err)
		}
		categorySpends = append(categorySpends, &cs)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating category spend rows",
			zap.Error(err),
			zap.Int("UserId", int(userID)),
			zap.Int("Month", month),
			zap.Int("Year", year),
			zap.Int("Limit", limit))
		return nil, fmt.Errorf("error iterating category spend rows: %w", err)
	}

	return categorySpends, nil
}

func (r *TransactionRepo) Save(ctx context.Context, trx *transaction.Transaction) error {
	r.logger.Debug("Saving transaction",
		zap.Int("UserId", int(trx.UserID())),
		zap.String("TransactionType", string(trx.TransactionType())),
		zap.Int64("Amount", int64(trx.Amount())))

	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction",
			zap.Error(err),
			zap.Int("UserId", int(trx.UserID())))
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
		r.logger.Error("Failed to insert transaction",
			zap.Error(err),
			zap.Int("UserId", int(trx.UserID())),
			zap.String("TransactionType", string(trx.TransactionType())),
			zap.Int64("Amount", int64(trx.Amount())))
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
			r.logger.Error("Failed to update saving goal amount",
				zap.Error(err),
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("GoalId", int(*trx.GoalID())),
				zap.Int64("Delta", delta))
			return fmt.Errorf("failed to update saving goal amount: %w", err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("Failed to check affected rows for saving goal",
				zap.Error(err),
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("GoalId", int(*trx.GoalID())))
			return fmt.Errorf("failed to check affected rows for saving goal: %w", err)
		}
		if rowsAffected == 0 {
			r.logger.Warn("Saving goal not found or does not belong to user",
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("GoalId", int(*trx.GoalID())))
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

		month := int(trx.TransactionDate().Month())
		year := trx.TransactionDate().Year()

		updateBudgetQuery := `
			UPDATE budgets
			SET amount = amount + $1, updated_at = NOW()
			WHERE user_id = $2 AND category_id = $3 AND month = $4 AND year = $5
		`
		_, err := sqlTx.ExecContext(ctx, updateBudgetQuery, delta, trx.UserID(), trx.CategoryID(), month, year)
		if err != nil {
			r.logger.Error("Failed to update budget amount",
				zap.Error(err),
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("CategoryId", int(trx.CategoryID())),
				zap.Int("Month", month),
				zap.Int("Year", year),
				zap.Int64("Delta", delta))
			return fmt.Errorf("failed to update budget amount: %w", err)
		}
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
			r.logger.Error("Failed to update wallet amount",
				zap.Error(err),
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("WalletId", int(trx.WalletID())),
				zap.Int64("Delta", delta))
			return fmt.Errorf("failed to update wallet amount: %w", err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("Failed to check affected rows for wallet",
				zap.Error(err),
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("WalletId", int(trx.WalletID())))
			return fmt.Errorf("failed to check affected rows for wallet: %w", err)
		}
		if rowsAffected == 0 {
			r.logger.Warn("Wallet not found or does not belong to user",
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("WalletId", int(trx.WalletID())))
			return errors.New("wallet not found or does not belong to user")
		}
	}

	if err := sqlTx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction",
			zap.Error(err),
			zap.Int("UserId", int(trx.UserID())))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepo) Update(ctx context.Context, trx *transaction.Transaction, userId transaction.UserID) error {
	r.logger.Debug("Updating transaction",
		zap.Int("TransactionId", int(trx.ID())),
		zap.Int("UserId", int(userId)))

	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction for update",
			zap.Error(err),
			zap.Int("TransactionId", int(trx.ID())),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sqlTx.Rollback()

	getOldQuery := `SELECT user_id, goal_id, wallet_id, category_id, amount, transaction_type, transaction_date FROM transactions WHERE id = $1 AND user_id = $2 FOR UPDATE`
	var (
		oldUserID     uint64
		oldGoalIDNull sql.NullInt64
		oldWalletID   uint64
		oldCategoryID uint64
		oldAmount     int64
		oldType       string
		oldTrxDate    time.Time
	)
	err = sqlTx.QueryRowContext(ctx, getOldQuery, trx.ID(), userId).Scan(
		&oldUserID,
		&oldGoalIDNull,
		&oldWalletID,
		&oldCategoryID,
		&oldAmount,
		&oldType,
		&oldTrxDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Transaction not found for update",
				zap.Int("TransactionId", int(trx.ID())),
				zap.Int("UserId", int(userId)))
			return fmt.Errorf("transaction not found: %w", err)
		}
		r.logger.Error("Failed to fetch old transaction",
			zap.Error(err),
			zap.Int("TransactionId", int(trx.ID())),
			zap.Int("UserId", int(userId)))
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
			r.logger.Error("Failed to update saving goal amount during reversal",
				zap.Error(err),
				zap.Int("UserId", int(oldUserID)),
				zap.Int("GoalId", int(oldGoalIDNull.Int64)),
				zap.Int64("ReverseDelta", reverseDelta))
			return fmt.Errorf("failed to update saving goal amount during reversal: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("Failed to check affected rows for saving goal reversal",
				zap.Error(err),
				zap.Int("UserId", int(oldUserID)),
				zap.Int("GoalId", int(oldGoalIDNull.Int64)))
			return fmt.Errorf("failed to check affected rows for saving goal reversal: %w", err)
		}
		if rowsAffected == 0 {
			r.logger.Warn("Saving goal not found or does not belong to user during reversal",
				zap.Int("UserId", int(oldUserID)),
				zap.Int("GoalId", int(oldGoalIDNull.Int64)))
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
			r.logger.Error("Failed to update wallet balance during reversal",
				zap.Error(err),
				zap.Int("UserId", int(oldUserID)),
				zap.Int("WalletId", int(oldWalletID)),
				zap.Int64("ReverseDelta", reverseDelta))
			return fmt.Errorf("failed to update wallet balance during reversal: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("Failed to check affected rows for wallet reversal",
				zap.Error(err),
				zap.Int("UserId", int(oldUserID)),
				zap.Int("WalletId", int(oldWalletID)))
			return fmt.Errorf("failed to check affected rows for wallet reversal: %w", err)
		}
		if rowsAffected == 0 {
			r.logger.Warn("Wallet not found or does not belong to user during reversal",
				zap.Int("UserId", int(oldUserID)),
				zap.Int("WalletId", int(oldWalletID)))
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

		month := int(oldTrxDate.Month())
		year := oldTrxDate.Year()

		updateBudgetQuery := `
			UPDATE budgets
			SET amount = amount + $1, updated_at = NOW()
			WHERE user_id = $2 AND category_id = $3 AND month = $4 AND year = $5
		`
		_, err := sqlTx.ExecContext(ctx, updateBudgetQuery, reverseDelta, oldUserID, oldCategoryID, month, year)
		if err != nil {
			r.logger.Error("Failed to update budget amount during reversal",
				zap.Error(err),
				zap.Int("UserId", int(oldUserID)),
				zap.Int("CategoryId", int(oldCategoryID)),
				zap.Int("Month", month),
				zap.Int("Year", year),
				zap.Int64("ReverseDelta", reverseDelta))
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
		r.logger.Error("Failed to update transaction",
			zap.Error(err),
			zap.Int("TransactionId", int(trx.ID())),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to update transaction: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to check affected rows for transaction update",
			zap.Error(err),
			zap.Int("TransactionId", int(trx.ID())),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to check affected rows for transaction update: %w", err)
	}
	if rowsAffected == 0 {
		r.logger.Warn("Transaction not found or does not belong to user",
			zap.Int("TransactionId", int(trx.ID())),
			zap.Int("UserId", int(userId)))
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
			r.logger.Error("Failed to update saving goal amount",
				zap.Error(err),
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("GoalId", int(*trx.GoalID())),
				zap.Int64("Delta", delta))
			return fmt.Errorf("failed to update saving goal amount: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("Failed to check affected rows for saving goal",
				zap.Error(err),
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("GoalId", int(*trx.GoalID())))
			return fmt.Errorf("failed to check affected rows for saving goal: %w", err)
		}
		if rowsAffected == 0 {
			r.logger.Warn("Saving goal not found or does not belong to user",
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("GoalId", int(*trx.GoalID())))
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

		month := int(trx.TransactionDate().Month())
		year := trx.TransactionDate().Year()

		updateBudgetQuery := `
			UPDATE budgets
			SET amount = amount + $1, updated_at = NOW()
			WHERE user_id = $2 AND category_id = $3 AND month = $4 AND year = $5
		`
		_, err := sqlTx.ExecContext(ctx, updateBudgetQuery, delta, trx.UserID(), trx.CategoryID(), month, year)
		if err != nil {
			r.logger.Error("Failed to update budget amount",
				zap.Error(err),
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("CategoryId", int(trx.CategoryID())),
				zap.Int("Month", month),
				zap.Int("Year", year),
				zap.Int64("Delta", delta))
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
			r.logger.Error("Failed to update wallet amount",
				zap.Error(err),
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("WalletId", int(trx.WalletID())),
				zap.Int64("Delta", delta))
			return fmt.Errorf("failed to update wallet amount: %w", err)
		}

		rowsAffected, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("Failed to check affected rows for wallet",
				zap.Error(err),
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("WalletId", int(trx.WalletID())))
			return fmt.Errorf("failed to check affected rows for wallet: %w", err)
		}
		if rowsAffected == 0 {
			r.logger.Warn("Wallet not found or does not belong to user",
				zap.Int("UserId", int(trx.UserID())),
				zap.Int("WalletId", int(trx.WalletID())))
			return errors.New("wallet not found or does not belong to user")
		}
	}

	if err := sqlTx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction update",
			zap.Error(err),
			zap.Int("TransactionId", int(trx.ID())),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *TransactionRepo) Delete(ctx context.Context, trxID transaction.TransactionID, userId transaction.UserID) error {
	r.logger.Debug("Deleting transaction",
		zap.Int("TransactionId", int(trxID)),
		zap.Int("UserId", int(userId)))

	sqlTx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to begin transaction for delete",
			zap.Error(err),
			zap.Int("TransactionId", int(trxID)),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer sqlTx.Rollback()

	getOldQuery := `SELECT user_id, goal_id, wallet_id, category_id, amount, transaction_type, transaction_date FROM transactions WHERE id = $1 AND user_id = $2 FOR UPDATE`
	var (
		oldUserID     uint64
		oldGoalIDNull sql.NullInt64
		oldWalletID   uint64
		oldCategoryID uint64
		oldAmount     int64
		oldType       string
		oldTrxDate    time.Time
	)
	err = sqlTx.QueryRowContext(ctx, getOldQuery, trxID, userId).Scan(
		&oldUserID,
		&oldGoalIDNull,
		&oldWalletID,
		&oldCategoryID,
		&oldAmount,
		&oldType,
		&oldTrxDate,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Transaction not found for delete",
				zap.Int("TransactionId", int(trxID)),
				zap.Int("UserId", int(userId)))
			return fmt.Errorf("transaction not found: %w", err)
		}
		r.logger.Error("Failed to fetch transaction to delete",
			zap.Error(err),
			zap.Int("TransactionId", int(trxID)),
			zap.Int("UserId", int(userId)))
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
			r.logger.Error("Failed to update saving goal amount during deletion reversal",
				zap.Error(err),
				zap.Int("UserId", int(oldUserID)),
				zap.Int("GoalId", int(oldGoalIDNull.Int64)),
				zap.Int64("ReverseDelta", reverseDelta))
			return fmt.Errorf("failed to update saving goal amount: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("Failed to check affected rows for saving goal deletion reversal",
				zap.Error(err),
				zap.Int("UserId", int(oldUserID)),
				zap.Int("GoalId", int(oldGoalIDNull.Int64)))
			return fmt.Errorf("failed to check affected rows for saving goal: %w", err)
		}
		if rowsAffected == 0 {
			r.logger.Warn("Saving goal not found or does not belong to user during deletion reversal",
				zap.Int("UserId", int(oldUserID)),
				zap.Int("GoalId", int(oldGoalIDNull.Int64)))
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
			r.logger.Error("Failed to update wallet balance during deletion reversal",
				zap.Error(err),
				zap.Int("UserId", int(oldUserID)),
				zap.Int("WalletId", int(oldWalletID)),
				zap.Int64("ReverseDelta", reverseDelta))
			return fmt.Errorf("failed to update wallet balance: %w", err)
		}
		rowsAffected, err := res.RowsAffected()
		if err != nil {
			r.logger.Error("Failed to check affected rows for wallet deletion reversal",
				zap.Error(err),
				zap.Int("UserId", int(oldUserID)),
				zap.Int("WalletId", int(oldWalletID)))
			return fmt.Errorf("failed to check affected rows for wallet: %w", err)
		}
		if rowsAffected == 0 {
			r.logger.Warn("Wallet not found or does not belong to user during deletion reversal",
				zap.Int("UserId", int(oldUserID)),
				zap.Int("WalletId", int(oldWalletID)))
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

		month := int(oldTrxDate.Month())
		year := oldTrxDate.Year()

		updateBudgetQuery := `
			UPDATE budgets
			SET amount = amount + $1, updated_at = NOW()
			WHERE user_id = $2 AND category_id = $3 AND month = $4 AND year = $5
		`
		_, err := sqlTx.ExecContext(ctx, updateBudgetQuery, reverseDelta, oldUserID, oldCategoryID, month, year)
		if err != nil {
			r.logger.Error("Failed to update budget amount during deletion reversal",
				zap.Error(err),
				zap.Int("UserId", int(oldUserID)),
				zap.Int("CategoryId", int(oldCategoryID)),
				zap.Int("Month", month),
				zap.Int("Year", year),
				zap.Int64("ReverseDelta", reverseDelta))
			return fmt.Errorf("failed to update budget amount: %w", err)
		}
	}

	// Delete transaction
	deleteQuery := `DELETE FROM transactions WHERE id = $1 AND user_id = $2`
	res, err := sqlTx.ExecContext(ctx, deleteQuery, trxID, userId)
	if err != nil {
		r.logger.Error("Failed to delete transaction",
			zap.Error(err),
			zap.Int("TransactionId", int(trxID)),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to delete transaction: %w", err)
	}
	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to check affected rows for transaction deletion",
			zap.Error(err),
			zap.Int("TransactionId", int(trxID)),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to check affected rows for transaction deletion: %w", err)
	}
	if rowsAffected == 0 {
		r.logger.Warn("Transaction not found or does not belong to user for deletion",
			zap.Int("TransactionId", int(trxID)),
			zap.Int("UserId", int(userId)))
		return errors.New("transaction not found or does not belong to user")
	}

	if err := sqlTx.Commit(); err != nil {
		r.logger.Error("Failed to commit transaction deletion",
			zap.Error(err),
			zap.Int("TransactionId", int(trxID)),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}
