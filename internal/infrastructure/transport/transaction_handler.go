package transport

import (
	"encoding/json"
	"math"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	service "walletwise/internal/application/transaction"
	"walletwise/internal/domain/transaction"
	"walletwise/internal/middleware"
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
	svc    *service.Service
	logger *zap.Logger
}

func NewTransactionHandler(svc *service.Service) *TransactionHandler {
	return &TransactionHandler{
		svc:    svc,
		logger: zap.L(),
	}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var req CreateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode create transaction request", zap.Error(err))
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.UserID == 0 {
		userIdCtx := r.Context().Value(middleware.UserIdKey)
		if userId, ok := userIdCtx.(uint64); ok && userId > 0 {
			req.UserID = userId
		}
	}

	if req.UserID == 0 || req.WalletID == 0 || req.Amount == 0 {
		h.logger.Warn("Create transaction validation failed: missing required fields",
			zap.Uint64("UserId", req.UserID),
			zap.Uint64("WalletId", req.WalletID),
			zap.Int64("Amount", req.Amount))
		WriteJSON(w, http.StatusBadRequest, "user_id, wallet_id, and amount are required", nil)
		return
	}

	h.logger.Debug("Creating transaction",
		zap.Uint64("UserId", req.UserID),
		zap.Uint64("WalletId", req.WalletID),
		zap.Int64("Amount", req.Amount),
		zap.String("Type", req.TransactionType))

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
		h.logger.Error("Failed to create transaction",
			zap.Error(err),
			zap.Uint64("UserId", req.UserID),
			zap.Uint64("WalletId", req.WalletID))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Transaction created successfully",
		zap.Uint64("TransactionId", uint64(tx.ID())),
		zap.Uint64("UserId", uint64(tx.UserID())),
		zap.Int64("Amount", int64(tx.Amount())),
		zap.String("Type", string(tx.TransactionType())))
	WriteJSON(w, http.StatusCreated, "Transaction created successfully", toTransactionResponse(tx))
}

func (h *TransactionHandler) GetTransactions(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	limit := q.Get("limit")
	intLimit, err := strconv.Atoi(limit)
	if err != nil {
		h.logger.Warn("Invalid limit parameter", zap.Error(err), zap.String("limit", limit))
		WriteJSON(w, http.StatusBadRequest, "Invalid limit parameter", nil)
		return
	}
	page := q.Get("page")
	if page == "" {
		page = q.Get("limit")
	}
	intPage, err := strconv.Atoi(page)
	if err != nil {
		h.logger.Warn("Invalid page parameter", zap.Error(err), zap.String("page", page))
		WriteJSON(w, http.StatusBadRequest, "Invalid page parameter", nil)
		return
	}

	if intLimit > 100 {
		intLimit = 100
	}

	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized get transactions: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	h.logger.Debug("Fetching transactions",
		zap.Uint64("UserId", userId),
		zap.Int("limit", intLimit),
		zap.Int("page", intPage))

	input := service.GetTransactionsInput{
		UserID:          userId,
		GoalID:          nil,
		Amount:          0,
		CategoryID:      0,
		WalletID:        0,
		TransactionType: "",
		Limit:           intLimit,
		Page:            intPage,
	}

	if valStr := q.Get("goal_id"); valStr != "" {
		if val, err := strconv.ParseUint(valStr, 10, 64); err == nil {
			input.GoalID = &val
		}
	}
	if valStr := q.Get("amount"); valStr != "" {
		if val, err := strconv.ParseUint(valStr, 10, 64); err == nil {
			input.Amount = val
		}
	}
	if valStr := q.Get("category_id"); valStr != "" {
		if val, err := strconv.ParseUint(valStr, 10, 64); err == nil {
			input.CategoryID = val
		}
	}
	if valStr := q.Get("wallet_id"); valStr != "" {
		if val, err := strconv.ParseUint(valStr, 10, 64); err == nil {
			input.WalletID = val
		}
	}
	if trxType := q.Get("transaction_type"); trxType != "" {
		input.TransactionType = trxType
	}
	if startDateStr := q.Get("start_date"); startDateStr != "" {
		if val, err := time.Parse(time.RFC3339, startDateStr); err == nil {
			input.StartDate = val
		}
	}
	if endDateStr := q.Get("end_date"); endDateStr != "" {
		if val, err := time.Parse(time.RFC3339, endDateStr); err == nil {
			input.EndDate = val
		}
	}

	transactions, totalData, err := h.svc.GetTransaction(r.Context(), input)
	if err != nil {
		h.logger.Error("Failed to get transactions",
			zap.Error(err),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responses := make([]TransactionResponse, 0, len(transactions))
	for _, tx := range transactions {
		responses = append(responses, toTransactionResponse(tx))
	}

	totalPages := int(math.Ceil(float64(totalData) / float64(intLimit)))
	responseData := map[string]interface{}{
		"data": responses,
		"meta": map[string]interface{}{
			"current_page": page,
			"limit":        intLimit,
			"total_items":  totalData,
			"total_pages":  totalPages,
		},
	}

	h.logger.Debug("Transactions retrieved successfully",
		zap.Uint64("UserId", userId),
		zap.Int("count", len(responses)),
		zap.Int("total_items", totalData))
	WriteJSON(w, http.StatusOK, "Transactions retrieved successfully", responseData)
}

func (h *TransactionHandler) GetTransactionByID(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized get transaction by ID: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	trxID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid transaction ID format",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid transaction ID format", nil)
		return
	}

	h.logger.Debug("Fetching transaction by ID",
		zap.Uint64("TransactionId", trxID),
		zap.Uint64("UserId", userId))

	tx, err := h.svc.GetTransactionByID(r.Context(), transaction.TransactionID(trxID), transaction.UserID(userId))
	if err != nil {
		h.logger.Warn("Transaction not found",
			zap.Error(err),
			zap.Uint64("TransactionId", trxID),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusNotFound, "Transaction not found", nil)
		return
	}

	h.logger.Debug("Transaction retrieved successfully",
		zap.Uint64("TransactionId", trxID),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "Transaction retrieved successfully", toTransactionResponse(tx))
}

func (h *TransactionHandler) GetTransactionById(w http.ResponseWriter, r *http.Request) {
	h.GetTransactionByID(w, r)
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized update transaction: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	trxID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid transaction ID format for update",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid transaction ID format", nil)
		return
	}

	var req UpdateTransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode update transaction request",
			zap.Error(err),
			zap.Uint64("TransactionId", trxID),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	h.logger.Debug("Updating transaction",
		zap.Uint64("TransactionId", trxID),
		zap.Uint64("UserId", userId))

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

	if err := h.svc.UpdateTransaction(r.Context(), input, transaction.UserID(userId)); err != nil {
		h.logger.Error("Failed to update transaction",
			zap.Error(err),
			zap.Uint64("TransactionId", trxID),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Transaction updated successfully",
		zap.Uint64("TransactionId", trxID),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "Transaction updated successfully", nil)
}

func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized delete transaction: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	trxID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid transaction ID format for deletion",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid transaction ID format", nil)
		return
	}

	h.logger.Debug("Deleting transaction",
		zap.Uint64("TransactionId", trxID),
		zap.Uint64("UserId", userId))

	if err := h.svc.DeleteTransaction(r.Context(), transaction.TransactionID(trxID), transaction.UserID(userId)); err != nil {
		h.logger.Error("Failed to delete transaction",
			zap.Error(err),
			zap.Uint64("TransactionId", trxID),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Transaction deleted successfully",
		zap.Uint64("TransactionId", trxID),
		zap.Uint64("UserId", userId))
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
		h.logger.Warn("Missing or invalid user_id or wallet_id for balance query",
			zap.String("userId", userIDStr),
			zap.String("walletId", walletIDStr))
		WriteJSON(w, http.StatusBadRequest, "user_id and wallet_id are required", nil)
		return
	}

	h.logger.Debug("Fetching user balance",
		zap.Uint64("UserId", userID),
		zap.Uint64("WalletId", walletID))

	balance, err := h.svc.GetUserBalance(r.Context(), userID, walletID)
	if err != nil {
		h.logger.Error("Failed to get user balance",
			zap.Error(err),
			zap.Uint64("UserId", userID),
			zap.Uint64("WalletId", walletID))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	res := map[string]int64{
		"balance": int64(*balance),
	}

	h.logger.Debug("User balance retrieved successfully",
		zap.Uint64("UserId", userID),
		zap.Uint64("WalletId", walletID))
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
		h.logger.Warn("Missing or invalid parameters for monthly summary",
			zap.String("userId", userIDStr),
			zap.String("month", monthStr),
			zap.String("year", yearStr))
		WriteJSON(w, http.StatusBadRequest, "user_id, month, and year are required", nil)
		return
	}

	h.logger.Debug("Fetching monthly summary",
		zap.Uint64("UserId", userID),
		zap.Int("month", month),
		zap.Int("year", year))

	summary, err := h.svc.GetMonthlySummary(r.Context(), userID, month, year)
	if err != nil {
		h.logger.Error("Failed to get monthly summary",
			zap.Error(err),
			zap.Uint64("UserId", userID),
			zap.Int("month", month),
			zap.Int("year", year))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	res := MonthlySummaryResponse{
		TotalIncome:  int64(summary.TotalIncome),
		TotalExpense: int64(summary.TotalExpense),
	}

	h.logger.Debug("Monthly summary retrieved successfully",
		zap.Uint64("UserId", userID),
		zap.Int("month", month),
		zap.Int("year", year))
	WriteJSON(w, http.StatusOK, "Monthly summary retrieved successfully", res)
}

func (h *TransactionHandler) GetHighestExpense(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized get highest expense: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}
	monthStr := q.Get("month")
	yearStr := q.Get("year")

	month, err2 := strconv.Atoi(monthStr)
	year, err3 := strconv.Atoi(yearStr)

	if err2 != nil || err3 != nil || userId == 0 || month == 0 || year == 0 {
		h.logger.Warn("Missing or invalid parameters for highest expense",
			zap.Uint64("UserId", userId),
			zap.String("month", monthStr),
			zap.String("year", yearStr))
		WriteJSON(w, http.StatusBadRequest, "user_id, month, and year are required", nil)
		return
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 1
	}

	h.logger.Debug("Fetching highest expense",
		zap.Uint64("UserId", userId),
		zap.Int("month", month),
		zap.Int("year", year),
		zap.Int("limit", limit))

	hiExpense, err := h.svc.GetHighestExpense(r.Context(), userId, month, year, limit)
	if err != nil {
		h.logger.Error("Failed to get highest expense",
			zap.Error(err),
			zap.Uint64("UserId", userId),
			zap.Int("month", month),
			zap.Int("year", year))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	if hiExpense == nil {
		h.logger.Debug("No highest expense found",
			zap.Uint64("UserId", userId),
			zap.Int("month", month),
			zap.Int("year", year))
		WriteJSON(w, http.StatusOK, "No expense found", nil)
		return
	}

	h.logger.Debug("Highest expense retrieved successfully",
		zap.Uint64("UserId", userId),
		zap.Int("month", month),
		zap.Int("year", year))
	WriteJSON(w, http.StatusOK, "Highest expense retrieved successfully", toTransactionResponse(hiExpense))
}

func (h *TransactionHandler) GetMostSpend(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized get most spend: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}
	monthStr := q.Get("month")
	yearStr := q.Get("year")

	month, err2 := strconv.Atoi(monthStr)
	year, err3 := strconv.Atoi(yearStr)

	if err2 != nil || err3 != nil || userId == 0 || month == 0 || year == 0 {
		h.logger.Warn("Missing or invalid parameters for most spend",
			zap.Uint64("UserId", userId),
			zap.String("month", monthStr),
			zap.String("year", yearStr))
		WriteJSON(w, http.StatusBadRequest, "user_id, month, and year are required", nil)
		return
	}

	limit, _ := strconv.Atoi(q.Get("limit"))
	if limit == 0 {
		limit = 5
	}

	h.logger.Debug("Fetching most spend categories",
		zap.Uint64("UserId", userId),
		zap.Int("month", month),
		zap.Int("year", year),
		zap.Int("limit", limit))

	spends, err := h.svc.GetMostSpend(r.Context(), userId, month, year, limit)
	if err != nil {
		h.logger.Error("Failed to get most spend",
			zap.Error(err),
			zap.Uint64("UserId", userId),
			zap.Int("month", month),
			zap.Int("year", year))
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

	h.logger.Debug("Most spend categories retrieved successfully",
		zap.Uint64("UserId", userId),
		zap.Int("count", len(responses)))
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
