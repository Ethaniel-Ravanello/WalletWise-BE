package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"walletwise/internal/domain/users"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

var _ users.Repository = (*UserRepo)(nil)

func (r *UserRepo) Save(ctx context.Context, u *users.User) error {
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
		return fmt.Errorf("failed to save user: %w", err)
	}
	return nil
}

func (r *UserRepo) FindByID(ctx context.Context, id users.UserID) (*users.User, error) {
	query := `SELECT id, username, email, password, monthly_limit, is_active, created_at, updated_at 
	          FROM users WHERE id = $1`

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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}

	u := users.ReconstituteUser(
		users.UserID(userID),
		username,
		email,
		password,
		users.MonthlyLimit(monthlyLimit),
		isActive,
		createdAt,
		updatedAt,
	)
	return u, nil
}

func (r *UserRepo) FindByEmail(ctx context.Context, email string) (*users.User, error) {
	query := `SELECT id, username, email, password, monthly_limit, is_active, created_at, updated_at 
	          FROM users WHERE email = $1`

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
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found: %w", err)
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}

	u := users.ReconstituteUser(
		users.UserID(userID),
		username,
		emailHolder,
		password,
		users.MonthlyLimit(monthlyLimit),
		isActive,
		createdAt,
		updatedAt,
	)
	return u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *users.User) error {
	query := `UPDATE users SET username = $1, email = $2, password = $3, monthly_limit = $4, is_active = $5, updated_at = $6 
	          WHERE id = $7`

	_, err := r.db.ExecContext(ctx, query,
		u.Username(),
		u.Email(),
		u.Password(),
		u.MonthlyLimit(),
		u.IsActive(),
		time.Now(),
		u.UserID(),
	)
	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, u *users.User) error {
	query := `DELETE FROM users WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, u.UserID())
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
