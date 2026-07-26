package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	service "walletwise/internal/application/budget"
)

type CreateBudgetRequest struct {
	UserID     uint64 `json:"user_id"`
	CategoryID uint64 `json:"category_id"`
	Month      int    `json:"month"`
	Year       int    `json:"year"`
	Amount     int64  `json:"amount"`
}

type UpdateBudgetRequest struct {
	CategoryID uint64 `json:"category_id"`
	Month      int    `json:"month"`
	Year       int    `json:"year"`
	Amount     int64  `json:"amount"`
}

type BudgetHandler struct {
	svc *service.Service
}

func NewBudgetHandler(svc *service.Service) *BudgetHandler {
	return &BudgetHandler{svc: svc}
}

func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.UserID == 0 || req.CategoryID == 0 || req.Month == 0 || req.Year == 0 {
		WriteJSON(w, http.StatusBadRequest, "user_id, category_id, month, and year are required", nil)
		return
	}

	input := service.BudgetInput{
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
		Month:      req.Month,
		Year:       req.Year,
		Amount:     req.Amount,
	}

	b, err := h.svc.CreateBudget(r.Context(), input)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "Budget created successfully", b)
}

func (h *BudgetHandler) GetBudgetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	b, err := h.svc.GetBudgetByID(r.Context(), budgetID)
	if err != nil {
		WriteJSON(w, http.StatusNotFound, "Budget not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Budget retrieved successfully", b)
}

func (h *BudgetHandler) GetBudgetsByMonth(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	userIDStr := queryParams.Get("user_id")
	monthStr := queryParams.Get("month")
	yearStr := queryParams.Get("year")

	if userIDStr == "" || monthStr == "" || yearStr == "" {
		WriteJSON(w, http.StatusBadRequest, "user_id, month, and year query parameters are required", nil)
		return
	}

	userID, err1 := strconv.ParseUint(userIDStr, 10, 64)
	month, err2 := strconv.Atoi(monthStr)
	year, err3 := strconv.Atoi(yearStr)

	if err1 != nil || err2 != nil || err3 != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user_id, month, or year format", nil)
		return
	}

	budgets, err := h.svc.GetBudgetsByMonth(r.Context(), userID, month, year)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if budgets == nil {
		budgets = []*service.BudgetDetailResponse{}
	}

	WriteJSON(w, http.StatusOK, "Budgets retrieved successfully", budgets)
}

func (h *BudgetHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	var req UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := service.BudgetUpdateInput{
		ID:         budgetID,
		CategoryID: req.CategoryID,
		Month:      req.Month,
		Year:       req.Year,
		Amount:     req.Amount,
	}

	if err := h.svc.UpdateBudget(r.Context(), input); err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Budget updated successfully", nil)
}

func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	if err := h.svc.DeleteBudget(r.Context(), budgetID); err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Budget deleted successfully", nil)
}


