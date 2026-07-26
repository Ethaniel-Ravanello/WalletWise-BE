package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	service "walletwise/internal/application/transaction"
	"walletwise/internal/domain/transaction"
)

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

type TransactionHandler struct {
	svc *service.Service
}

func NewTransactionHandler(svc *service.Service) *TransactionHandler {
	return &TransactionHandler{svc: svc}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.UserID == 0 || req.WalletID == 0 || req.Amount == 0 {
		WriteJSON(w, http.StatusBadRequest, "user_id, wallet_id, and amount are required", nil)
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

	tx, err := h.svc.CreateTransaction(r.Context(), input)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "Transaction created successfully", toTransactionResponse(tx))
}

func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	userIDStr := q.Get("user_id")
	if userIDStr == "" {
		userIDStr = q.Get("userId")
	}
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 10
	}

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

	transactions, err := h.svc.GetTransaction(r.Context(), input)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responses := make([]TransactionResponse, 0, len(transactions))
	for _, tx := range transactions {
		responses = append(responses, toTransactionResponse(tx))
	}

	WriteJSON(w, http.StatusOK, "Transactions retrieved successfully", responses)
}

func (h *TransactionHandler) GetTransactionById(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	trxID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid transaction ID format", nil)
		return
	}

	tx, err := h.svc.GetTransactionById(r.Context(), transaction.TransactionID(trxID))
	if err != nil {
		WriteJSON(w, http.StatusNotFound, "Transaction not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Transaction retrieved successfully", toTransactionResponse(tx))
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	trxID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid transaction ID format", nil)
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
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

	if err := h.svc.UpdateTransaction(r.Context(), input); err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Transaction updated successfully", nil)
}

func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	trxID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid transaction ID format", nil)
		return
	}

	if err := h.svc.DeleteTransaction(r.Context(), transaction.TransactionID(trxID)); err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Transaction deleted successfully", nil)
}

func (h *TransactionHandler) GetUserBalance(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userIDStr := q.Get("user_id")
	if userIDStr == "" {
		userIDStr = q.Get("userId")
	}
	walletIDStr := q.Get("wallet_id")
	if walletIDStr == "" {
		walletIDStr = q.Get("walletId")
	}

	userID, err1 := strconv.ParseUint(userIDStr, 10, 64)
	walletID, err2 := strconv.ParseUint(walletIDStr, 10, 64)

	if err1 != nil || err2 != nil || userID == 0 || walletID == 0 {
		WriteJSON(w, http.StatusBadRequest, "user_id and wallet_id are required", nil)
		return
	}

	balance, err := h.svc.GetUserBalance(r.Context(), userID, walletID)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	res := map[string]int64{
		"balance": int64(*balance),
	}

	WriteJSON(w, http.StatusOK, "User balance retrieved successfully", res)
}

func (h *TransactionHandler) GetMonthlySummary(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userIDStr := q.Get("user_id")
	if userIDStr == "" {
		userIDStr = q.Get("userId")
	}
	monthStr := q.Get("month")
	yearStr := q.Get("year")

	userID, err1 := strconv.ParseUint(userIDStr, 10, 64)
	month, err2 := strconv.Atoi(monthStr)
	year, err3 := strconv.Atoi(yearStr)

	if err1 != nil || err2 != nil || err3 != nil || userID == 0 || month == 0 || year == 0 {
		WriteJSON(w, http.StatusBadRequest, "user_id, month, and year are required", nil)
		return
	}

	summary, err := h.svc.GetMonthlySummary(r.Context(), userID, month, year)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	res := MonthlySummaryResponse{
		TotalIncome:  int64(summary.TotalIncome),
		TotalExpense: int64(summary.TotalExpense),
	}

	WriteJSON(w, http.StatusOK, "Monthly summary retrieved successfully", res)
}

func (h *TransactionHandler) GetHighestExpense(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userIDStr := q.Get("user_id")
	if userIDStr == "" {
		userIDStr = q.Get("userId")
	}
	monthStr := q.Get("month")
	yearStr := q.Get("year")

	userID, err1 := strconv.ParseUint(userIDStr, 10, 64)
	month, err2 := strconv.Atoi(monthStr)
	year, err3 := strconv.Atoi(yearStr)

	if err1 != nil || err2 != nil || err3 != nil || userID == 0 || month == 0 || year == 0 {
		WriteJSON(w, http.StatusBadRequest, "user_id, month, and year are required", nil)
		return
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 1
	}

	hiExpense, err := h.svc.GetHighestExpense(r.Context(), userID, month, year, limit)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if hiExpense == nil {
		WriteJSON(w, http.StatusOK, "No expense found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Highest expense retrieved successfully", toTransactionResponse(hiExpense))
}

func (h *TransactionHandler) GetMostSpend(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userIDStr := q.Get("user_id")
	if userIDStr == "" {
		userIDStr = q.Get("userId")
	}
	monthStr := q.Get("month")
	yearStr := q.Get("year")

	userID, err1 := strconv.ParseUint(userIDStr, 10, 64)
	month, err2 := strconv.Atoi(monthStr)
	year, err3 := strconv.Atoi(yearStr)

	if err1 != nil || err2 != nil || err3 != nil || userID == 0 || month == 0 || year == 0 {
		WriteJSON(w, http.StatusBadRequest, "user_id, month, and year are required", nil)
		return
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 5
	}

	spends, err := h.svc.GetMostSpend(r.Context(), userID, month, year, limit)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responses := make([]CategorySpendResponse, 0, len(spends))
	for _, spend := range spends {
		responses = append(responses, CategorySpendResponse{
			Category: spend.Category,
			Total:    spend.Total,
		})
	}

	WriteJSON(w, http.StatusOK, "Most spend categories retrieved successfully", responses)
}

func toTransactionResponse(tx *transaction.Transaction) TransactionResponse {
	if tx == nil {
		return TransactionResponse{}
	}

	var goalID *uint64
	if tx.GoalID() != nil {
		val := uint64(*tx.GoalID())
		goalID = &val
	}

	return TransactionResponse{
		ID:              uint64(tx.ID()),
		UserID:          uint64(tx.UserID()),
		GoalID:          goalID,
		Amount:          int64(tx.Amount()),
		CategoryID:      uint64(tx.CategoryID()),
		Description:     tx.Description(),
		TransactionType: string(tx.TransactionType()),
		WalletID:        uint64(tx.WalletID()),
		TransactionDate: tx.TransactionDate(),
	}
}

