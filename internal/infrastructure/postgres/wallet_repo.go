package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"walletwise/internal/domain/wallet"
)

type WalletRepo struct {
	db *sql.DB
}

func NewWalletRepo(db *sql.DB) *WalletRepo { return &WalletRepo{db: db} }

var _ wallet.Repository = (*WalletRepo)(nil)

func (w WalletRepo) SearchAll(ctx context.Context, userID wallet.UserID) ([]*wallet.Wallet, error) {
	const sql = `SELECT id, user_id, name, type, created_at, updated_at FROM wallets WHERE user_id = $1`

	var wallets []*wallet.Wallet

	var (
		id         uint64
		tempUserID uint64
		name       string
		walletType string
		created_at time.Time
		updated_at time.Time
	)

	rows, err := w.db.QueryContext(ctx, sql, userID)
	if err != nil {
		return nil, errors.New("Error Scanning Rows: " + err.Error())
	}
	defer rows.Close()
	for rows.Next() {
		if err = rows.Scan(
			&id,
			&tempUserID,
			name,
			walletType,
			created_at,
			updated_at,
		); err != nil {
			return nil, errors.New("Internal Error: " + err.Error())
		}
	}
	fixWallet := wallet.ReconstituteWallet(
		wallet.WalletID(id),
		wallet.UserID(tempUserID),
		name,
		walletType,
		created_at,
		updated_at)
	wallets = append(wallets, fixWallet)
	return wallets, nil
}

func (w WalletRepo) SearchByID(ctx context.Context, userID wallet.UserID, walletID wallet.WalletID) (*wallet.Wallet, error) {
	const sql = `SELECT id, user_id, name, type, created_at, updated_at FROM wallets WHERE id = $1`

	var (
		id         uint64
		tempUserID uint64
		name       string
		walletType string
		created_at time.Time
		updated_at time.Time
	)
	rows, err := w.db.QueryContext(ctx, sql, userID)
	if err != nil {
		return nil, errors.New("Error Scanning Rows: " + err.Error())
	}
	defer rows.Close()
	err = rows.Scan(
		&id,
		&tempUserID,
		&name,
		&walletType,
		&created_at,
		&updated_at)
	if err != nil {
		return nil, errors.New("Internal Error: " + err.Error())
	}
	wlt := wallet.ReconstituteWallet(
		wallet.WalletID(id),
		wallet.UserID(tempUserID),
		name,
		walletType,
		created_at,
		updated_at)
	return wlt, nil
}

func (w WalletRepo) Save(ctx context.Context, wallet *wallet.Wallet) error {
	const sql = `INSERT INTO wallets (user_id, name, type, created_at, updated_at) VALUES ($1, $2, $3, $4, $5)`

	_, err := w.db.ExecContext(ctx, sql,
		wallet.UserID(),
		wallet.Name(),
		wallet.WalletType(),
		time.Now(),
		time.Now())
	if err != nil {
		return errors.New("Internal Error: " + err.Error())
	}
	return nil
}

func (w WalletRepo) Update(ctx context.Context, wallet *wallet.Wallet) error {
	const sql = `UPDATE wallets SET name = $1, wallet_type = $2, updated_at = $3 WHERE user_id = $3`

	_, err := w.db.ExecContext(ctx, sql,
		wallet.Name(),
		wallet.WalletType(),
		time.Now())
	if err != nil {
		return errors.New("Internal Error: " + err.Error())
	}
	return nil
}

func (w WalletRepo) Delete(ctx context.Context, wallet *wallet.Wallet) error {
	const sql = `DELETE FROM wallets WHERE id = $1`

	_, err := w.db.ExecContext(ctx, sql, wallet.ID())

	if err != nil {
		return errors.New("Internal Error: " + err.Error())
	}
	return nil
}

func (w WalletRepo) SearchHighestBalance(ctx context.Context, userID wallet.UserID) (*wallet.Wallet, error) {
	const sql = `
	SELECT
	w.id,
		w.name,
		w.type,
	w.created_at,
		w.updated_at,
	-- Ini rumus ajaibnya: Jika income ditambah, jika expense dikurang
	COALESCE(SUM(
		CASE
	WHEN t.transaction_type = 'income' THEN t.amount
	WHEN t.transaction_type = 'expense' THEN -t.amount
	ELSE 0
	END
	), 0) AS calculated_balance
	FROM wallets w
	LEFT JOIN transactions t ON w.id = t.wallet_id
	WHERE w.user_id = $1
	GROUP BY w.id
	ORDER BY calculated_balance DESC
	LIMIT 1
	`

	var (
		dbId              uint64
		dbUserId          uint64
		dbName            string
		dbTransactionType string
		dbCreated         time.Time
		dbUpdated         time.Time
	)
	err := w.db.QueryRowContext(ctx, sql, userID).Scan(
		&dbId,
		&dbUserId,
		&dbName,
		&dbTransactionType,
		&dbCreated,
		&dbUpdated)
	if err != nil {
		return nil, errors.New("Error Scanning Rows: " + err.Error())
	}
	walletEntity := wallet.ReconstituteWallet(
		wallet.WalletID(dbId),
		wallet.UserID(dbUserId),
		dbName,
		dbTransactionType,
		dbCreated,
		dbUpdated,
	)
	return walletEntity, nil
}

func (w WalletRepo) SearchMostActive(ctx context.Context, userID wallet.UserID) (*wallet.Wallet, error) {
	const sql = `SELECT 
            w.id, 
            w.name, 
            w.type, 
            w.created_at, 
            w.updated_at,
            COUNT(t.id) AS total_transactions -- Menghitung berapa kali dompet ini dipakai
        FROM wallets w
        LEFT JOIN transactions t ON w.id = t.wallet_id
        WHERE w.user_id = $1
        GROUP BY w.id
        ORDER BY total_transactions DESC -- Urutkan dari jumlah transaksi terbanyak
        LIMIT 1`
	return nil, errors.New("Test")
}

func (w WalletRepo) SearchTotalBalance(ctx context.Context, userID wallet.UserID) (*wallet.Wallet, error) {
	//TODO implement me
	panic("implement me")
}
