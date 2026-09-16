package wallet

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"go.uber.org/zap"

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
	repo   wallet.Repository
	logger *zap.Logger
}

func NewService(repo wallet.Repository) *Service {
	return &Service{
		repo:   repo,
		logger: zap.L(),
	}
}

// NewWalletService is an alias constructor for backward compatibility
func NewWalletService(repo wallet.Repository) *Service {
	return NewService(repo)
}

func (s *Service) CreateWallet(ctx context.Context, input WalletInput) error {
	s.logger.Debug("Creating wallet in application service",
		zap.Uint64("UserId", input.UserID),
		zap.String("WalletName", input.WalletName),
		zap.String("WalletType", input.WalletType))

	allWallets, err := s.repo.SearchAll(ctx, wallet.UserID(input.UserID))
	if err != nil {
		s.logger.Error("Failed to search existing wallets for creation check",
			zap.Error(err),
			zap.Uint64("UserId", input.UserID))
		return fmt.Errorf("search existing wallets: %w", err)
	}

	if len(allWallets) >= 10 {
		s.logger.Warn("Maximum number of wallets reached for user",
			zap.Uint64("UserId", input.UserID),
			zap.Int("wallet_count", len(allWallets)))
		return ErrMaxWalletsReached
	}

	for _, existWallet := range allWallets {
		if strings.EqualFold(existWallet.Name(), input.WalletName) {
			s.logger.Warn("Wallet with duplicate name already exists",
				zap.Uint64("UserId", input.UserID),
				zap.String("WalletName", input.WalletName))
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
		s.logger.Error("Failed to create wallet domain entity",
			zap.Error(err),
			zap.Uint64("UserId", input.UserID),
			zap.String("WalletName", input.WalletName))
		return fmt.Errorf("create wallet entity: %w", err)
	}

	if err := s.repo.Save(ctx, newWallet); err != nil {
		s.logger.Error("Failed to save wallet in repository",
			zap.Error(err),
			zap.Uint64("UserId", input.UserID),
			zap.String("WalletName", input.WalletName))
		return fmt.Errorf("save wallet: %w", err)
	}

	s.logger.Info("Wallet created successfully in application service",
		zap.Uint64("UserId", input.UserID),
		zap.String("WalletName", input.WalletName))
	return nil
}

func (s *Service) SearchAllWallet(ctx context.Context, userID uint64) ([]*wallet.Wallet, error) {
	s.logger.Debug("Searching all wallets for user", zap.Uint64("UserId", userID))

	allWallets, err := s.repo.SearchAll(ctx, wallet.UserID(userID))
	if err != nil {
		s.logger.Error("Failed to search wallets in repository",
			zap.Error(err),
			zap.Uint64("UserId", userID))
		return nil, fmt.Errorf("search wallets: %w", err)
	}

	if len(allWallets) == 0 {
		s.logger.Debug("No wallets found for user", zap.Uint64("UserId", userID))
		return nil, ErrWalletsEmpty
	}

	s.logger.Debug("Wallets retrieved successfully",
		zap.Uint64("UserId", userID),
		zap.Int("count", len(allWallets)))
	return allWallets, nil
}

func (s *Service) SearchWalletByID(ctx context.Context, walletID uint64, userId uint64) (*wallet.Wallet, error) {
	s.logger.Debug("Searching wallet by ID in repository",
		zap.Uint64("WalletId", walletID),
		zap.Uint64("UserId", userId))

	w, err := s.repo.SearchByID(ctx, wallet.ID(walletID), wallet.UserID(userId))
	if err != nil {
		s.logger.Warn("Wallet not found by ID",
			zap.Error(err),
			zap.Uint64("WalletId", walletID),
			zap.Uint64("UserId", userId))
		return nil, fmt.Errorf("search wallet by id: %w", err)
	}

	s.logger.Debug("Wallet retrieved successfully",
		zap.Uint64("WalletId", walletID),
		zap.Uint64("UserId", userId))
	return w, nil
}

func (s *Service) UpdateWallet(ctx context.Context, walletUpdate WalletUpdateInput, userId uint64) error {
	if walletUpdate.UserID != userId {
		s.logger.Warn("Unauthorized wallet update: user ID mismatch",
			zap.Uint64("WalletUserId", walletUpdate.UserID),
			zap.Uint64("UserId", userId))
		return ErrInvalidUserID
	}

	walletUpdate.UserID = userId

	s.logger.Debug("Updating wallet",
		zap.Uint64("WalletId", walletUpdate.ID),
		zap.Uint64("UserId", userId))

	existingWallet, err := s.repo.SearchByID(ctx, wallet.ID(walletUpdate.ID), wallet.UserID(walletUpdate.UserID))
	if err != nil {
		s.logger.Warn("Wallet not found for update",
			zap.Error(err),
			zap.Uint64("WalletId", walletUpdate.ID),
			zap.Uint64("UserId", userId))
		return ErrWalletNotFound
	}
	if existingWallet.UserID() != wallet.UserID(walletUpdate.UserID) {
		s.logger.Warn("Unauthorized access attempt to update wallet",
			zap.Uint64("WalletId", walletUpdate.ID),
			zap.Uint64("UserId", userId))
		return ErrUnauthorizedWallet
	}

	if !strings.EqualFold(walletUpdate.WalletName, existingWallet.Name()) {
		allWallets, err := s.repo.SearchAll(ctx, wallet.UserID(walletUpdate.UserID))
		if err != nil {
			s.logger.Error("Failed to search wallets for duplicate check",
				zap.Error(err),
				zap.Uint64("UserId", walletUpdate.UserID))
			return fmt.Errorf("search wallets for duplicate check: %w", err)
		}
		for _, w := range allWallets {
			if w.ID() != existingWallet.ID() && strings.EqualFold(w.Name(), walletUpdate.WalletName) {
				s.logger.Warn("Cannot update wallet: duplicate wallet name",
					zap.Uint64("WalletId", walletUpdate.ID),
					zap.String("WalletName", walletUpdate.WalletName))
				return ErrDuplicateWalletName
			}
		}
	}

	err = existingWallet.UpdateWallet(
		wallet.ID(walletUpdate.ID),
		wallet.UserID(walletUpdate.UserID),
		walletUpdate.WalletName,
		walletUpdate.WalletType,
		existingWallet.Balance(),
		existingWallet.CreatedAt(),
		time.Now(),
	)
	if err != nil {
		s.logger.Error("Failed to update wallet domain entity",
			zap.Error(err),
			zap.Uint64("WalletId", walletUpdate.ID))
		return fmt.Errorf("update wallet entity: %w", err)
	}

	if err := s.repo.Update(ctx, existingWallet, wallet.UserID(walletUpdate.UserID)); err != nil {
		s.logger.Error("Failed to update wallet in repository",
			zap.Error(err),
			zap.Uint64("WalletId", walletUpdate.ID),
			zap.Uint64("UserId", walletUpdate.UserID))
		return fmt.Errorf("update wallet in repo: %w", err)
	}

	s.logger.Info("Wallet updated successfully in application service",
		zap.Uint64("WalletId", walletUpdate.ID),
		zap.Uint64("UserId", userId))
	return nil
}

func (s *Service) DeleteWallet(ctx context.Context, input WalletUpdateInput, userId uint64) error {
	if input.UserID != userId {
		s.logger.Warn("Unauthorized wallet deletion: user ID mismatch",
			zap.Uint64("InputUserId", input.UserID),
			zap.Uint64("UserId", userId))
		return ErrInvalidUserID
	}

	input.UserID = userId

	s.logger.Debug("Deleting wallet",
		zap.Uint64("WalletId", input.ID),
		zap.Uint64("UserId", userId))

	existingWallet, err := s.repo.SearchByID(ctx, wallet.ID(input.ID), wallet.UserID(input.UserID))
	if err != nil {
		s.logger.Warn("Wallet not found for deletion",
			zap.Error(err),
			zap.Uint64("WalletId", input.ID),
			zap.Uint64("UserId", userId))
		return ErrWalletNotFound
	}

	if existingWallet.UserID() != wallet.UserID(input.UserID) {
		s.logger.Warn("Unauthorized access attempt to delete wallet",
			zap.Uint64("WalletId", input.ID),
			zap.Uint64("UserId", userId))
		return ErrUnauthorizedWallet
	}

	if err := s.repo.Delete(ctx, existingWallet, wallet.UserID(input.UserID)); err != nil {
		s.logger.Error("Failed to delete wallet in repository",
			zap.Error(err),
			zap.Uint64("WalletId", input.ID),
			zap.Uint64("UserId", input.UserID))
		return fmt.Errorf("delete wallet: %w", err)
	}

	s.logger.Info("Wallet deleted successfully in application service",
		zap.Uint64("WalletId", input.ID),
		zap.Uint64("UserId", userId))
	return nil
}

func (s *Service) SearchHighestBalanceWallet(ctx context.Context, userID wallet.UserID) (*wallet.Wallet, error) {
	if userID <= 0 {
		s.logger.Warn("Invalid user ID for highest balance search", zap.Uint64("UserId", uint64(userID)))
		return nil, ErrInvalidUserID
	}

	s.logger.Debug("Searching highest balance wallet in repository", zap.Uint64("UserId", uint64(userID)))

	highWallet, err := s.repo.SearchHighestBalance(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to find highest balance wallet in repository",
			zap.Error(err),
			zap.Uint64("UserId", uint64(userID)))
		return nil, fmt.Errorf("get highest balance wallet: %w", err)
	}

	s.logger.Debug("Highest balance wallet found successfully", zap.Uint64("UserId", uint64(userID)))
	return highWallet, nil
}

func (s *Service) SearchMostActiveWallet(ctx context.Context, userID wallet.UserID) (*wallet.Wallet, error) {
	if userID <= 0 {
		s.logger.Warn("Invalid user ID for most active wallet search", zap.Uint64("UserId", uint64(userID)))
		return nil, ErrInvalidUserID
	}

	s.logger.Debug("Searching most active wallet in repository", zap.Uint64("UserId", uint64(userID)))

	activeWallet, err := s.repo.SearchMostActive(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to find most active wallet in repository",
			zap.Error(err),
			zap.Uint64("UserId", uint64(userID)))
		return nil, fmt.Errorf("get most active wallet: %w", err)
	}

	s.logger.Debug("Most active wallet found successfully", zap.Uint64("UserId", uint64(userID)))
	return activeWallet, nil
}

func (s *Service) SearchTotalBalanceWallet(ctx context.Context, userID wallet.UserID) (uint64, error) {
	if userID <= 0 {
		s.logger.Warn("Invalid user ID for total balance search", zap.Uint64("UserId", uint64(userID)))
		return 0, ErrInvalidUserID
	}

	s.logger.Debug("Searching total balance in repository", zap.Uint64("UserId", uint64(userID)))

	totalBalance, err := s.repo.SearchTotalBalance(ctx, userID)
	if err != nil {
		s.logger.Error("Failed to get total balance in repository",
			zap.Error(err),
			zap.Uint64("UserId", uint64(userID)))
		return 0, fmt.Errorf("get total balance: %w", err)
	}

	s.logger.Debug("Total balance retrieved successfully",
		zap.Uint64("UserId", uint64(userID)),
		zap.Uint64("TotalBalance", totalBalance))
	return totalBalance, nil
}
