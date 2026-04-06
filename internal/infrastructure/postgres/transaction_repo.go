package postgres

import (
	"context"
	"database/sql"
	"errors"
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
	q := `SELECT id,user_id, goal_id, amount, category, description, transaction_type, transaction_source, transaction_date, created_at, updated_at FROM transactions
				WHERE user_id = ?`
	args := []interface{}{filter.UserID}
	if filter.GoalID != nil {
		q += " AND goal_id = ?"
		args = append(args, *filter.GoalID)
	}
	if filter.Amount != nil {
		q += " AND amount = ?"
		args = append(args, *filter.Amount)
	}
	if filter.Category != nil {
		q += " AND category = ?"
		args = append(args, *filter.Category)
	}
	if filter.StartDate != nil {
		q += " AND date >= ?"
		args = append(args, *filter.StartDate)
	}
	if filter.EndDate != nil {
		q += " AND date <= ?"
		args = append(args, *filter.EndDate)
	}
	q += " ORDER BY transaction_date DESC"
	if filter.Limit > 0 {
		q += " LIMIT ?"
		args = append(args, filter.Limit)
	}

	rows, err := t.db.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, errors.New("Error Getting Data: " + err.Error())
	}
	defer rows.Close()
	var transactions []*transaction.Transaction

	var (
		id                 transaction.TransactionId
		user_id, goal_id   uint64
		amount             int64
		category           string
		description        string
		transaction_type   string
		transaction_source string
		transaction_date   time.Time
		created_at         time.Time
		updated_at         time.Time
	)

	for rows.Next() {
		if err = rows.Scan(
			&id,
			&user_id,
			&goal_id,
			&amount,
			&category,
			&description,
			&transaction_type,
			&transaction_source,
			&transaction_date,
			&created_at,
			&updated_at,
		); err != nil {
			return nil, errors.New("Error Getting Data: " + err.Error())
		}
		trx := transaction.Reconstitute(
			id,
			user_id,
			goal_id,
			transaction.Money(amount),
			category,
			description,
			transaction.Type(transaction_type),
			transaction_source,
			transaction_date,
			created_at,
			updated_at,
		)
		transactions = append(transactions, trx)
	}
	if err := rows.Err(); err != nil {
		return nil, errors.New("Error Getting Data: " + err.Error())
	}
	return transactions, nil
}

func (t *TransactionRepo) SearchByID(ctx context.Context, trxId transaction.TransactionId) (*transaction.Transaction, error) {
	const q = `SELECT id,user_id, goal_id, amount, category, description, transaction_type, transaction_source, transaction_date, created_at, updated_at FROM transactions
				WHERE id = ?`

	var (
		trxID                                  transaction.TransactionId
		userID, goalID                         uint64
		amount                                 int64
		category, desc                         string
		transaction_type                       string
		transaction_source                     string
		transaction_date, createdAt, updatedAt time.Time
	)

	row := t.db.QueryRowContext(ctx, q, trxId)

	err := row.Scan(
		&trxID,
		&userID,
		&goalID,
		&amount,
		&category,
		&desc,
		&transaction_type,
		&transaction_source,
		&transaction_date,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Transaction not found")
		}
	}
	trx := transaction.Reconstitute(
		trxID,
		userID,
		goalID,
		transaction.Money(amount),
		category,
		desc,
		transaction.Type(transaction_type),
		transaction_source,
		transaction_date,
		createdAt,
		updatedAt,
	)
	return trx, nil
}

func (t *TransactionRepo) GetBalance(ctx context.Context, userId uint64) (transaction.Money, error) {
	const q = `SELECT COALESCE(SUM(CASE WHEN type = 'DEBIT' then amount
    								WHEN type = 'CREDIT' then -amount), 0) 
									FROM transactions
									WHERE user_id = ?`

	var amount int64
	err := t.db.QueryRowContext(ctx, q, userId).Scan(&amount)
	if err != nil {
		return 0, errors.New("Error Getting Balance: " + err.Error())
	}
	return transaction.Money(amount), nil
}

func (t *TransactionRepo) GetMonthlySummary(ctx context.Context, userId uint64, month int, year int) (transaction.MonthlySummary, error) {
	const netFlow = `SELECT
						COALESCE(SUM(CASE WHEN type = 'DEBIT' then amount ELSE 0 END), 0),
						COALESCE(SUM(CASE WHEN type = 'CREDIT' then amount ELSE 0 END), 0)
					 FROM transactions
					 WHERE user_id = ? AND MONTH(date) = ? AND YEAR(date) = ?`
	var totalIncome, totalExpense int64

	err := t.db.QueryRowContext(ctx, netFlow, userId, month, year).Scan(&totalIncome, &totalExpense)
	if err != nil {
		return transaction.MonthlySummary{}, errors.New("Error Getting Monthly Summary: " + err.Error())
	}
	return transaction.MonthlySummary{
		TotalIncome:  transaction.Money(totalIncome),
		TotalExpense: transaction.Money(totalExpense),
	}, nil
}

func (t *TransactionRepo) GetHighestExpense(ctx context.Context, userId uint64, month int, year int, limit int) (*transaction.Transaction, error) {
	const trxExpense = `SELECT id, user_id, goal_id, amount, category, description, transaction_type, transaction_source, transaction_date, created_at, updated_at
						FROM transactions
						WHERE user_id = ? AND MONTH(date) = ? AND YEAR(date) = ? AND types = 'DEBIT'
						ORDER BY amount DESC
						LIMIT ?`

	var (
		trxID                                  transaction.TransactionId
		userID, goalID                         uint64
		amount                                 int64
		category, desc                         string
		transaction_type                       string
		transaction_source                     string
		transaction_date, createdAt, updatedAt time.Time
	)
	row := t.db.QueryRowContext(ctx, trxExpense, userId, month, year, limit)
	err := row.Scan(
		&trxID,
		&userID,
		&goalID,
		&amount,
		&category,
		&desc,
		&transaction_type,
		&transaction_source,
		&transaction_date,
		&createdAt,
		&updatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("Transaction not found")
		}
	}
	trx := transaction.Reconstitute(
		trxID,
		userID,
		goalID,
		transaction.Money(amount),
		category,
		desc,
		transaction.Type(transaction_type),
		transaction_source,
		transaction_date,
		createdAt,
		updatedAt)
	return trx, nil
}

func (t *TransactionRepo) GetMostSpend(ctx context.Context, userId uint64, month int, year int, limit int) ([]*transaction.CategorySpend, error) {
	const q = `SELECT category, SUM(amount)
				FROM transactions
				WHERE user_id = ? AND YEAR(date) = ? AND MONTH(date) = ? 
				GROUP BY category
				ORDER BY total DESC
				LIMIT ?`

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
	const q = `INSERT INTO transactions(user_id, goal_id, amount, category, description, transaction_type, transaction_source, transaction_date, created_at, updated_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	_, err := t.db.ExecContext(ctx, q,
		tx.UserID(),
		tx.GoalID(),
		tx.Amount(),
		tx.Category(),
		tx.Description(),
		tx.TransactionType(),
		tx.TransactionSource(),
		tx.TransactionDate(),
		time.Now(),
		time.Now(),
	)
	return err
}

func (t *TransactionRepo) Update(ctx context.Context, tx *transaction.Transaction) error {
	const q = `UPDATE transactions SET amount=?, category=?, description=?, transaction_type=?, transaction_source=?, transaction_date=?, updated_at=? WHERE id = ?`

	_, err := t.db.ExecContext(ctx, q,
		tx.Amount(),
		tx.Category(),
		tx.Description(),
		tx.TransactionType(),
		tx.TransactionSource(),
		tx.TransactionDate(),
		time.Now(),
		tx.ID(),
	)
	return err
}

func (t *TransactionRepo) Delete(ctx context.Context, trxId transaction.TransactionId) error {
	const q = `DELETE FROM transactions WHERE id = ?`

	_, err := t.db.ExecContext(ctx, q, trxId)
	return err
}
