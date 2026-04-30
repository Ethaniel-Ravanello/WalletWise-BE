package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"
	"walletwise/internal/domain/categories"
)

type CategoriesRepo struct {
	db *sql.DB
}

func NewCategoriesRepo(db *sql.DB) *CategoriesRepo { return &CategoriesRepo{db: db} }

var _ categories.Repository = (*CategoriesRepo)(nil)

func (c CategoriesRepo) SearchAll(ctx context.Context) ([]*categories.Categories, error) {
	const sql = `SELECT id, name, type, icon, created_at FROM categories`

	rows, err := c.db.QueryContext(ctx, sql)
	if err != nil {
		return nil, errors.New("cannot find user: " + err.Error())
	}
	defer rows.Close()
	var newCategories []*categories.Categories
	var (
		id             uint64
		name           string
		categoriesType string
		icon           string
		createdAt      time.Time
	)
	for rows.Next() {
		if err := rows.Scan(&id, &name, &categoriesType, &icon, &createdAt); err != nil {
			return nil, errors.New("cannot scan row: " + err.Error())
		}
	}
	ctgrs := categories.ReconstituteCategories(
		categories.CategoriesID(id),
		name,
		categoriesType,
		icon,
		createdAt)
	newCategories = append(newCategories, ctgrs)
	return newCategories, nil
}
