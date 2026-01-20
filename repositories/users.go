package repositories

import (
	"context"
	"database/sql"
	"errors"
	"walletwise/models"
)

type UserRepository struct {
	db *sql.DB
}

func (r *UserRepository) GetUserById(ctx context.Context, id int) (*models.User, error) {
	const query = `
	SELECT id, username, email, password, monthly_limit, is_active, created_at, updated_at
	FROM users
	WHERE id = $1
`

	var user models.User

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID,
		&user.Username,
		&user.Email,
		&user.Password,
		&user.MonthlyLimit,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepository) CreateUser(ctx context.Context, userData *models.User) error {
	const query = `
		INSERT INTO users
		(username, email, password, monthly_limit, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at
		`

	return r.db.QueryRowContext(
		ctx,
		query,
		userData.Username,
		userData.Email,
		userData.Password,
		userData.MonthlyLimit,
		userData.IsActive,
		userData.CreatedAt,
		userData.UpdatedAt,
	).Scan(&userData.ID, &userData.CreatedAt, &userData.UpdatedAt)
}

func (r *UserRepository) UpdateUserById(ctx context.Context, id int, userData *models.User) error {
	const query = `
		UPDATE users
		SET 
		    username = $1,
		    email = $2,
		    password = $3,
		    monthly_limit = $4,
		    is_active = $5,
		    updated_at = $6
		WHERE id = $7
	`

	res, err := r.db.ExecContext(ctx, query,
		userData.Username,
		userData.Email,
		userData.Password,
		userData.MonthlyLimit,
		userData.IsActive,
		userData.CreatedAt,
		userData.UpdatedAt,
		id,
	)

	if err != nil {
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("No rows affected")
	}
	return err
}

func (r *UserRepository) DeleteUserById(id int) error {
	const query = `DELETE FROM users WHERE id = $1`

	res, err := r.db.Exec(query, id)

	if err != nil {
		return err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return errors.New("No rows affected")
	}
	return err
}
