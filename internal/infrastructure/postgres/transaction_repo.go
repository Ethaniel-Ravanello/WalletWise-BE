package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
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

func (t *TransactionRepo) Search(ctx context.Context, filter transaction.FilterTrx) ([]*transaction.Transaction, error) {
	var counter = 1
	q := fmt.Sprintf("SELECT id, user_id, goal_id, category_id, amount, description, transaction_type, wallet_id, transaction_date, created_at, updated_at FROM transactions WHERE user_id = $%d", counter)
	args := []interface{}{filter.UserID}
	counter++

	if filter.GoalID != nil {
		q += fmt.Sprintf(" AND goal_id = $%d", counter)
		args = append(args, *filter.GoalID)
		counter++
	}
	if filter.Amount != nil {
		q += fmt.Sprintf(" AND amount = $%d", counter)
		args = append(args, *filter.Amount)
		counter++
	}
	if filter.CategoryId != nil {
		q += fmt.Sprintf(" AND category_id  = $%d", counter)
		args = append(args, *filter.CategoryId)
		counter++
	}
	if filter.TransactionTypes != nil {
		q += fmt.Sprintf(" AND transaction_type = $%d", counter)
		args = append(args, *filter.TransactionTypes)
		counter++
	}
	if filter.StartDate != nil {
		q += fmt.Sprintf(" AND transaction_date >= $%d", counter)
		args = append(args, *filter.StartDate)
		counter++
	}
	if filter.EndDate != nil {
		q += fmt.Sprintf(" AND transaction_date <= $%d", counter)
		args = append(args, *filter.EndDate)
		counter++
	}
	if filter.WalletID != nil {
		q += fmt.Sprintf(" AND wallet_id = $%d", counter)
	}
	q += " ORDER BY transaction_date DESC"
	if filter.Limit > 0 {
		q += fmt.Sprintf(" LIMIT $%d", counter)
		args = append(args, filter.Limit)
		counter++
	}

	rows, err := t.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, errors.New("Error Getting Data: " + err.Error())
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)
	var transactions []*transaction.Transaction

	var (
		id, userID                            uint64
		goalIDNull                            sql.NullInt64
		amount                                int64
		categoryID                            uint64
		description, transactionType          string
		walletID                              uint64
		transactionDate, createdAt, updatedAt time.Time
	)

	for rows.Next() {
		if err = rows.Scan(
			&id,
			&userID,
			&goalIDNull,
			&amount,
			&categoryID,
			&description,
			&transactionType,
			&walletID,
			&transactionDate,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, errors.New("Error Getting Data: " + err.Error())
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
		return nil, errors.New("Error Getting Data: " + err.Error())
	}
	return transactions, nil
}

func (t *TransactionRepo) SearchByID(ctx context.Context, trxId transaction.TransactionID) (*transaction.Transaction, error) {
	const q = `SELECT id, user_id, goal_id, category_id, amount, description, transaction_type, wallet_id, transaction_date, created_at, updated_at FROM transactions
				WHERE id = $1`

	var (
		transactionID, userID                 uint64
		goalIDNull                            sql.NullInt64
		amount                                int64
		categoryID                            uint64
		description, transactionType          string
		walletID                              uint64
		transactionDate, createdAt, updatedAt time.Time
	)

	row := t.db.QueryRowContext(ctx, q, trxId)

	err := row.Scan(
		&transactionID,
		&userID,
		&goalIDNull,
		&amount,
		&categoryID,
		&description,
		&transactionType,
		&walletID,
		&transactionDate,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Transaction not found")
		}
	}
	var finalGoalID *transaction.GoalID
	if goalIDNull.Valid {
		val := transaction.GoalID(goalIDNull.Int64)
		finalGoalID = &val
	}
	trx := transaction.Reconstitute(
		transaction.TransactionID(transactionID),
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

func (t *TransactionRepo) GetBalance(ctx context.Context, userId transaction.UserID, walletId transaction.WalletID) (transaction.Money, error) {
	const q = `SELECT COALESCE(SUM(CASE WHEN type = 'income' then amount
    								WHEN type = 'expense' then -amount), 0) 
									FROM transactions
									WHERE user_id = $1 AND WHERE wallet_id = $2`

	var amount transaction.Money
	err := t.db.QueryRowContext(ctx, q, userId, walletId).Scan(&amount)
	if err != nil {
		return 0, errors.New("Error Getting Balance: " + err.Error())
	}
	return amount, nil
}

func (t *TransactionRepo) GetMonthlySummary(ctx context.Context, userId transaction.UserID, month int, year int) (transaction.MonthlySummary, error) {
	const netFlow = `SELECT
						COALESCE(SUM(CASE WHEN type = 'income' then amount ELSE 0 END), 0),
						COALESCE(SUM(CASE WHEN type = 'expense' then amount ELSE 0 END), 0)
					 FROM transactions
					 WHERE user_id = $1 AND MONTH(date) = $2 AND YEAR(date) = $3`
	var totalIncome, totalExpense transaction.Money

	err := t.db.QueryRowContext(ctx, netFlow, userId, month, year).Scan(&totalIncome, &totalExpense)
	if err != nil {
		return transaction.MonthlySummary{}, errors.New("Error Getting Monthly Summary: " + err.Error())
	}
	return transaction.MonthlySummary{
		TotalIncome:  totalIncome,
		TotalExpense: totalExpense,
	}, nil
}

func (t *TransactionRepo) GetHighestExpense(ctx context.Context, userId transaction.UserID, month int, year int, limit int) (*transaction.Transaction, error) {
	const trxExpense = `SELECT id, user_id, goal_id, amount, category_id, description, transaction_type, wallet_id, transaction_date, created_at, updated_at
						FROM transactions
						WHERE user_id = $1 AND MONTH(date) = $2 AND YEAR(date) = $3 AND types = 'expense'
						ORDER BY amount DESC
						LIMIT $4`

	var (
		transactionID, userID                 uint64
		goalIDNull                            sql.NullInt64
		amount                                int64
		categoryID                            uint64
		description, transactionType          string
		walletID                              uint64
		transactionDate, createdAt, updatedAt time.Time
	)
	row := t.db.QueryRowContext(ctx, trxExpense, userId, month, year, limit)
	err := row.Scan(
		&transactionID,
		&userID,
		&goalIDNull,
		&amount,
		&categoryID,
		&description,
		&transactionType,
		&walletID,
		&transactionDate,
		&createdAt,
		&updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Transaction not found")
		}
	}
	var finalGoalID *transaction.GoalID
	if goalIDNull.Valid {
		val := transaction.GoalID(goalIDNull.Int64)
		finalGoalID = &val
	}

	trx := transaction.Reconstitute(
		transaction.TransactionID(transactionID),
		transaction.UserID(userID),
		finalGoalID,
		transaction.Money(amount),
		transaction.CategoryID(categoryID),
		description,
		transaction.TransactionType(transactionType),
		transaction.WalletID(walletID),
		transactionDate,
		createdAt,
		updatedAt)
	return trx, nil
}

func (t *TransactionRepo) GetMostSpend(ctx context.Context, userId transaction.UserID, month int, year int, limit int) ([]*transaction.CategorySpend, error) {
	const q = `SELECT category, SUM(amount)
				FROM transactions
				WHERE user_id = $1 AND YEAR(date) = $2 AND MONTH(date) = $3 
				GROUP BY category
				ORDER BY total DESC
				LIMIT $4`

	rows, err := t.db.QueryContext(ctx, q, userId, year, month, limit)
	if err != nil {
		return nil, errors.New("Error Getting Most Spend: " + err.Error())
	}
	defer rows.Close()
	var trxs []*transaction.CategorySpend
	for rows.Next() {
		var cs transaction.CategorySpend
		err := rows.Scan(&cs.Category, &cs.Total)
		if err != nil {
			return nil, errors.New("Error Getting Most Spend: " + err.Error())
		}
		trxs = append(trxs, &cs)
	}
	return trxs, nil
}

func (t *TransactionRepo) Save(ctx context.Context, tx *transaction.Transaction) error {
	const q = `INSERT INTO transactions(user_id, goal_id, amount, category_id, description, transaction_type, wallet_id, transaction_date, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`

	_, err := t.db.ExecContext(ctx, q,
		tx.UserID(),
		tx.GoalID(),
		tx.Amount(),
		tx.CategoryID(),
		tx.Description(),
		tx.TransactionType(),
		tx.WalletID(),
		tx.TransactionDate(),
		time.Now(),
		time.Now(),
	)
	return err
}

func (t *TransactionRepo) Update(ctx context.Context, tx *transaction.Transaction) error {
	const q = `UPDATE transactions SET userId=$1, goalId=$2, amount=$3, category=$4, description=$5, transaction_type=$6, wallet_id=$7, transaction_date=$8, updated_at=$9 WHERE id = $10`

	_, err := t.db.ExecContext(ctx, q,
		tx.UserID(),
		tx.GoalID(),
		tx.Amount(),
		tx.CategoryID(),
		tx.Description(),
		tx.TransactionType(),
		tx.WalletID(),
		tx.TransactionDate(),
		time.Now(),
		tx.ID(),
	)
	return err
}

func (t *TransactionRepo) Delete(ctx context.Context, trxId transaction.TransactionID) error {
	const q = `DELETE FROM transactions WHERE id = $1`

	_, err := t.db.ExecContext(ctx, q, trxId)
	return err
}
