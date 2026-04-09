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
	}
	WriteJson(w, http.StatusCreated, "Success Create Transaction", trx)
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
		if err != nil && limit >= 0 {
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
		amount, err := strconv.Atoi(amountStr)
		if err != nil && amount >= 0 {
			moneyAmt := transaction.Money(amount)
			input.Amount = &moneyAmt
		}
	}
	if typeStr := query.Get("type"); typeStr != "" {
		convType := transaction.Type(typeStr)
		input.Type = &convType
	}
	if categoryStr := query.Get("category"); categoryStr != "" {
		input.Category = &categoryStr
	}
	if startDatestr := query.Get("startDate"); startDatestr != "" {
		parsedStartDate, err := time.Parse(time.DateOnly, startDatestr)
		if err != nil {
			WriteJson(w, http.StatusBadRequest, "Start Date Format Error", nil)
		}
		input.StartDate = &parsedStartDate
	}
	if endDatestr := query.Get("endDate"); endDatestr != "" {
		parsedEndDate, err := time.Parse(time.DateOnly, endDatestr)
		if err != nil {
			WriteJson(w, http.StatusBadRequest, "End Date Format Error", nil)
		}
		input.EndDate = &parsedEndDate
	}
	trxList, err := h.service.GetTransaction(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Get Transaction Error", nil)
	}
	WriteJson(w, http.StatusOK, "Success Get Transaction", trxList)
}

func (h *TransactionHandler) UpdateTransaction(w http.ResponseWriter, r *http.Request) {
	var input service.TrxUpdate

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteJson(w, http.StatusBadRequest, "Format JSON Tidak Valid", nil)
	}
	err := h.service.UpdateTransaction(r.Context(), &input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Update Transaction Error", nil)
	}
	WriteJson(w, http.StatusOK, "Success Update Transaction", nil)
}
