package transport

import (
	"net/http"
	"time"
	service "walletwise/internal/application/categories"
)

type CategoryResponse struct {
	ID        uint64    `json:"id"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Icon      string    `json:"icon"`
	CreatedAt time.Time `json:"created_at"`
}

type CategoriesHandler struct {
	service *service.Service
}

func NewCategoriesHandler(s *service.Service) *CategoriesHandler {
	return &CategoriesHandler{service: s}
}

func (c *CategoriesHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := c.service.GetAllCategories(r.Context())
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var responses []CategoryResponse
	for _, cat := range categories {
		responses = append(responses, CategoryResponse{
			ID:   uint64(cat.ID()),
			Name: cat.Name(),
		})
	}

	if responses == nil {
		responses = []CategoryResponse{}
	}

	WriteJson(w, http.StatusOK, "Success Get Category List", responses)
}
