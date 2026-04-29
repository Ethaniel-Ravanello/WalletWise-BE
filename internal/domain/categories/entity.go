package categories

import "time"

type CategoriesID uint64

type Categories struct {
	id             CategoriesID
	name           string
	categoriesType string
	icon           string
	createdAt      time.Time
}

func ReconstituteCategories(
	id CategoriesID,
	name string,
	categoriesType string,
	icon string,
	createdAt time.Time) *Categories {
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
