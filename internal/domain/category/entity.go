package category

import (
	"errors"
	"time"
)

type CategoryID uint64

type Category struct {
	id           CategoryID
	name         string
	categoryType string
	icon         string
	createdAt    time.Time
}

func NewCategory(
	name string,
	categoryType string,
	icon string,
	createdAt time.Time,
) (*Category, error) {
	if name == "" {
		return nil, errors.New("category name is required")
	}
	if categoryType == "" {
		return nil, errors.New("category type is required")
	}
	if createdAt.IsZero() {
		createdAt = time.Now()
	}

	return &Category{
		name:         name,
		categoryType: categoryType,
		icon:         icon,
		createdAt:    createdAt,
	}, nil
}

func ReconstituteCategory(
	id CategoryID,
	name string,
	categoryType string,
	icon string,
	createdAt time.Time,
) *Category {
	return &Category{
		id:           id,
		name:         name,
		categoryType: categoryType,
		icon:         icon,
		createdAt:    createdAt,
	}
}

func (c *Category) ID() CategoryID           { return c.id }
func (c *Category) Name() string             { return c.name }
func (c *Category) CategoryType() string     { return c.categoryType }
func (c *Category) CategoriesType() string   { return c.categoryType } // Backwards-compatible alias if needed
func (c *Category) Icon() string             { return c.icon }
func (c *Category) CreatedAt() time.Time     { return c.createdAt }


