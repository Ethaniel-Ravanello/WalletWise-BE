package wallet

import (
	"context"
	"time"
	"walletwise/internal/domain/wallet"
)

type WalletInput struct {
	userId     uint64
	walletName string
	walletType string
	balance    uint64
}

type WalletUpdateInput struct {
	ID         uint64
	userID     uint64
	walletName uint64
	walletType uint64
	balance    uint64
}

type Service struct {
	repo wallet.Repository
}

func NewWalletService(repo wallet.Repository) *Service { return &Service{repo: repo} }

func (w *Service) CreateWallet(ctx context.Context, input WalletInput) error {
	wallets, err := wallet.NewWallet(
		wallet.UserID(input.userId),
		input.walletName,
		input.walletType,
		wallet.Balance(input.balance),
		time.Now(),
		time.Now())
	if err != nil {
		return err
	}
	err = w.repo.Save(ctx, wallets)
	if err != nil {
		return err
	}
	return nil
}
