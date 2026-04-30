package transport

import (
	"net/http"
	service "walletwise/internal/application/categories"
)

type CategoriesHandler struct {
	service *service.Service
}

func NewCategoriesHandler(s *service.Service) *CategoriesHandler {
	return &CategoriesHandler{service: s}
}

func (c *CategoriesHandler) GetAllCategories(w http.ResponseWriter, r *http.Request) {
	categories, err := c.service.GetAllCategories(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	WriteJson(w, http.StatusOK, "Success Get Category List", categories)
}
