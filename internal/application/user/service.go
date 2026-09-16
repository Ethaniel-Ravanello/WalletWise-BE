package user

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"

	"walletwise/internal/domain/user"
	"walletwise/pkg/jwt"
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
	repo   user.Repository
	logger *zap.Logger
}

func NewService(repo user.Repository) *Service {
	return &Service{
		repo:   repo,
		logger: zap.L(),
	}
}

func (s *Service) CreateUser(ctx context.Context, input UserInput) (*user.User, error) {
	s.logger.Debug("Hashing password for new user",
		zap.String("Username", input.Username),
		zap.String("Email", input.Email))

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash password",
			zap.Error(err),
			zap.String("Email", input.Email))
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
		s.logger.Error("Failed to create user domain entity",
			zap.Error(err),
			zap.String("Email", input.Email))
		return nil, fmt.Errorf("create user entity: %w", err)
	}

	if err := s.repo.Save(ctx, newUser); err != nil {
		s.logger.Error("Failed to save user in repository",
			zap.Error(err),
			zap.String("Email", input.Email))
		return nil, fmt.Errorf("save user: %w", err)
	}

	s.logger.Info("User created successfully in application service",
		zap.Uint64("UserId", uint64(newUser.ID())),
		zap.String("Username", newUser.Username()),
		zap.String("Email", newUser.Email()))
	return newUser, nil
}

func (s *Service) SearchUserById(ctx context.Context, userID uint64) (*user.User, error) {
	s.logger.Debug("Searching user by ID in repository", zap.Uint64("UserId", userID))

	u, err := s.repo.FindByID(ctx, user.UserID(userID))
	if err != nil {
		s.logger.Warn("User not found by ID in repository",
			zap.Error(err),
			zap.Uint64("UserId", userID))
		return nil, fmt.Errorf("find user by id: %w", err)
	}

	s.logger.Debug("User found by ID", zap.Uint64("UserId", userID))
	return u, nil
}

func (s *Service) SearchUserByEmail(ctx context.Context, email string) (*user.User, error) {
	s.logger.Debug("Searching user by email in repository", zap.String("Email", email))

	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("User not found by email in repository",
			zap.Error(err),
			zap.String("Email", email))
		return nil, fmt.Errorf("find user by email: %w", err)
	}

	s.logger.Debug("User found by email", zap.String("Email", email))
	return u, nil
}

func (s *Service) UpdateUser(ctx context.Context, input *UserUpdateInput, userId uint64) error {
	if input.ID != userId {
		s.logger.Warn("Unauthorized user update: input ID does not match session user ID",
			zap.Uint64("InputId", input.ID),
			zap.Uint64("UserId", userId))
		return fmt.Errorf("invalid user ID")
	}
	input.ID = userId

	s.logger.Debug("Updating user details", zap.Uint64("UserId", userId))

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Failed to hash updated password",
			zap.Error(err),
			zap.Uint64("UserId", userId))
		return fmt.Errorf("hash password: %w", err)
	}

	existingUser, err := s.SearchUserById(ctx, userId)
	if err != nil {
		s.logger.Warn("Existing user not found for update",
			zap.Error(err),
			zap.Uint64("UserId", userId))
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
		s.logger.Error("Failed to update user domain entity",
			zap.Error(err),
			zap.Uint64("UserId", userId))
		return fmt.Errorf("update user entity: %w", err)
	}

	if err := s.repo.Update(ctx, existingUser); err != nil {
		s.logger.Error("Failed to update user in repository",
			zap.Error(err),
			zap.Uint64("UserId", userId))
		return fmt.Errorf("update user in repo: %w", err)
	}

	s.logger.Info("User updated successfully in application service", zap.Uint64("UserId", userId))
	return nil
}

func (s *Service) DeleteUser(ctx context.Context, input UserUpdateInput, userId uint64) error {
	if input.ID != userId {
		s.logger.Warn("Unauthorized user deletion: input ID does not match session user ID",
			zap.Uint64("InputId", input.ID),
			zap.Uint64("UserId", userId))
		return fmt.Errorf("invalid user ID")
	}

	input.ID = userId

	s.logger.Debug("Deleting user", zap.Uint64("UserId", userId))

	existingUser, err := s.SearchUserById(ctx, input.ID)
	if err != nil {
		s.logger.Warn("Existing user not found for deletion",
			zap.Error(err),
			zap.Uint64("UserId", userId))
		return fmt.Errorf("find existing user: %w", err)
	}

	if err := s.repo.Delete(ctx, existingUser); err != nil {
		s.logger.Error("Failed to delete user in repository",
			zap.Error(err),
			zap.Uint64("UserId", userId))
		return fmt.Errorf("delete user: %w", err)
	}

	s.logger.Info("User deleted successfully in application service", zap.Uint64("UserId", userId))
	return nil
}

func (s *Service) Login(ctx context.Context, email string, password string) (string, error) {
	s.logger.Debug("Verifying user login credentials", zap.String("Email", email))

	u, err := s.repo.FindByEmail(ctx, email)
	if err != nil {
		s.logger.Warn("User login failed: email not found",
			zap.Error(err),
			zap.String("Email", email))
		return "", fmt.Errorf("Invalid User")
	}

	err = bcrypt.CompareHashAndPassword([]byte(u.Password()), []byte(password))
	if err != nil {
		s.logger.Warn("User login failed: password mismatch", zap.String("Email", email))
		return "", errors.New("invalid email or password")
	}

	token, err := jwt.GenerateJwtCustomClaims(uint64(u.ID()))
	if err != nil {
		s.logger.Error("Failed to generate JWT claims for user",
			zap.Error(err),
			zap.Uint64("UserId", uint64(u.ID())),
			zap.String("Email", email))
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	s.logger.Info("User authenticated successfully and JWT generated",
		zap.Uint64("UserId", uint64(u.ID())),
		zap.String("Email", email))
	return token, nil
}

