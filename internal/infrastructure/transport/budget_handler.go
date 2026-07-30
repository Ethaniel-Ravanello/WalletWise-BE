package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"walletwise/internal/middleware"

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
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	req.UserID = userId

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
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}

	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	b, err := h.svc.GetBudgetByID(r.Context(), budgetID, userId)
	if err != nil {
		WriteJSON(w, http.StatusNotFound, "Budget not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Budget retrieved successfully", b)
}

func (h *BudgetHandler) GetBudgetsByMonth(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	monthStr := queryParams.Get("month")
	yearStr := queryParams.Get("year")

	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}

	if userId == 0 || monthStr == "" || yearStr == "" {
		WriteJSON(w, http.StatusBadRequest, "user_id, month, and year query parameters are required", nil)
		return
	}

	month, err2 := strconv.Atoi(monthStr)
	year, err3 := strconv.Atoi(yearStr)

	if err2 != nil || err3 != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user_id, month, or year format", nil)
		return
	}

	budgets, err := h.svc.GetBudgetsByMonth(r.Context(), userId, month, year)
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
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}

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

	if err := h.svc.UpdateBudget(r.Context(), input, userId); err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Budget updated successfully", nil)
}

func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}

	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	if err := h.svc.DeleteBudget(r.Context(), budgetID, userId); err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Budget deleted successfully", nil)
}
