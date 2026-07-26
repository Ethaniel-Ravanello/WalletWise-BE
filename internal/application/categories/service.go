package categories

import (
	"context"
	"errors"
	"walletwise/internal/domain/categories"
)

type Service struct {
	repo categories.Repository
}

func NewService(repo categories.Repository) *Service { return &Service{repo: repo} }

func (c *Service) GetAllCategories(ctx context.Context) ([]*categories.Categories, error) {
	allCategories, err := c.repo.SearchAll(ctx)

	if err != nil {
		return nil, errors.New("cannot get all categories: " + err.Error())
	}
	return allCategories, nil
}
