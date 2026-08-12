package transport

import (
	"net/http"
	"time"

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
	svc *service.Service
}

type CategoriesHandler = CategoryHandler

func NewCategoryHandler(svc *service.Service) *CategoryHandler {
	return &CategoryHandler{svc: svc}
}

func (h *CategoryHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categoriesList, err := h.svc.GetAllCategories(r.Context())
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responses := make([]CategoryResponse, 0, len(categoriesList))
	for _, cat := range categoriesList {
		responses = append(responses, toCategoryResponse(cat))
	}

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
