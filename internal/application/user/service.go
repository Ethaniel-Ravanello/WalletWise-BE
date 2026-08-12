package user

import (
	"context"
	"errors"
	"fmt"
	"golang.org/x/crypto/bcrypt"
	"time"
	"walletwise/pkg/jwt"

	"walletwise/internal/domain/user"
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
	repo user.Repository
}

func NewService(repo user.Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateUser(ctx context.Context, input UserInput) (*user.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)

	if err != nil {
		return nil, err
	}

	newUser, err := user.NewUser(
		input.Username,
		input.Email,
		string(hashedPassword),
		user.MonthlyLimit(input.MonthlyLimit),
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

func (s *Service) SearchUserById(ctx context.Context, userID uint64) (*user.User, error) {
	u, err := s.repo.FindByID(ctx, user.UserID(userID))
	if err != nil {
		return nil, fmt.Errorf("find user by id: %w", err)
	}
	fmt.Println("ini service", u)
	return u, nil
}

func (s *Service) SearchUserByEmail(ctx context.Context, email string) (*user.User, error) {
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("find user by email: %w", err)
	}
	return u, nil
}

func (s *Service) UpdateUser(ctx context.Context, input *UserUpdateInput, userId uint64) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
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
		string(hashedPassword),
		user.MonthlyLimit(input.MonthlyLimit),
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
	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("Invalid User")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password()), []byte(password))
	if err != nil {
		return "", errors.New("invalid email or password")
	}

	token, err := jwt.GenerateJwtCustomClaims(uint64(u.ID()))
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	fmt.Println(token)
	return token, nil
}

