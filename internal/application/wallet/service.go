package wallet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"walletwise/internal/domain/wallet"
)

var (
	ErrMaxWalletsReached   = errors.New("maximum number of wallets reached")
	ErrDuplicateWalletName = errors.New("wallet with this name already exists")
	ErrWalletNotFound      = errors.New("wallet not found")
	ErrUnauthorizedWallet  = errors.New("unauthorized access to wallet")
	ErrInvalidUserID       = errors.New("invalid user id")
	ErrWalletsEmpty        = errors.New("no wallets found")
)

type WalletInput struct {
	ID         uint64
	UserID     uint64
	WalletName string
	WalletType string
	Balance    uint64
}

type WalletUpdateInput struct {
	ID         uint64
	UserID     uint64
	WalletName string
	WalletType string
	Balance    uint64
}

type Service struct {
	repo wallet.Repository
}

func NewService(repo wallet.Repository) *Service {
	return &Service{repo: repo}
}

// NewWalletService is an alias constructor for backward compatibility
func NewWalletService(repo wallet.Repository) *Service {
	return NewService(repo)
}

func (s *Service) CreateWallet(ctx context.Context, input WalletInput) error {
	allWallets, err := s.repo.SearchAll(ctx, wallet.UserID(input.UserID))
	if err != nil {
		return fmt.Errorf("search existing wallets: %w", err)
	}

	if len(allWallets) >= 10 {
		return ErrMaxWalletsReached
	}

	for _, existWallet := range allWallets {
		if strings.EqualFold(existWallet.Name(), input.WalletName) {
			return ErrDuplicateWalletName
		}
	}

	newWallet, err := wallet.NewWallet(
		wallet.UserID(input.UserID),
		input.WalletName,
		input.WalletType,
		0,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("create wallet entity: %w", err)
	}

	if err := s.repo.Save(ctx, newWallet); err != nil {
		return fmt.Errorf("save wallet: %w", err)
	}
	return nil
}

func (s *Service) SearchAllWallet(ctx context.Context, userID uint64) ([]*wallet.Wallet, error) {
	allWallets, err := s.repo.SearchAll(ctx, wallet.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("search wallets: %w", err)
	}
	if len(allWallets) == 0 {
		return nil, ErrWalletsEmpty
	}
	return allWallets, nil
}

func (s *Service) SearchWalletByID(ctx context.Context, walletID uint64, userId uint64) (*wallet.Wallet, error) {
	w, err := s.repo.SearchByID(ctx, wallet.WalletID(walletID), wallet.UserID(userId))
	if err != nil {
		return nil, fmt.Errorf("search wallet by id: %w", err)
	}
	return w, nil
}

func (s *Service) UpdateWallet(ctx context.Context, walletUpdate WalletUpdateInput, userId uint64) error {
	if walletUpdate.UserID != userId {
		return ErrInvalidUserID
	}

	walletUpdate.UserID = userId

	existingWallet, err := s.repo.SearchByID(ctx, wallet.WalletID(walletUpdate.ID), wallet.UserID(walletUpdate.UserID))
	if err != nil {
		return ErrWalletNotFound
	}
	if existingWallet.UserID() != wallet.UserID(walletUpdate.UserID) {
		return ErrUnauthorizedWallet
	}

	if !strings.EqualFold(walletUpdate.WalletName, existingWallet.Name()) {
		allWallets, err := s.repo.SearchAll(ctx, wallet.UserID(walletUpdate.UserID))
		if err != nil {
			return fmt.Errorf("search wallets for duplicate check: %w", err)
		}
		for _, w := range allWallets {
			if w.ID() != existingWallet.ID() && strings.EqualFold(w.Name(), walletUpdate.WalletName) {
				return ErrDuplicateWalletName
			}
		}
	}

	err = existingWallet.UpdateWallet(
		wallet.WalletID(walletUpdate.ID),
		wallet.UserID(walletUpdate.UserID),
		walletUpdate.WalletName,
		walletUpdate.WalletType,
		existingWallet.Balance(),
		existingWallet.CreatedAt(),
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("update wallet entity: %w", err)
	}

	if err := s.repo.Update(ctx, existingWallet, wallet.UserID(walletUpdate.UserID)); err != nil {
		return fmt.Errorf("update wallet in repo: %w", err)
	}
	return nil
}

func (s *Service) DeleteWallet(ctx context.Context, input WalletInput, userId uint64) error {
	if input.UserID != userId {
		return ErrInvalidUserID
	}

	input.UserID = userId

	existingWallet, err := s.repo.SearchByID(ctx, wallet.WalletID(input.ID), wallet.UserID(input.UserID))
	if err != nil {
		return ErrWalletNotFound
	}

	if existingWallet.UserID() != wallet.UserID(input.UserID) {
		return ErrUnauthorizedWallet
	}

	if err := s.repo.Delete(ctx, existingWallet, wallet.UserID(input.UserID)); err != nil {
		return fmt.Errorf("delete wallet: %w", err)
	}
	return nil
}

func (s *Service) SearchHighestBalanceWallet(ctx context.Context, userID wallet.UserID) (*wallet.Wallet, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	highWallet, err := s.repo.SearchHighestBalance(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get highest balance wallet: %w", err)
	}
	return highWallet, nil
}

func (s *Service) SearchMostActiveWallet(ctx context.Context, userID wallet.UserID) (*wallet.Wallet, error) {
	if userID <= 0 {
		return nil, ErrInvalidUserID
	}
	activeWallet, err := s.repo.SearchMostActive(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get most active wallet: %w", err)
	}
	return activeWallet, nil
}

func (s *Service) SearchTotalBalanceWallet(ctx context.Context, userID wallet.UserID) (uint64, error) {
	if userID <= 0 {
		return 0, ErrInvalidUserID
	}
	totalBalance, err := s.repo.SearchTotalBalance(ctx, userID)
	if err != nil {
		return 0, fmt.Errorf("get total balance: %w", err)
	}
	return totalBalance, nil
}
