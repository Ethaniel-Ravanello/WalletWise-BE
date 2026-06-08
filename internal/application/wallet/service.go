package wallet

import (
	"context"
	"errors"
	"strings"
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
	walletName string
	walletType string
	balance    uint64
}

type Service struct {
	repo wallet.Repository
}

func NewWalletService(repo wallet.Repository) *Service { return &Service{repo: repo} }

func (w *Service) CreateWallet(ctx context.Context, input WalletInput) error {
	allWallets, err := w.repo.SearchAll(ctx, wallet.UserID(input.userId))
	if err != nil {
		return err
	}
	if len(allWallets) == 10 {
		return errors.New("Wallet Mencapai Jumlah Maksimal")
	}

	for _, existWallet := range allWallets {
		if strings.EqualFold(existWallet.Name(), input.walletName) {
			return errors.New("Duplikat Wallet")
		}
	}

	wallets, err := wallet.NewWallet(
		wallet.UserID(input.userId),
		input.walletName,
		input.walletType,
		0,
		time.Now(),
		time.Now())
	if err != nil {
		return err
	}
	return w.repo.Save(ctx, wallets)
}

func (w *Service) SearchAllWallet(ctx context.Context, userID uint64) ([]*wallet.Wallet, error) {
	allWallets, err := w.repo.SearchAll(ctx, wallet.UserID(userID))
	if err != nil {
		return nil, err
	}
	if len(allWallets) == 0 {
		return nil, errors.New("Wallet Kosong")
	}
	return allWallets, nil
}

func (w *Service) SearchWalletByID(ctx context.Context, walletID uint64) (*wallet.Wallet, error) {
	wallet, err := w.repo.SearchByID(ctx, wallet.WalletID(walletID))
	if err != nil {
		return nil, err
	}
	return wallet, nil
}

func (w *Service) UpdateWallet(ctx context.Context, walletUpdate WalletUpdateInput) error {
	checkWallet, err := w.repo.SearchByID(ctx, wallet.WalletID(walletUpdate.ID))
	if err != nil {
		return errors.New("Wallet Tidak Ditemukan")
	}
	if checkWallet.UserID() != wallet.UserID(walletUpdate.userID) {
		return errors.New("Invalid User Wallet")
	}
	if !strings.EqualFold(walletUpdate.walletName, checkWallet.Name()) {
		checkName, err := w.repo.SearchByID(ctx, wallet.WalletID(walletUpdate.ID))
		if err != nil {
			return errors.New("Internal Error")
		}
		if strings.EqualFold(checkName.Name(), walletUpdate.walletName) {
			return errors.New("Wallet Duplicate")
		}
	}
	err = checkWallet.UpdateWallet(
		wallet.WalletID(walletUpdate.ID),
		wallet.UserID(walletUpdate.userID),
		walletUpdate.walletName,
		walletUpdate.walletType,
		checkWallet.Balance(),
		checkWallet.CreatedAt(),
		time.Now())
	if err != nil {
		return err
	}
	return w.repo.Update(ctx, checkWallet)
}

func (w *Service) DeleteWallet(ctx context.Context, input WalletInput) error {
	checkWallet, err := w.repo.SearchByID(ctx, wallet.WalletID(input.userId))
	if err != nil {
		return errors.New("Wallet Tidak Ditemukan")
	}

	if checkWallet.UserID() != wallet.UserID(input.userId) {
		return errors.New("Invalid User Wallet")
	}
	return w.repo.Delete(ctx, checkWallet)
}

func (w *Service) SearchHighestBalanceWallet(ctx context.Context, userId wallet.UserID) (*wallet.Wallet, error) {
	if userId <= 0 {
		return nil, errors.New("Invalid User ID")
	}
	highWallet, err := w.repo.SearchHighestBalance(ctx, userId)
	if err != nil {
		return nil, err
	}
	return highWallet, nil
}

func (w *Service) SearchMostActiveWallet(ctx context.Context, userId wallet.UserID) (*wallet.Wallet, error) {
	if userId <= 0 {
		return nil, errors.New("Invalid User ID")
	}
	activeWallet, err := w.repo.SearchMostActive(ctx, userId)
	if err != nil {
		return nil, err
	}
	return activeWallet, nil
}

func (w *Service) SearchTotalBalanceWallet(ctx context.Context, userId wallet.UserID) (uint64, error) {
	if userId <= 0 {
		return 0, errors.New("Invalid User ID")
	}
	totalBalance, err := w.repo.SearchTotalBalance(ctx, userId)
	if err != nil {
		return 0, err
	}
	return totalBalance, nil
}
