package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"go.uber.org/zap"

	"walletwise/internal/domain/user"
)

type UserRepo struct {
	db     *sql.DB
	logger *zap.Logger
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{
		db:     db,
		logger: zap.L(),
	}
}

var _ user.Repository = (*UserRepo)(nil)

func (r *UserRepo) Save(ctx context.Context, u *user.User) error {
	r.logger.Debug("Saving user",
		zap.String("Username", u.Username()),
		zap.String("Email", u.Email()))

	query := `INSERT INTO users (username, email, password, monthly_limit, is_active, created_at, updated_at) 
	          VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := r.db.ExecContext(ctx, query,
		u.Username(),
		u.Email(),
		u.Password(),
		u.MonthlyLimit(),
		u.IsActive(),
		time.Now(),
		time.Now(),
	)
	if err != nil {
		r.logger.Error("Failed to save user",
			zap.Error(err),
			zap.String("Username", u.Username()),
			zap.String("Email", u.Email()))
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

func (r *UserRepo) FindByID(ctx context.Context, id user.UserID) (*user.User, error) {
	query := `SELECT id, username, email, password, monthly_limit, is_active, created_at, updated_at 
	          FROM users WHERE id = $1`
	r.logger.Debug("FindByID query", zap.String("Query", query))

	var (
		userID       uint64
		username     string
		email        string
		password     string
		monthlyLimit uint64
		isActive     bool
		createdAt    time.Time
		updatedAt    time.Time
	)

	row := r.db.QueryRowContext(ctx, query, id)
	err := row.Scan(
		&userID,
		&username,
		&email,
		&password,
		&monthlyLimit,
		&isActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("User not found by id", zap.Int("UserId", int(id)))
			return nil, fmt.Errorf("user not found: %w", err)
		}
		r.logger.Error("Failed to find user by id",
			zap.Error(err),
			zap.Int("UserId", int(id)))
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	u := user.ReconstituteUser(
		user.UserID(userID),
		username,
		email,
		password,
		user.MonthlyLimit(monthlyLimit),
		isActive,
		createdAt,
		updatedAt,
	)
	return u, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*user.User, error) {
	query := `SELECT id, username, email, password, monthly_limit, is_active, created_at, updated_at 
	          FROM users WHERE email = $1`
	r.logger.Debug("FindByEmail query", zap.String("Query", query))

	var (
		userID       uint64
		username     string
		emailHolder  string
		password     string
		monthlyLimit uint64
		isActive     bool
		createdAt    time.Time
		updatedAt    time.Time
	)

	row := r.db.QueryRowContext(ctx, query, email)
	err := row.Scan(
		&userID,
		&username,
		&emailHolder,
		&password,
		&monthlyLimit,
		&isActive,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			r.logger.Warn("User not found by email", zap.String("Email", email))
			return nil, fmt.Errorf("user not found: %w", err)
		}
		r.logger.Error("Failed to find user by email",
			zap.Error(err),
			zap.String("Email", email))
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	u := user.ReconstituteUser(
		user.UserID(userID),
		username,
		emailHolder,
		password,
		user.MonthlyLimit(monthlyLimit),
		isActive,
		createdAt,
		updatedAt,
	)
	return u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *user.User) error {
	r.logger.Debug("Updating user", zap.Int("UserId", int(u.ID())))

	query := `UPDATE users SET username = $1, email = $2, password = $3, monthly_limit = $4, is_active = $5, updated_at = $6 
	          WHERE id = $7`

	res, err := r.db.ExecContext(ctx, query,
		u.Username(),
		u.Email(),
		u.Password(),
		u.MonthlyLimit(),
		u.IsActive(),
		time.Now(),
		u.ID(),
	)
	if err != nil {
		r.logger.Error("Failed to update user",
			zap.Error(err),
			zap.Int("UserId", int(u.ID())))
		return fmt.Errorf("failed to update user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to check affected rows for user update",
			zap.Error(err),
			zap.Int("UserId", int(u.ID())))
		return fmt.Errorf("failed to check affected rows for user update: %w", err)
	}
	if rowsAffected == 0 {
		r.logger.Warn("User not found for update", zap.Int("UserId", int(u.ID())))
		return errors.New("user not found")
	}

	return nil
}

func (r *UserRepo) Delete(ctx context.Context, u *user.User) error {
	r.logger.Debug("Deleting user", zap.Int("UserId", int(u.ID())))

	query := `DELETE FROM users WHERE id = $1`

	res, err := r.db.ExecContext(ctx, query, u.ID())
	if err != nil {
		r.logger.Error("Failed to delete user",
			zap.Error(err),
			zap.Int("UserId", int(u.ID())))
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		r.logger.Error("Failed to check affected rows for user deletion",
			zap.Error(err),
			zap.Int("UserId", int(u.ID())))
		return fmt.Errorf("failed to check affected rows for user deletion: %w", err)
	}
	if rowsAffected == 0 {
		r.logger.Warn("User not found for deletion", zap.Int("UserId", int(u.ID())))
		return errors.New("user not found")
	}

	return nil
}
