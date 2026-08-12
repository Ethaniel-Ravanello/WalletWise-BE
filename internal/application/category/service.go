package category

import (
	"context"
	"fmt"

	"walletwise/internal/domain/category"
)

type Service struct {
	repo category.Repository
}

func NewService(repo category.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

func (s *Service) GetAllCategories(ctx context.Context) ([]*category.Category, error) {
	allCategories, err := s.repo.SearchAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("get all categories: %w", err)
	}
	return allCategories, nil
}


