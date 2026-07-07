package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	// Sesuaikan path import lu ya
	service "walletwise/internal/application/budget"
	"walletwise/internal/domain/budget"
)

// --- DTO / Response Structs ---

type BudgetResponse struct {
	ID         uint64    `json:"id"`
	UserID     uint64    `json:"user_id"`
	CategoryID uint64    `json:"category_id"`
	Month      int       `json:"month"`
	Year       int       `json:"year"`
	Amount     int64     `json:"amount"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

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

// --- Handler ---

type BudgetHandler struct {
	service *service.Service
}

func NewBudgetHandler(s *service.Service) *BudgetHandler {
	return &BudgetHandler{service: s}
}

// CreateBudget menangani pembuatan budget baru (POST)
func (h *BudgetHandler) CreateBudget(w http.ResponseWriter, r *http.Request) {
	var req CreateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := service.BudgetInput{
		UserID:     req.UserID,
		CategoryID: req.CategoryID,
		Month:      req.Month,
		Year:       req.Year,
		Amount:     req.Amount,
	}

	b, err := h.service.CreateBudget(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusCreated, "Success create budget", toBudgetResponse(b))
}

// GetBudgetByID mengambil detail satu budget (GET)
func (h *BudgetHandler) GetBudgetByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	b, err := h.service.GetBudgetByID(r.Context(), budgetID)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success get budget", toBudgetResponse(b))
}

// GetBudgetsByMonth mengambil daftar budget user di bulan tertentu (GET)
func (h *BudgetHandler) GetBudgetsByMonth(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID, _ := strconv.ParseUint(q.Get("user_id"), 10, 64)
	month, _ := strconv.Atoi(q.Get("month"))
	year, _ := strconv.Atoi(q.Get("year"))

	if userID == 0 || month == 0 || year == 0 {
		WriteJson(w, http.StatusBadRequest, "user_id, month, and year query parameters are required", nil)
		return
	}

	budgets, err := h.service.GetBudgetsByMonth(r.Context(), userID, month, year)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var responses []BudgetResponse
	for _, b := range budgets {
		responses = append(responses, toBudgetResponse(b))
	}

	if responses == nil {
		responses = []BudgetResponse{}
	}

	WriteJson(w, http.StatusOK, "Success get budgets", responses)
}

// UpdateBudget memperbarui data budget (PUT/PATCH)
func (h *BudgetHandler) UpdateBudget(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	var req UpdateBudgetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := service.BudgetUpdateInput{
		ID:         budgetID,
		CategoryID: req.CategoryID,
		Month:      req.Month,
		Year:       req.Year,
		Amount:     req.Amount,
	}

	if err := h.service.UpdateBudget(r.Context(), input); err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success update budget", nil)
}

// DeleteBudget menghapus data budget (DELETE)
func (h *BudgetHandler) DeleteBudget(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	budgetID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid budget ID format", nil)
		return
	}

	if err := h.service.DeleteBudget(r.Context(), budgetID); err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success delete budget", nil)
}

// --- Helper Functions ---

// toBudgetResponse memetakan Entity Budget ke Struct Response API
func toBudgetResponse(b *budget.Budget) BudgetResponse {
	if b == nil {
		return BudgetResponse{}
	}
	return BudgetResponse{
		ID:         uint64(b.ID()),
		UserID:     uint64(b.UserID()),
		CategoryID: uint64(b.CategoryID()),
		Month:      b.Month(),
		Year:       b.Year(),
		Amount:     b.Amount(),
		CreatedAt:  b.CreatedAt(),
		UpdatedAt:  b.UpdatedAt(),
	}
}
