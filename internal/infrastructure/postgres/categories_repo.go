package postgres

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"walletwise/internal/domain/category"
)

type CategoriesRepo struct {
	db *sql.DB
}

func NewCategoriesRepo(db *sql.DB) *CategoriesRepo {
	return &CategoriesRepo{db: db}
}

var _ category.Repository = (*CategoriesRepo)(nil)

func (r *CategoriesRepo) SearchAll(ctx context.Context) ([]*category.Category, error) {
	query := `SELECT id, name, type, icon, created_at FROM categories`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query categories: %w", err)
	}
	defer rows.Close()

	var categoriesList []*category.Category
	for rows.Next() {
		var (
			id           uint64
			name         string
			categoryType string
			icon         string
			createdAt    time.Time
		)

		if err := rows.Scan(&id, &name, &categoryType, &icon, &createdAt); err != nil {
			return nil, fmt.Errorf("failed to scan category row: %w", err)
		}

		cat := category.ReconstituteCategory(
			category.CategoryID(id),
			name,
			categoryType,
			icon,
			createdAt,
		)
		categoriesList = append(categoriesList, cat)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating category rows: %w", err)
	}

	return categoriesList, nil
}

