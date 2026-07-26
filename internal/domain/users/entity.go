package users

import (
	"errors"
	"time"
)

type UserID uint64
type MonthlyLimit uint64

type User struct {
	id           UserID
	username     string
	email        string
	password     string
	monthlyLimit MonthlyLimit
	isActive     bool
	createdAt    time.Time
	updatedAt    time.Time
}

func NewUser(
	username string,
	email string,
	password string,
	monthlyLimit MonthlyLimit,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
) (*User, error) {

	if username == "" {
		return nil, errors.New("username is required")
	}
	if email == "" {
		return nil, errors.New("email is required")
	}
	if password == "" {
		return nil, errors.New("password is required")
	}
	if monthlyLimit == 0 {
		return nil, errors.New("monthly limit must be greater than zero")
	}
	if createdAt.IsZero() {
		createdAt = time.Now()
	}
	if updatedAt.IsZero() {
		updatedAt = time.Now()
	}

	return &User{
		username:     username,
		email:        email,
		password:     password,
		monthlyLimit: monthlyLimit,
		isActive:     isActive,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}, nil
}

func ReconstituteUser(
	id UserID,
	username string,
	email string,
	password string,
	monthlyLimit MonthlyLimit,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time,
) *User {
	return &User{
		id:           id,
		username:     username,
		email:        email,
		password:     password,
		monthlyLimit: monthlyLimit,
		isActive:     isActive,
		createdAt:    createdAt,
		updatedAt:    updatedAt,
	}
}

func (u *User) UpdateUser(
	username string,
	email string,
	password string,
	monthlyLimit MonthlyLimit,
	isActive bool,
) error {
	if username == "" {
		return errors.New("username is required")
	}
	if email == "" {
		return errors.New("email is required")
	}
	if password == "" {
		return errors.New("password is required")
	}
	if password == u.password {
		return errors.New("new password cannot be the same as old password")
	}
	if monthlyLimit == 0 {
		return errors.New("monthly limit must be greater than zero")
	}

	u.username = username
	u.email = email
	u.password = password
	u.isActive = isActive
	u.monthlyLimit = monthlyLimit
	u.updatedAt = time.Now()

	return nil
}

func (u *User) UserID() UserID             { return u.id }
func (u *User) Username() string           { return u.username }
func (u *User) Email() string              { return u.email }
func (u *User) Password() string           { return u.password }
func (u *User) MonthlyLimit() MonthlyLimit { return u.monthlyLimit }
func (u *User) IsActive() bool             { return u.isActive }
func (u *User) CreatedAt() time.Time       { return u.createdAt }
func (u *User) UpdatedAt() time.Time       { return u.updatedAt }

