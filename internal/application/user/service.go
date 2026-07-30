package user

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
	"walletwise/pkg/jwt"

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

func NewService(repo users.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, input UserInput) (*users.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	newUser, err := users.NewUser(
		input.Username,
		input.Email,
		string(hashedPassword),
		users.MonthlyLimit(input.MonthlyLimit),
		input.IsActive,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("create user entity: %w", err)
	}

	if err := s.repo.Save(ctx, newUser); err != nil {
		return nil, fmt.Errorf("save user: %w", err)
	}
	return newUser, nil
}

func (s *Service) SearchUserById(ctx context.Context, userID uint64) (*users.User, error) {
	user, err := s.repo.FindByID(ctx, users.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	return user, nil
}

func (s *Service) SearchUserByEmail(ctx context.Context, email string) (*users.User, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return user, nil
}

func (s *Service) UpdateUser(ctx context.Context, input *UserUpdateInput, userId uint64) error {
	if input.ID != userId {
		return fmt.Errorf("invalid user ID")
	}

	input.ID = userId

	existingUser, err := s.SearchUserById(ctx, userId)
	if err != nil {
		return fmt.Errorf("find existing user: %w", err)
	}

	err = existingUser.UpdateUser(
		input.Username,
		input.Email,
		input.Password,
		users.MonthlyLimit(input.MonthlyLimit),
		input.IsActive,
	)
	if err != nil {
		return fmt.Errorf("update user entity: %w", err)
	}

	if err := s.repo.Update(ctx, existingUser); err != nil {
		return fmt.Errorf("update user in repo: %w", err)
	}
	return nil
}

func (s *Service) DeleteUser(ctx context.Context, input UserUpdateInput, userId uint64) error {
	if input.ID != userId {
		return fmt.Errorf("invalid user ID")
	}

	input.ID = userId

	existingUser, err := s.SearchUserById(ctx, input.ID)
	if err != nil {
		return fmt.Errorf("find existing user: %w", err)
	}

	if err := s.repo.Delete(ctx, existingUser); err != nil {
		return fmt.Errorf("delete user: %w", err)
	}
	return nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (string, error) {
	user, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("Invalid User")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password()), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	token, err := jwt.GenerateJwtCustomClaims(uint64(user.UserID()))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	fmt.Println(token)
	return token, nil
}
