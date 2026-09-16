package transport

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	service "walletwise/internal/application/category"
	"walletwise/internal/domain/category"
)

type CategoryResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"created_at"`
}

type CategoryHandler struct {
	svc    *service.Service
	logger *zap.Logger
}

type CategoriesHandler = CategoryHandler

func NewCategoryHandler(svc *service.Service) *CategoryHandler {
	return &CategoryHandler{
		svc:    svc,
		logger: zap.L(),
	}
}

func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	h.logger.Debug("Fetching all categories")

	categoriesList, err := h.svc.GetAllCategories(r.Context())
	if err != nil {
		h.logger.Error("Failed to get categories", zap.Error(err))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responses := make([]CategoryResponse, 0, len(categoriesList))
	for _, cat := range categoriesList {
		responses = append(responses, toCategoryResponse(cat))
	}

	h.logger.Debug("Categories retrieved successfully", zap.Int("count", len(responses)))
	WriteJSON(w, http.StatusOK, "Categories retrieved successfully", responses)
}

func toCategoryResponse(cat *category.Category) CategoryResponse {
	if cat == nil {
		return CategoryResponse{}
	}
	return CategoryResponse{
		ID:        uint64(cat.ID()),
		Name:      cat.Name(),
		Type:      cat.CategoryType(),
		Icon:      cat.Icon(),
		CreatedAt: cat.CreatedAt(),
	}
}
