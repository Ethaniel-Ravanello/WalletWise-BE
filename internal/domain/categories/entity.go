package categories

import (
	"errors"
	"time"
)

type CategoriesID uint64

type Categories struct {
	id             CategoriesID
	name           string
	categoriesType string
	icon           string
	createdAt      time.Time
}

func NewCategories(
	name string,
	categoriesType string,
	icon string,
	createdAt time.Time,
) (*Categories, error) {
	if name == "" {
		return nil, errors.New("category name is required")
	}
	if categoriesType == "" {
		return nil, errors.New("category type is required")
	}
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return &Categories{
		name:           name,
		categoriesType: categoriesType,
		icon:           icon,
		createdAt:      createdAt,
	}, nil
}

func ReconstituteCategories(
	id CategoriesID,
	name string,
	categoriesType string,
	icon string,
	createdAt time.Time,
) *Categories {
	return &Categories{
		id:             id,
		name:           name,
		categoriesType: categoriesType,
		icon:           icon,
		createdAt:      createdAt,
	}
}

func (c *Categories) ID() CategoriesID       { return c.id }
func (c *Categories) Name() string           { return c.name }
func (c *Categories) CategoriesType() string { return c.categoriesType }
func (c *Categories) Icon() string           { return c.icon }
func (c *Categories) CreatedAt() time.Time   { return c.createdAt }

