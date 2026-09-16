package category

import (
	"context"
	"fmt"

	"go.uber.org/zap"

	"walletwise/internal/domain/category"
)

type Service struct {
	repo   category.Repository
	logger *zap.Logger
}

func NewService(repo category.Repository) *Service {
	return &Service{
		repo:   repo,
		logger: zap.L(),
	}
}

func (s *Service) GetAllCategories(ctx context.Context) ([]*category.Category, error) {
	s.logger.Debug("Fetching all categories from repository")

	allCategories, err := s.repo.SearchAll(ctx)
	if err != nil {
		s.logger.Error("Failed to get all categories", zap.Error(err))
		return nil, fmt.Errorf("get all categories: %w", err)
	}

	s.logger.Debug("Categories retrieved successfully", zap.Int("count", len(allCategories)))
	return allCategories, nil
}


