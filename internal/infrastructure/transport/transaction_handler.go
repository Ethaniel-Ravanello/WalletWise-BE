package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"walletwise/internal/domain/transaction"
	// Sesuaikan path import service Anda
	service "walletwise/internal/application/transaction"
)

// --- DTO / Response Structs ---

type TransactionResponse struct {
	ID              uint64    `json:"id"`
	UserID          uint64    `json:"user_id"`
	GoalID          *uint64   `json:"goal_id,omitempty"`
	Amount          int64     `json:"amount"`
	CategoryID      uint64    `json:"category_id"`
	Description     string    `json:"description"`
	TransactionType string    `json:"transaction_type"`
	WalletID        uint64    `json:"wallet_id"`
	TransactionDate time.Time `json:"transaction_date"`
}

type CreateTransactionRequest struct {
	UserID          uint64    `json:"user_id"`
	GoalID          uint64    `json:"goal_id"`
	Amount          int64     `json:"amount"`
	CategoryID      uint64    `json:"category_id"`
	Description     string    `json:"description"`
	TransactionType string    `json:"transaction_type"`
	WalletID        uint64    `json:"wallet_id"`
	Date            time.Time `json:"date"`
}

type UpdateTransactionRequest struct {
	GoalID          uint64    `json:"goal_id"`
	Amount          int64     `json:"amount"`
	CategoryID      uint64    `json:"category_id"`
	Description     string    `json:"description"`
	TransactionType string    `json:"transaction_type"`
	WalletID        uint64    `json:"wallet_id"`
	Date            time.Time `json:"date"`
}

type MonthlySummaryResponse struct {
	TotalIncome  int64 `json:"total_income"`
	TotalExpense int64 `json:"total_expense"`
}

type CategorySpendResponse struct {
	Category string `json:"category"`
	Total    int64  `json:"total"`
}

// --- Handler ---

type TransactionHandler struct {
	service *service.Service
}

func NewTransactionHandler(s *service.Service) *TransactionHandler {
	return &TransactionHandler{service: s}
}

// CreateTransaction menangani pembuatan transaksi baru (POST)
func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := &service.TrxInput{
		UserID:          req.UserID,
		GoalID:          req.GoalID,
		Amount:          req.Amount,
		CategoryID:      req.CategoryID,
		Description:     req.Description,
		TransactionType: req.TransactionType,
		WalletID:        req.WalletID,
		Date:            req.Date,
	}

	tx, err := h.service.CreateTransaction(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusCreated, "Success create transaction", toTransactionResponse(tx))
}

// GetTransactions mengambil daftar transaksi berdasarkan filter (GET)
func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	userID, _ := strconv.ParseUint(q.Get("user_id"), 10, 64)
	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 10 // Default limit
	}

	// Inisialisasi default pointer untuk menghindari Panic (Nil-Pointer Dereference) di Service
	var defaultZero uint64 = 0
	var defaultStr string = ""

	input := &service.GetTransactionsInput{
		UserID:          userID,
		Limit:           limit,
		GoalID:          &defaultZero,
		Amount:          &defaultZero,
		CategoryID:      &defaultZero,
		WalletID:        &defaultZero,
		TransactionType: &defaultStr,
	}

	// Parsing parameter opsional jika ada di query string
	if valStr := q.Get("goal_id"); valStr != "" {
		if val, err := strconv.ParseUint(valStr, 10, 64); err == nil {
			input.GoalID = &val
		}
	}
	if valStr := q.Get("amount"); valStr != "" {
		if val, err := strconv.ParseUint(valStr, 10, 64); err == nil {
			input.Amount = &val
		}
	}
	if valStr := q.Get("category_id"); valStr != "" {
		if val, err := strconv.ParseUint(valStr, 10, 64); err == nil {
			input.CategoryID = &val
		}
	}
	if valStr := q.Get("wallet_id"); valStr != "" {
		if val, err := strconv.ParseUint(valStr, 10, 64); err == nil {
			input.WalletID = &val
		}
	}
	if trxType := q.Get("transaction_type"); trxType != "" {
		input.TransactionType = &trxType
	}
	if startDateStr := q.Get("start_date"); startDateStr != "" {
		if val, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			input.StartDate = &val
		}
	}
	if endDateStr := q.Get("end_date"); endDateStr != "" {
		if val, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			input.EndDate = &val
		}
	}

	transactions, err := h.service.GetTransaction(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var responses []TransactionResponse
	for _, tx := range transactions {
		responses = append(responses, toTransactionResponse(tx))
	}

	if responses == nil {
		responses = []TransactionResponse{}
	}

	WriteJson(w, http.StatusOK, "Success get transactions", responses)
}

// GetTransactionById mengambil detail satu transaksi (GET)
func (h *TransactionHandler) GetTransactionById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	trxID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid transaction ID", nil)
		return
	}

	tx, err := h.service.GetTransactionById(r.Context(), transaction.TransactionID(trxID))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success get transaction", toTransactionResponse(tx))
}

// UpdateTransaction memperbarui data transaksi (PUT/PATCH)
func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	trxID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid transaction ID", nil)
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := &service.TrxUpdate{
		ID:              trxID,
		GoalID:          req.GoalID,
		Amount:          req.Amount,
		CategoryID:      req.CategoryID,
		Description:     req.Description,
		TransactionType: req.TransactionType,
		WalletID:        req.WalletID,
		Date:            req.Date,
	}

	if err := h.service.UpdateTransaction(r.Context(), input); err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success update transaction", nil)
}

// DeleteTransaction menghapus data transaksi (DELETE)
func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	trxID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid transaction ID", nil)
		return
	}

	if err := h.service.DeleteTransaction(r.Context(), transaction.TransactionID(trxID)); err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success delete transaction", nil)
}

// GetUserBalance mengambil total saldo dari spesifik Wallet User (GET)
func (h *TransactionHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseUint(r.URL.Query().Get("user_id"), 10, 64)
	walletID, _ := strconv.ParseUint(r.URL.Query().Get("wallet_id"), 10, 64)

	if userID == 0 || walletID == 0 {
		WriteJson(w, http.StatusBadRequest, "user_id and wallet_id are required", nil)
		return
	}

	balance, err := h.service.GetUserBalance(r.Context(), userID, walletID)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success get user balance", map[string]int64{
		"balance": int64(*balance),
	})
}

// GetMonthlySummary mengambil total Income & Expense per bulan (GET)
func (h *TransactionHandler) GetMonthlySummary(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseUint(r.URL.Query().Get("user_id"), 10, 64)
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))

	if userID == 0 || month == 0 || year == 0 {
		WriteJson(w, http.StatusBadRequest, "user_id, month, and year are required", nil)
		return
	}

	summary, err := h.service.GetMonthlySummary(r.Context(), userID, month, year)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	res := MonthlySummaryResponse{
		TotalIncome:  int64(summary.TotalIncome),
		TotalExpense: int64(summary.TotalExpense),
	}

	WriteJson(w, http.StatusOK, "Success get monthly summary", res)
}

// GetHighestExpense mengambil data pengeluaran tertinggi (GET)
func (h *TransactionHandler) GetHighestExpense(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userID, _ := strconv.ParseUint(q.Get("user_id"), 10, 64)
	month, _ := strconv.Atoi(q.Get("month"))
	year, _ := strconv.Atoi(q.Get("year"))
	limit, _ := strconv.Atoi(q.Get("limit"))

	if userID == 0 || month == 0 || year == 0 {
		WriteJson(w, http.StatusBadRequest, "user_id, month, and year are required", nil)
		return
	}

	if limit == 0 {
		limit = 1 // Default limit jika tidak dikirimkan
	}

	hiExpense, err := h.service.GetHighestExpense(r.Context(), userID, month, year, limit)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if hiExpense == nil {
		WriteJson(w, http.StatusOK, "No expense found", nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success get highest expense", toTransactionResponse(hiExpense))
}

// GetMostSpend mengambil daftar kategori dengan pengeluaran terbesar (GET)
func (h *TransactionHandler) GetMostSpend(w http.ResponseWriter, r *http.Request) {
	userID, _ := strconv.ParseUint(r.URL.Query().Get("user_id"), 10, 64)
	month, _ := strconv.Atoi(r.URL.Query().Get("month"))
	year, _ := strconv.Atoi(r.URL.Query().Get("year"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if limit == 0 {
		limit = 5 // default limit
	}

	if userID == 0 || month == 0 || year == 0 {
		WriteJson(w, http.StatusBadRequest, "user_id, month, and year are required", nil)
		return
	}

	spends, err := h.service.GetMostSpend(r.Context(), userID, month, year, limit)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var responses []CategorySpendResponse
	for _, spend := range spends {
		responses = append(responses, CategorySpendResponse{
			Category: spend.Category,
			Total:    spend.Total,
		})
	}

	if responses == nil {
		responses = []CategorySpendResponse{}
	}

	WriteJson(w, http.StatusOK, "Success get most spend categories", responses)
}

// --- Helper Functions ---

// toTransactionResponse memetakan Entity Transaction ke Struct Response API
func toTransactionResponse(tx *transaction.Transaction) TransactionResponse {
	var goalID *uint64
	if tx.GoalID() != nil {
		val := uint64(*tx.GoalID())
		goalID = &val
	}

	// Parsing ID dari string ke uint64 (berdasarkan method WalletID() di Entity lu)
	walletIDStr := tx.WalletID()
	walletID, _ := strconv.ParseUint(walletIDStr, 10, 64)

	return TransactionResponse{
		ID:              uint64(tx.ID()),
		UserID:          uint64(tx.UserID()),
		GoalID:          goalID,
		Amount:          int64(tx.Amount()),
		CategoryID:      uint64(tx.CategoryID()),
		Description:     tx.Description(),
		TransactionType: string(tx.TransactionType()),
		WalletID:        walletID,
		TransactionDate: tx.TransactionDate(),
	}
}
