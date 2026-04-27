package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"walletwise/internal/domain/transaction"

	service "walletwise/internal/application/transaction"
)

type TransactionHandler struct {
	service *service.Service
}

func NewTransactionHandler(s *service.Service) *TransactionHandler {
	return &TransactionHandler{service: s}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	var input service.TrxInput

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteJson(w, http.StatusBadRequest, "Format JSON Tidak Valid", nil)
		return
	}

	trx, err := h.service.CreateTransaction(r.Context(), &input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Create Transaction Error", nil)
		return
	}
	WriteJson(w, http.StatusCreated, "Success Create Transaction", trx)
	return
}

func (h *TransactionHandler) GetTransaction(w http.ResponseWriter, r *http.Request) {
	var userId uint64 = 1
	//userIdd := r.Context().Value("user_id").(uint64)
	input := &service.GetTransactionsInput{
		UserID: userId,
		Limit:  10,
	}

	query := r.URL.Query()

	if limitStr := query.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil && limit <= 0 {
			input.Limit = limit
		}
	}

	if goalStr := query.Get("goalId"); goalStr != "" {
		goalId, err := strconv.ParseUint(goalStr, 10, 64)
		if err != nil && goalId >= 0 {
			input.GoalID = &goalId
		}
	}
	if amountStr := query.Get("amount"); amountStr != "" {
		amount, err := strconv.ParseUint(amountStr, 10, 64)
		if err != nil && amount >= 0 {
			moneyAmt := amount
			input.Amount = &moneyAmt
		}
	}
	if typeStr := query.Get("type"); typeStr != "" {
		convType := typeStr
		input.TransactionType = &convType
	}
	if categoryStr := query.Get("category"); categoryStr != "" {
		newCategoryID, err := strconv.ParseUint(categoryStr, 10, 64)
		if err != nil && newCategoryID >= 0 {
			input.CategoryID = &newCategoryID
		}
	}
	if startDatestr := query.Get("startDate"); startDatestr != "" {
		parsedStartDate, err := time.Parse(time.DateOnly, startDatestr)
		if err != nil {
			WriteJson(w, http.StatusBadRequest, "Start Date Format Error", nil)
			return
		}
		input.StartDate = &parsedStartDate
	}
	if endDatestr := query.Get("endDate"); endDatestr != "" {
		parsedEndDate, err := time.Parse(time.DateOnly, endDatestr)
		if err != nil {
			WriteJson(w, http.StatusBadRequest, "End Date Format Error", nil)
			return
		}
		input.EndDate = &parsedEndDate
	}
	trxList, err := h.service.GetTransaction(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Get Transaction Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "Success Get Transaction", trxList)
	return
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	var input service.TrxUpdate

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteJson(w, http.StatusBadRequest, "Format JSON Tidak Valid", nil)
		return
	}
	err := h.service.UpdateTransaction(r.Context(), &input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Update Transaction Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "Success Update Transaction", nil)
	return
}

func (h *TransactionHandler) DeleteTransaction(w http.ResponseWriter, r *http.Request) {
	var transactionId uint64 = 1
	//trxId := r.PathValue("id")

	err := h.service.DeleteTransaction(r.Context(), transaction.TransactionID(transactionId))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Delete Transaction Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "Success Delete Transaction", nil)
	return
}

func (h *TransactionHandler) GetTransactionById(w http.ResponseWriter, r *http.Request) {
	trxId := r.PathValue("transactionID")
	intTransctionId, err := strconv.ParseUint(trxId, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing TransactionId", nil)
		return
	}

	newTransactionId := transaction.TransactionID(intTransctionId)

	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing TransactionId", nil)
		return
	}
	trx, err := h.service.GetTransactionById(r.Context(), newTransactionId)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Get Transaction Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "Success Get Transaction", trx)
}
