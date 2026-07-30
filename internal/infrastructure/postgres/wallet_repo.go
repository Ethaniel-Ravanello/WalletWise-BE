package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"walletwise/internal/domain/wallet"
)

type WalletRepo struct {
	db *sql.DB
}

func NewWalletRepo(db *sql.DB) *WalletRepo {
	return &WalletRepo{db: db}
}

var _ wallet.Repository = (*WalletRepo)(nil)

func (r *WalletRepo) SearchAll(ctx context.Context, userID wallet.UserID) ([]*wallet.Wallet, error) {
	query := `SELECT id, user_id, name, wallet_type, balance, created_at, updated_at FROM wallets WHERE user_id = $1`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query wallets: %w", err)
	}
	defer rows.Close()

	var wallets []*wallet.Wallet
	for rows.Next() {
		var (
			id         uint64
			dbUserID   uint64
			name       string
			walletType string
			balance    uint64
			createdAt  time.Time
			updatedAt  time.Time
		)

		if err := rows.Scan(
			&id,
			&dbUserID,
			&name,
			&walletType,
			&balance,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan wallet row: %w", err)
		}

		wlt := wallet.ReconstituteWallet(
			wallet.WalletID(id),
			wallet.UserID(dbUserID),
			name,
			walletType,
			wallet.Balance(balance),
			createdAt,
			updatedAt,
		)
		wallets = append(wallets, wlt)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating wallet rows: %w", err)
	}

	return wallets, nil
}

func (r *WalletRepo) SearchByID(ctx context.Context, walletID wallet.WalletID, userId wallet.UserID) (*wallet.Wallet, error) {
	query := `SELECT id, user_id, name, wallet_type, balance, created_at, updated_at FROM wallets WHERE id = $1 AND user_id = $2`

	var (
		id         uint64
		dbUserID   uint64
		name       string
		walletType string
		balance    uint64
		createdAt  time.Time
		updatedAt  time.Time
	)

	row := r.db.QueryRowContext(ctx, query, walletID, userId)
	err := row.Scan(
		&id,
		&dbUserID,
		&name,
		&walletType,
		&balance,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("wallet not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find wallet by id: %w", err)
	}

	wlt := wallet.ReconstituteWallet(
		wallet.WalletID(id),
		wallet.UserID(dbUserID),
		name,
		walletType,
		wallet.Balance(balance),
		createdAt,
		updatedAt,
	)
	return wlt, nil
}

func (r *WalletRepo) Save(ctx context.Context, wlt *wallet.Wallet) error {
	query := `INSERT INTO wallets (user_id, name, wallet_type, balance, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		wlt.UserID(),
		wlt.Name(),
		wlt.WalletType(),
		wlt.Balance(),
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("failed to save wallet: %w", err)
	}
	return nil
}

func (r *WalletRepo) Update(ctx context.Context, wlt *wallet.Wallet, userId wallet.UserID) error {
	query := `UPDATE wallets SET name = $1, wallet_type = $2, balance = $3, updated_at = $4 WHERE id = $5 AND user_id = $6`

	_, err := r.db.ExecContext(ctx, query,
		wlt.Name(),
		wlt.WalletType(),
		wlt.Balance(),
		time.Now(),
		wlt.ID(),
		userId,
	)
	if err != nil {
		return fmt.Errorf("failed to update wallet: %w", err)
	}
	return nil
}

func (r *WalletRepo) Delete(ctx context.Context, wlt *wallet.Wallet, userId wallet.UserID) error {
	query := `DELETE FROM wallets WHERE id = $1 AND user_id = $2`

	_, err := r.db.ExecContext(ctx, query, wlt.ID(), userId)
	if err != nil {
		return fmt.Errorf("failed to delete wallet: %w", err)
	}
	return nil
}

func (r *WalletRepo) SearchHighestBalance(ctx context.Context, userID wallet.UserID) (*wallet.Wallet, error) {
	query := `
        SELECT id, user_id, name, wallet_type, balance, created_at, updated_at 
        FROM wallets 
        WHERE user_id = $1 
        ORDER BY balance DESC 
        LIMIT 1
    `

	var (
		id         uint64
		dbUserID   uint64
		name       string
		walletType string
		balance    uint64
		createdAt  time.Time
		updatedAt  time.Time
	)

	row := r.db.QueryRowContext(ctx, query, userID)
	err := row.Scan(
		&id,
		&dbUserID,
		&name,
		&walletType,
		&balance,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("wallet not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find highest balance wallet: %w", err)
	}

	wlt := wallet.ReconstituteWallet(
		wallet.WalletID(id),
		wallet.UserID(dbUserID),
		name,
		walletType,
		wallet.Balance(balance),
		createdAt,
		updatedAt,
	)
	return wlt, nil
}

func (r *WalletRepo) SearchMostActive(ctx context.Context, userID wallet.UserID) (*wallet.Wallet, error) {
	query := `
        SELECT 
            w.id, 
            w.user_id,
            w.name, 
            w.wallet_type, 
            w.balance, 
            w.created_at, 
            w.updated_at,
            COUNT(t.id) AS tx_count
        FROM wallets w
        LEFT JOIN transactions t ON w.id = t.wallet_id
        WHERE w.user_id = $1
        GROUP BY w.id, w.user_id, w.name, w.wallet_type, w.balance, w.created_at, w.updated_at
        ORDER BY tx_count DESC
        LIMIT 1
    `

	var (
		id         uint64
		dbUserID   uint64
		name       string
		walletType string
		balance    uint64
		createdAt  time.Time
		updatedAt  time.Time
		txCount    int64
	)

	row := r.db.QueryRowContext(ctx, query, userID)
	err := row.Scan(
		&id,
		&dbUserID,
		&name,
		&walletType,
		&balance,
		&createdAt,
		&updatedAt,
		&txCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("wallet not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find most active wallet: %w", err)
	}

	wlt := wallet.ReconstituteWallet(
		wallet.WalletID(id),
		wallet.UserID(dbUserID),
		name,
		walletType,
		wallet.Balance(balance),
		createdAt,
		updatedAt,
	)
	return wlt, nil
}

func (r *WalletRepo) SearchTotalBalance(ctx context.Context, userID wallet.UserID) (uint64, error) {
	query := `SELECT COALESCE(SUM(balance), 0) 
	          FROM wallets 
	          WHERE user_id = $1`

	var balance uint64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("failed to calculate total balance: %w", err)
	}
	return balance, nil
}
