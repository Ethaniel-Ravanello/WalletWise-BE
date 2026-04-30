package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"walletwise/internal/domain/users"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

var _ users.Repository = (*UserRepo)(nil)

func (u UserRepo) Save(ctx context.Context, user *users.User) error {
	const sql = `INSERT INTO users (username, email, password, monthly_limit, is_active, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`

	_, err := u.db.ExecContext(ctx, sql,
		user.Username(),
		user.Email(),
		user.Password(),
		user.MonthlyLimit(),
		true,
		time.Now(),
		time.Now(),
	)
	if err != nil {
		return errors.New("cannot save users: " + err.Error())
	}
	return nil
}

func (u UserRepo) FindByID(ctx context.Context, id users.UserID) (*users.User, error) {
	const sql = `SELECT id, username, email, password, monthly_limmit, is_active, created_at, updated_at FROM users WHERE id = $1`

	var (
		userID        uint64
		username      string
		email         string
		password      string
		monthly_limit uint64
		isActive      bool
		createdAt     time.Time
		updatedAt     time.Time
	)
	row := u.db.QueryRowContext(ctx, sql, id)
	err := row.Scan(
		&userID,
		&username,
		&email,
		&password,
		&monthly_limit,
		&isActive,
		&createdAt,
		&updatedAt)
	if err != nil {
		return nil, errors.New("cannot find users: " + err.Error())
	}
	users := users.ReconstituteUser(
		users.UserID(userID),
		username,
		email,
		password,
		users.MonthlyLimit(monthly_limit),
		isActive,
		createdAt,
		updatedAt)
	return users, nil
}

func (u UserRepo) FindByEmail(ctx context.Context, email string) (*users.User, error) {
	const sql = `SELECT id, username, email, password, monthly_limit, isActive, created_at, updated_at FROM users WHERE email = $1`

	var (
		userID        uint64
		username      string
		emailHolder   string
		password      string
		monthly_limit uint64
		isActive      bool
		createdAt     time.Time
		updatedAt     time.Time
	)
	row := u.db.QueryRowContext(ctx, sql, email)
	err := row.Scan(
		&userID,
		&username,
		&emailHolder,
		&password,
		&monthly_limit,
		&isActive,
		&createdAt,
		&updatedAt)
	if err != nil {
		return nil, errors.New("cannot find users: " + err.Error())
	}
	users := users.ReconstituteUser(
		users.UserID(userID),
		username,
		email,
		password,
		users.MonthlyLimit(monthly_limit),
		isActive,
		createdAt,
		updatedAt,
	)
	return users, nil
}

func (u UserRepo) Update(ctx context.Context, user *users.User) error {
	const sql = `UPDATE users SET username=$1, email=$2, password=$3, monthly_limit=$4, is_active=$6, updated_at = $7 where id = $8`

	_, err := u.db.ExecContext(ctx, sql,
		user.Username(),
		user.Email(),
		user.Password(),
		user.MonthlyLimit(),
		user.IsActive(),
		time.Now(),
		user.UserID())
	if err != nil {
		return errors.New("cannot update users: " + err.Error())
	}
	return nil
}

func (u UserRepo) Delete(ctx context.Context, user *users.User) error {
	const sql = `DELETE FROM users WHERE id = $1`

	_, err := u.db.ExecContext(ctx, sql, user.UserID())
	if err != nil {
		return errors.New("cannot delete users: " + err.Error())
	}
	return nil
}
