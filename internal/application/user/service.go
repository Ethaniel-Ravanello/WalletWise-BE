package user

import (
	"context"
	"fmt"
	"time"
	"walletwise/internal/domain/users"
)

type UserInput struct {
	Username     string
	Email        string
	Password     string
	MonthlyLimit uint64
	IsActive     bool
}

type UserUpdateInput struct {
	ID           uint64
	Username     string
	Email        string
	Password     string
	MonthlyLimit uint64
	IsActive     bool
}

type Service struct {
	repo users.Repository
}

func NewService(repo users.Repository) *Service { return &Service{repo: repo} }

func (u *Service) CreateUser(ctx context.Context, input UserInput) (*users.User, error) {
	trx, err := users.NewUser(
		input.Username,
		input.Email,
		input.Password,
		users.MonthlyLimit(input.MonthlyLimit),
		input.IsActive,
		time.Now(),
		time.Now())
	if err != nil {
		return nil, fmt.Errorf("error creating new users: %w", err)
	}
	if err := u.repo.Save(ctx, trx); err != nil {
		return nil, fmt.Errorf("error saving users: %w", err)
	}
	return trx, nil
}

func (u *Service) SearchUserById(ctx context.Context, userId users.UserID) (*users.User, error) {
	user, err := u.repo.FindByID(ctx, userId)
	if err != nil {
		return nil, fmt.Errorf("error finding users: %w", err)
	}
	return user, nil
}

func (u *Service) SearchUserByEmail(ctx context.Context, email string) (*users.User, error) {
	user, err := u.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("error finding users: %w", err)
	}
	return user, nil
}

func (u *Service) UpdateUser(ctx context.Context, userInput *UserUpdateInput) error {
	existingTrx, err := u.SearchUserById(ctx, users.UserID(userInput.ID))
	if err != nil {
		return fmt.Errorf("error finding users: %w", err)
	}
	err = existingTrx.UpdateUser(
		userInput.Username,
		userInput.Email,
		userInput.Password,
		users.MonthlyLimit(userInput.MonthlyLimit),
		userInput.IsActive,
	)
	err = u.repo.Update(ctx, existingTrx)
	if err != nil {
		return fmt.Errorf("error updating users: %w", err)
	}
	return nil
}

func (u *Service) DeleteUser(ctx context.Context, userInput UserUpdateInput) error {
	existingTrx, err := u.SearchUserById(ctx, users.UserID(userInput.ID))

	err = u.repo.Delete(ctx, existingTrx)
	if err != nil {
		return fmt.Errorf("error deleting users: %w", err)
	}
	return nil
}
