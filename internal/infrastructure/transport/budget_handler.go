package transport

import (
	"encoding/json"
	"net/http"
	"strconv"

	"go.uber.org/zap"

	service "walletwise/internal/application/budget"
	"walletwise/internal/middleware"
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
	svc    *service.Service
	logger *zap.Logger
}

func NewBudgetHandler(svc *service.Service) *BudgetHandler {
	return &BudgetHandler{
		svc:    svc,
		logger: zap.L(),
	}
}

func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized create budget: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode create budget request",
			zap.Error(err),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	req.UserID = userId

	if req.UserID == 0 || req.CategoryID == 0 || req.Month == 0 || req.Year == 0 {
		h.logger.Warn("Create budget validation failed: missing required fields",
			zap.Uint64("UserId", req.UserID),
			zap.Uint64("CategoryId", req.CategoryID),
			zap.Int("Month", req.Month),
			zap.Int("Year", req.Year))
		WriteJSON(w, http.StatusBadRequest, "user_id, category_id, month, and year are required", nil)
		return
	}

	h.logger.Debug("Creating budget",
		zap.Uint64("UserId", req.UserID),
		zap.Uint64("CategoryId", req.CategoryID),
		zap.Int("Month", req.Month),
		zap.Int("Year", req.Year),
		zap.Int64("Amount", req.Amount))

	input := service.BudgetInput{
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
		Month:      req.Month,
		Year:       req.Year,
		Amount:     req.Amount,
	}

	b, err := h.svc.CreateBudget(r.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create budget",
			zap.Error(err),
			zap.Uint64("UserId", req.UserID),
			zap.Uint64("CategoryId", req.CategoryID),
			zap.Int("Month", req.Month),
			zap.Int("Year", req.Year))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Budget created successfully",
		zap.Uint64("UserId", req.UserID),
		zap.Uint64("CategoryId", req.CategoryID),
		zap.Int("Month", req.Month),
		zap.Int("Year", req.Year),
		zap.Int64("Amount", req.Amount))
	WriteJSON(w, http.StatusCreated, "Budget created successfully", b)
}

func (h *BudgetHandler) GetBudgetByID(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized get budget by id: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid budget ID format",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	h.logger.Debug("Fetching budget by ID",
		zap.Uint64("BudgetId", budgetID),
		zap.Uint64("UserId", userId))

	b, err := h.svc.GetBudgetByID(r.Context(), budgetID, userId)
	if err != nil {
		h.logger.Warn("Budget not found",
			zap.Error(err),
			zap.Uint64("BudgetId", budgetID),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusNotFound, "Budget not found", nil)
		return
	}

	h.logger.Debug("Budget retrieved successfully",
		zap.Uint64("BudgetId", budgetID),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "Budget retrieved successfully", b)
}

func (h *BudgetHandler) GetBudgetsByMonth(w http.ResponseWriter, r *http.Request) {
	queryParams := r.URL.Query()

	monthStr := queryParams.Get("month")
	yearStr := queryParams.Get("year")

	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized get budgets by month: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	if userId == 0 || monthStr == "" || yearStr == "" {
		h.logger.Warn("Missing required query parameters for budgets by month",
			zap.Uint64("UserId", userId),
			zap.String("month", monthStr),
			zap.String("year", yearStr))
		WriteJSON(w, http.StatusBadRequest, "user_id, month, and year query parameters are required", nil)
		return
	}

	month, err2 := strconv.Atoi(monthStr)
	year, err3 := strconv.Atoi(yearStr)

	if err2 != nil || err3 != nil {
		h.logger.Warn("Invalid month or year format",
			zap.String("month", monthStr),
			zap.String("year", yearStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid user_id, month, or year format", nil)
		return
	}

	h.logger.Debug("Fetching budgets by month",
		zap.Uint64("UserId", userId),
		zap.Int("month", month),
		zap.Int("year", year))

	budgets, err := h.svc.GetBudgetsByMonth(r.Context(), userId, month, year)
	if err != nil {
		h.logger.Error("Failed to get budgets by month",
			zap.Error(err),
			zap.Uint64("UserId", userId),
			zap.Int("month", month),
			zap.Int("year", year))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if budgets == nil {
		budgets = []*service.BudgetDetailResponse{}
	}

	h.logger.Debug("Budgets retrieved successfully",
		zap.Uint64("UserId", userId),
		zap.Int("count", len(budgets)))
	WriteJSON(w, http.StatusOK, "Budgets retrieved successfully", budgets)
}

func (h *BudgetHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized update budget: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid budget ID format for update",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	var req UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode update budget request",
			zap.Error(err),
			zap.Uint64("BudgetId", budgetID),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	h.logger.Debug("Updating budget",
		zap.Uint64("BudgetId", budgetID),
		zap.Uint64("UserId", userId))

	input := service.BudgetUpdateInput{
		ID:         budgetID,
		CategoryID: req.CategoryID,
		Month:      req.Month,
		Year:       req.Year,
		Amount:     req.Amount,
	}

	if err := h.svc.UpdateBudget(r.Context(), input, userId); err != nil {
		h.logger.Error("Failed to update budget",
			zap.Error(err),
			zap.Uint64("BudgetId", budgetID),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Budget updated successfully",
		zap.Uint64("BudgetId", budgetID),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "Budget updated successfully", nil)
}

func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized delete budget: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid budget ID format for deletion",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	h.logger.Debug("Deleting budget",
		zap.Uint64("BudgetId", budgetID),
		zap.Uint64("UserId", userId))

	if err := h.svc.DeleteBudget(r.Context(), budgetID, userId); err != nil {
		h.logger.Error("Failed to delete budget",
			zap.Error(err),
			zap.Uint64("BudgetId", budgetID),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Budget deleted successfully",
		zap.Uint64("BudgetId", budgetID),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "Budget deleted successfully", nil)
}
