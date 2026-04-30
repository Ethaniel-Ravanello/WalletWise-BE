package users

import (
	"errors"
	"time"
)

type UserID uint64
type MonthlyLimit uint64

type User struct {
	id           UserID       `json:"id"`
	username     string       `json:"username"`
	email        string       `json:"email"`
	password     string       `json:"password"`
	monthlyLimit MonthlyLimit `json:"monthly_limit"`
	isActive     bool         `json:"is_active"`
	createdAt    time.Time    `json:"created_at"`
	updatedAt    time.Time    `json:"updated_at"`
}

func NewUser(
	username string,
	email string,
	password string,
	monthlyLimit MonthlyLimit,
	isActive bool,
	createdAt time.Time,
	updatedAt time.Time) (*User, error) {

	if username == "" {
		return nil, errors.New("invalid users name")
	}
	if email == "" {
		return nil, errors.New("invalid users email")
	}
	if password == "" {
		return nil, errors.New("invalid users password")
	}
	if monthlyLimit <= 0 {
		return nil, errors.New("invalid users limit")
	}
	if createdAt.IsZero() {
		return nil, errors.New("invalid users createdAt")
	}
	if updatedAt.IsZero() {
		return nil, errors.New("invalid users updatedAt")
	}
	return &User{
		username:     username,
		email:        email,
		password:     password,
		monthlyLimit: monthlyLimit,
		isActive:     true,
		createdAt:    time.Now(),
		updatedAt:    time.Now(),
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
	updatedAt time.Time) *User {
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
		return errors.New("invalid users name")
	}
	if email == "" {
		return errors.New("invalid users email")
	}
	if isActive == u.isActive {
		return errors.New("User Status Cant Be The Same As Before")
	}
	if password == "" {
		return errors.New("invalid users password")
	}
	if password == u.password {
		return errors.New("User Password Cant Be The Same As Before")
	}
	if monthlyLimit <= 0 {
		return errors.New("invalid users limit")
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
