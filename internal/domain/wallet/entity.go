package wallet

import (
	"errors"
	"time"
)

type ID uint64
type UserID uint64
type Balance uint64

type Wallet struct {
	id         ID
	userID     UserID
	walletName string
	walletType string
	balance    Balance
	createdAt  time.Time
	updatedAt  time.Time
}

func NewWallet(userID UserID, walletName string, walletType string, balance Balance, createdAt time.Time, updatedAt time.Time) (*Wallet, error) {
	if userID == 0 {
		return nil, errors.New("user ID is required")
	}
	if walletName == "" {
		return nil, errors.New("wallet name is required")
	}
	if walletType == "" {
		return nil, errors.New("wallet type is required")
	}
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	return &Wallet{
		userID:     userID,
		walletName: walletName,
		walletType: walletType,
		balance:    balance,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}, nil
}

func ReconstituteWallet(
	id ID,
	userID UserID,
	walletName string,
	walletType string,
	balance Balance,
	createdAt time.Time,
	updatedAt time.Time,
) *Wallet {
	return &Wallet{
		id:         id,
		userID:     userID,
		walletName: walletName,
		walletType: walletType,
		balance:    balance,
		createdAt:  createdAt,
		updatedAt:  updatedAt,
	}
}

func (w *Wallet) UpdateWallet(id ID, userID UserID, walletName string, walletType string, balance Balance, createdAt time.Time, updatedAt time.Time) error {
	if userID == 0 {
		return errors.New("user ID is required")
	}
	if walletName == "" {
		return errors.New("wallet name is required")
	}
	if walletType == "" {
		return errors.New("wallet type is required")
	}
	if createdAt.IsZero() {
		createdAt = w.createdAt
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}
	w.id = id
	w.userID = userID
	w.walletName = walletName
	w.walletType = walletType
	w.balance = balance
	w.createdAt = createdAt
	w.updatedAt = updatedAt
	return nil
}

func (w *Wallet) ID() ID               { return w.id }
func (w *Wallet) UserID() UserID       { return w.userID }
func (w *Wallet) Name() string         { return w.walletName }
func (w *Wallet) WalletType() string   { return w.walletType }
func (w *Wallet) Balance() Balance     { return w.balance }
func (w *Wallet) CreatedAt() time.Time { return w.createdAt }
func (w *Wallet) UpdatedAt() time.Time { return w.updatedAt }
