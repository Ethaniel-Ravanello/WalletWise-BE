package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"walletwise/internal/domain/wallet"
)

type WalletRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewWalletRepo(db *sql.DB) *WalletRepo {
	return &WalletRepo{
		db:     db,
		logger: zap.L(),
	}
}

var _ wallet.Repository = (*WalletRepo)(nil)

func (r *WalletRepo) SearchAll(ctx context.Context, userID wallet.UserID) ([]*wallet.Wallet, error) {
	query := `SELECT id, user_id, wallet_name, wallet_type, balance, created_at, updated_at FROM wallets WHERE user_id = $1`
	r.logger.Debug("SearchAll wallets query", zap.String("Query", query))

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		r.logger.Error("Failed to query wallets",
			zap.Error(err),
			zap.Int("UserId", int(userID)))
		return nil, fmt.Errorf("failed to query wallets: %w", err)
	}
	defer rows.Close()

	var wallets []*wallet.Wallet
	for rows.Next() {
		var (
			id         uint64
			dbUserID   uint64
			walletName string
			walletType string
			balance    uint64
			createdAt  time.Time
			updatedAt  time.Time
		)

		if err := rows.Scan(
			&id,
			&dbUserID,
			&walletName,
			&walletType,
			&balance,
			&createdAt,
			&updatedAt,
		); err != nil {
			r.logger.Error("Failed to scan wallet row",
				zap.Error(err),
				zap.Int("UserId", int(userID)))
			return nil, fmt.Errorf("failed to scan wallet row: %w", err)
		}

		wlt := wallet.ReconstituteWallet(
			wallet.ID(id),
			wallet.UserID(dbUserID),
			walletName,
			walletType,
			wallet.Balance(balance),
			createdAt,
			updatedAt,
		)
		wallets = append(wallets, wlt)
	}

	if err := rows.Err(); err != nil {
		r.logger.Error("Error iterating wallet rows",
			zap.Error(err),
			zap.Int("UserId", int(userID)))
		return nil, fmt.Errorf("error iterating wallet rows: %w", err)
	}

	return wallets, nil
}

func (r *WalletRepo) SearchByID(ctx context.Context, walletID wallet.ID, userId wallet.UserID) (*wallet.Wallet, error) {
	query := `SELECT id, user_id, wallet_name, wallet_type, balance, created_at, updated_at FROM wallets WHERE id = $1 AND user_id = $2`
	r.logger.Debug("SearchByID wallet query", zap.String("Query", query))

	var (
		id         uint64
		dbUserID   uint64
		walletName string
		walletType string
		balance    uint64
		createdAt  time.Time
		updatedAt  time.Time
	)

	row := r.db.QueryRowContext(ctx, query, walletID, userId)
	err := row.Scan(
		&id,
		&dbUserID,
		&walletName,
		&walletType,
		&balance,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Wallet not found",
				zap.Int("WalletId", int(walletID)),
				zap.Int("UserId", int(userId)))
			return nil, fmt.Errorf("wallet not found: %w", err)
		}
		r.logger.Error("Failed to find wallet by id",
			zap.Error(err),
			zap.Int("WalletId", int(walletID)),
			zap.Int("UserId", int(userId)))
		return nil, fmt.Errorf("failed to find wallet by id: %w", err)
	}

	wlt := wallet.ReconstituteWallet(
		wallet.ID(id),
		wallet.UserID(dbUserID),
		walletName,
		walletType,
		wallet.Balance(balance),
		createdAt,
		updatedAt,
	)
	return wlt, nil
}

func (r *WalletRepo) Save(ctx context.Context, wlt *wallet.Wallet) error {
	r.logger.Debug("Saving wallet",
		zap.Int("UserId", int(wlt.UserID())),
		zap.String("WalletName", wlt.Name()))

	query := `INSERT INTO wallets (user_id, wallet_name, wallet_type, balance, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(ctx, query,
		wlt.UserID(),
		wlt.Name(),
		wlt.WalletType(),
		wlt.Balance(),
		time.Now(),
		time.Now(),
	)
	if err != nil {
		r.logger.Error("Failed to save wallet",
			zap.Error(err),
			zap.Int("UserId", int(wlt.UserID())),
			zap.String("WalletName", wlt.Name()))
		return fmt.Errorf("failed to save wallet: %w", err)
	}
	return nil
}

func (r *WalletRepo) Update(ctx context.Context, wlt *wallet.Wallet, userId wallet.UserID) error {
	r.logger.Debug("Updating wallet",
		zap.Int("WalletId", int(wlt.ID())),
		zap.Int("UserId", int(userId)))

	query := `UPDATE wallets SET wallet_name = $1, wallet_type = $2, balance = $3, updated_at = $4 WHERE id = $5 AND user_id = $6`

	res, err := r.db.ExecContext(ctx, query,
		wlt.Name(),
		wlt.WalletType(),
		wlt.Balance(),
		time.Now(),
		wlt.ID(),
		userId,
	)
	if err != nil {
		r.logger.Error("Failed to update wallet",
			zap.Error(err),
			zap.Int("WalletId", int(wlt.ID())),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to update wallet: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to check affected rows for wallet update",
			zap.Error(err),
			zap.Int("WalletId", int(wlt.ID())),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to check affected rows for wallet update: %w", err)
	}
	if rowsAffected == 0 {
		r.logger.Warn("Wallet not found for update",
			zap.Int("WalletId", int(wlt.ID())),
			zap.Int("UserId", int(userId)))
		return errors.New("wallet not found")
	}

	return nil
}

func (r *WalletRepo) Delete(ctx context.Context, wlt *wallet.Wallet, userId wallet.UserID) error {
	r.logger.Debug("Deleting wallet",
		zap.Int("WalletId", int(wlt.ID())),
		zap.Int("UserId", int(userId)))

	query := `DELETE FROM wallets WHERE id = $1 AND user_id = $2`

	res, err := r.db.ExecContext(ctx, query, wlt.ID(), userId)
	if err != nil {
		r.logger.Error("Failed to delete wallet",
			zap.Error(err),
			zap.Int("WalletId", int(wlt.ID())),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to delete wallet: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to check affected rows for wallet deletion",
			zap.Error(err),
			zap.Int("WalletId", int(wlt.ID())),
			zap.Int("UserId", int(userId)))
		return fmt.Errorf("failed to check affected rows for wallet deletion: %w", err)
	}
	if rowsAffected == 0 {
		r.logger.Warn("Wallet not found for deletion",
			zap.Int("WalletId", int(wlt.ID())),
			zap.Int("UserId", int(userId)))
		return errors.New("wallet not found")
	}

	return nil
}

func (r *WalletRepo) SearchHighestBalance(ctx context.Context, userID wallet.UserID) (*wallet.Wallet, error) {
	query := `
        SELECT id, user_id, wallet_name, wallet_type, balance, created_at, updated_at 
        FROM wallets 
        WHERE user_id = $1 
        ORDER BY balance DESC 
        LIMIT 1
    `
	r.logger.Debug("SearchHighestBalance query", zap.String("Query", query))

	var (
		id         uint64
		dbUserID   uint64
		walletName string
		walletType string
		balance    uint64
		createdAt  time.Time
		updatedAt  time.Time
	)

	row := r.db.QueryRowContext(ctx, query, userID)
	err := row.Scan(
		&id,
		&dbUserID,
		&walletName,
		&walletType,
		&balance,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Highest balance wallet not found", zap.Int("UserId", int(userID)))
			return nil, fmt.Errorf("wallet not found: %w", err)
		}
		r.logger.Error("Failed to find highest balance wallet",
			zap.Error(err),
			zap.Int("UserId", int(userID)))
		return nil, fmt.Errorf("failed to find highest balance wallet: %w", err)
	}

	wlt := wallet.ReconstituteWallet(
		wallet.ID(id),
		wallet.UserID(dbUserID),
		walletName,
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
            w.wallet_name, 
            w.wallet_type, 
            w.balance, 
            w.created_at, 
            w.updated_at,
            COUNT(t.id) AS tx_count
        FROM wallets w
        LEFT JOIN transactions t ON w.id = t.wallet_id
        WHERE w.user_id = $1
        GROUP BY w.id, w.user_id, w.wallet_name, w.wallet_type, w.balance, w.created_at, w.updated_at
        ORDER BY tx_count DESC
        LIMIT 1
    `
	r.logger.Debug("SearchMostActive query", zap.String("Query", query))

	var (
		id         uint64
		dbUserID   uint64
		walletName string
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
		&walletName,
		&walletType,
		&balance,
		&createdAt,
		&updatedAt,
		&txCount,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("Most active wallet not found", zap.Int("UserId", int(userID)))
			return nil, fmt.Errorf("wallet not found: %w", err)
		}
		r.logger.Error("Failed to find most active wallet",
			zap.Error(err),
			zap.Int("UserId", int(userID)))
		return nil, fmt.Errorf("failed to find most active wallet: %w", err)
	}

	wlt := wallet.ReconstituteWallet(
		wallet.ID(id),
		wallet.UserID(dbUserID),
		walletName,
		walletType,
		wallet.Balance(balance),
		createdAt,
		updatedAt,
	)
	return wlt, nil
}

func (r *WalletRepo) SearchTotalBalance(ctx context.Context, userID wallet.UserID) (uint64, error) {
	query := `SELECT COALESCE(SUM(balance), 0)::BIGINT 
	          FROM wallets 
	          WHERE user_id = $1`
	r.logger.Debug("SearchTotalBalance query", zap.String("Query", query))

	var balance int64
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&balance)
	if err != nil {
		r.logger.Error("Failed to calculate total balance",
			zap.Error(err),
			zap.Int("UserId", int(userID)))
		return 0, fmt.Errorf("failed to calculate total balance: %w", err)
	}
	if balance < 0 {
		balance = 0
	}
	return uint64(balance), nil
}
