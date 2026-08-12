package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"walletwise/internal/middleware"

	service "walletwise/internal/application/wallet"
	"walletwise/internal/domain/wallet"
)

type WalletResponse struct {
	ID         uint64    `json:"id"`
	UserID     uint64    `json:"user_id"`
	WalletName string    `json:"wallet_name"`
	WalletType string    `json:"wallet_type"`
	Balance    uint64    `json:"balance"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type CreateWalletRequest struct {
	UserID     uint64 `json:"user_id"`
	WalletName string `json:"wallet_name"`
	WalletType string `json:"wallet_type"`
}

type UpdateWalletRequest struct {
	UserID     uint64 `json:"user_id"`
	WalletName string `json:"wallet_name"`
	WalletType string `json:"wallet_type"`
	Balance    uint64 `json:"balance"`
}

type WalletHandler struct {
	svc *service.Service
}

func NewWalletHandler(svc *service.Service) *WalletHandler {
	return &WalletHandler{svc: svc}
}

func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req CreateWalletRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.UserID == 0 {
		userIdCtx := r.Context().Value(middleware.UserIdKey)
		if userId, ok := userIdCtx.(uint64); ok {
			req.UserID = userId
		}
	}

	if req.UserID == 0 || req.WalletName == "" || req.WalletType == "" {
		WriteJSON(w, http.StatusBadRequest, "user_id, wallet_name, and wallet_type are required", nil)
		return
	}

	input := service.WalletInput{
		UserID:     req.UserID,
		WalletName: req.WalletName,
		WalletType: req.WalletType,
	}
	fmt.Println(input)
	err := h.svc.CreateWallet(r.Context(), input)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "Wallet created successfully", nil)
}

// CreateWallets is an alias for CreateWallet for backward compatibility.
func (h *WalletHandler) CreateWallets(w http.ResponseWriter, r *http.Request) {
	h.CreateWallet(w, r)
}

func (h *WalletHandler) SearchAllWallets(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("userId")
	}
	if userIDStr == "" {
		userIdCtx := r.Context().Value(middleware.UserIdKey)
		if userId, ok := userIdCtx.(uint64); ok && userId > 0 {
			userIDStr = strconv.FormatUint(userId, 10)
		}
	}
	if userIDStr == "" {
		WriteJSON(w, http.StatusBadRequest, "user_id is required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user_id format", nil)
		return
	}

	wallets, err := h.svc.SearchAllWallet(r.Context(), userID)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responses := make([]WalletResponse, 0, len(wallets))
	for _, wlt := range wallets {
		responses = append(responses, toWalletResponse(wlt))
	}

	WriteJSON(w, http.StatusOK, "Wallets retrieved successfully", responses)
}

func (h *WalletHandler) GetWallets(w http.ResponseWriter, r *http.Request) {
	h.SearchAllWallets(w, r)
}

func (h *WalletHandler) SearchWalletsByID(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		WriteJSON(w, http.StatusBadRequest, "Wallet ID is required", nil)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid wallet ID format", nil)
		return
	}

	walletData, err := h.svc.SearchWalletByID(r.Context(), id, userId)
	if err != nil {
		WriteJSON(w, http.StatusNotFound, "Wallet not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Wallet retrieved successfully", toWalletResponse(walletData))
}

func (h *WalletHandler) GetWalletByID(w http.ResponseWriter, r *http.Request) {
	h.SearchWalletsByID(w, r)
}

func (h *WalletHandler) UpdateWallet(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid wallet ID format", nil)
		return
	}

	var req UpdateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.UserID == 0 {
		req.UserID = userId
	}

	input := service.WalletUpdateInput{
		ID:         id,
		UserID:     req.UserID,
		WalletName: req.WalletName,
		WalletType: req.WalletType,
		Balance:    req.Balance,
	}

	err = h.svc.UpdateWallet(r.Context(), input, userId)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Wallet updated successfully", nil)
}

func (h *WalletHandler) DeleteWallet(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid wallet ID format", nil)
		return
	}

	var req UpdateWalletRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	userID := req.UserID
	if userID == 0 {
		userIDStr := r.URL.Query().Get("user_id")
		if userIDStr == "" {
			userIDStr = r.URL.Query().Get("userId")
		}
		if userIDStr != "" {
			userID, _ = strconv.ParseUint(userIDStr, 10, 64)
		}
	}
	if userID == 0 {
		userID = userId
	}

	input := service.WalletUpdateInput{
		ID:         id,
		UserID:     userID,
		WalletName: req.WalletName,
		WalletType: req.WalletType,
		Balance:    req.Balance,
	}

	err = h.svc.DeleteWallet(r.Context(), input, userId)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Wallet deleted successfully", nil)
}

func (h *WalletHandler) SearchHighestBalance(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("userId")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("user_id")
	}
	if userIDStr == "" {
		userIdCtx := r.Context().Value(middleware.UserIdKey)
		if userId, ok := userIdCtx.(uint64); ok && userId > 0 {
			userIDStr = strconv.FormatUint(userId, 10)
		}
	}
	if userIDStr == "" {
		WriteJSON(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	walletData, err := h.svc.SearchHighestBalanceWallet(r.Context(), wallet.UserID(userID))
	if err != nil {
		WriteJSON(w, http.StatusNotFound, "Highest balance wallet not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Highest balance wallet retrieved successfully", toWalletResponse(walletData))
}

func (h *WalletHandler) SearchMostActive(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("userId")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("user_id")
	}
	if userIDStr == "" {
		userIdCtx := r.Context().Value(middleware.UserIdKey)
		if userId, ok := userIdCtx.(uint64); ok && userId > 0 {
			userIDStr = strconv.FormatUint(userId, 10)
		}
	}
	if userIDStr == "" {
		WriteJSON(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	walletData, err := h.svc.SearchMostActiveWallet(r.Context(), wallet.UserID(userID))
	if err != nil {
		WriteJSON(w, http.StatusNotFound, "Most active wallet not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Most active wallet retrieved successfully", toWalletResponse(walletData))
}

func (h *WalletHandler) SearchTotalBalance(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.PathValue("userId")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("user_id")
	}
	if userIDStr == "" {
		userIdCtx := r.Context().Value(middleware.UserIdKey)
		if userId, ok := userIdCtx.(uint64); ok && userId > 0 {
			userIDStr = strconv.FormatUint(userId, 10)
		}
	}
	if userIDStr == "" {
		WriteJSON(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	totalBalance, err := h.svc.SearchTotalBalanceWallet(r.Context(), wallet.UserID(userID))
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response := map[string]uint64{
		"total_balance": totalBalance,
	}

	WriteJSON(w, http.StatusOK, "Total balance calculated successfully", response)
}

func toWalletResponse(w *wallet.Wallet) WalletResponse {
	if w == nil {
		return WalletResponse{}
	}
	return WalletResponse{
		ID:         uint64(w.ID()),
		UserID:     uint64(w.UserID()),
		WalletName: w.Name(),
		WalletType: w.WalletType(),
		Balance:    uint64(w.Balance()),
		CreatedAt:  w.CreatedAt(),
		UpdatedAt:  w.UpdatedAt(),
	}
}
