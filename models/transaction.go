package models

import "time"

type Transaction struct {
	ID          uint      `json:"id"`
	UserId      uint      `json:"user_id"`
	Amount      float64   `json:"amount"`
	Category    string    `json:"category"`
	Description string    `json:"description"`
	Type        string    `json:"type"`
	Source      string    `json:"source"`
	Date        time.Time `json:"date"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
