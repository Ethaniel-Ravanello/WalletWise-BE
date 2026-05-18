package wallet

import (
	"errors"
	"time"
)

type WalletID uint64
type UserID uint64
type Balance uint64

type Wallet struct {
	id         WalletID
	userID     UserID
	walletName string
	walletType string
	balance    Balance
	createdAt  time.Time
	updatedAt  time.Time
}

func NewWallet(userID UserID, walletName string, walletType string, balance Balance, createdAt time.Time, updatedAt time.Time) (*Wallet, error) {
	if userID <= 0 {
		return nil, errors.New(`invalid user id`)
	}
	if walletName == "" {
		return nil, errors.New(`invalid name`)
	}
	if walletType == "" {
		return nil, errors.New(`invalid wallet type`)
	}
	if createdAt.IsZero() {
		return nil, errors.New(`invalid creation date`)
	}
	if updatedAt.IsZero() {
		return nil, errors.New(`invalid update date`)
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
	id WalletID,
	userID UserID,
	walletName string,
	walletType string,
	balance Balance,
	createdAt time.Time,
	updatedAt time.Time) *Wallet {
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

func (w *Wallet) UpdateWallet(id WalletID, userId UserID, walletName string, walletType string, balance Balance, createdAt time.Time, updatedAt time.Time) error {
	if userId <= 0 {
		return errors.New(`invalid user id`)
	}
	if walletName == "" {
		return errors.New(`invalid name`)
	}
	if walletType == "" {
		return errors.New(`invalid wallet type`)
	}
	if createdAt.IsZero() {
		return errors.New(`invalid creation date`)
	}
	if updatedAt.IsZero() {
		return errors.New(`invalid update date`)
	}
	w.userID = userId
	w.id = id
	w.walletName = walletName
	w.walletType = walletType
	w.balance = balance
	w.createdAt = createdAt
	w.updatedAt = updatedAt
	return nil
}

func (w *Wallet) ID() WalletID         { return w.id }
func (w *Wallet) UserID() UserID       { return w.userID }
func (w *Wallet) Name() string         { return w.walletName }
func (w *Wallet) WalletType() string   { return w.walletType }
func (w *Wallet) Balance() Balance     { return w.balance }
func (w *Wallet) CreatedAt() time.Time { return w.createdAt }
func (w *Wallet) UpdatedAt() time.Time { return w.updatedAt }
