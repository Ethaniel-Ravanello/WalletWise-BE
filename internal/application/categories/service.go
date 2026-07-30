package categories

import (
	"context"
	"fmt"

	"walletwise/internal/domain/categories"
)

type Service struct {
	repo categories.Repository
}

func NewService(repo categories.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetAllCategories(ctx context.Context) ([]*categories.Categories, error) {
	allCategories, err := s.repo.SearchAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all categories: %w", err)
	}
	return allCategories, nil
}

