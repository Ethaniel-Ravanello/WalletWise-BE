package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	service "walletwise/internal/application/wallet"
	"walletwise/internal/domain/wallet"
	"walletwise/internal/middleware"
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
	svc    *service.Service
	logger *zap.Logger
}

func NewWalletHandler(svc *service.Service) *WalletHandler {
	return &WalletHandler{
		svc:    svc,
		logger: zap.L(),
	}
}

func (h *WalletHandler) CreateWallet(w http.ResponseWriter, r *http.Request) {
	var req CreateWalletRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode create wallet request", zap.Error(err))
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
		h.logger.Warn("Create wallet validation failed: missing required fields",
			zap.Uint64("UserId", req.UserID),
			zap.String("WalletName", req.WalletName),
			zap.String("WalletType", req.WalletType))
		WriteJSON(w, http.StatusBadRequest, "user_id, wallet_name, and wallet_type are required", nil)
		return
	}

	h.logger.Debug("Creating wallet",
		zap.Uint64("UserId", req.UserID),
		zap.String("WalletName", req.WalletName),
		zap.String("WalletType", req.WalletType))

	input := service.WalletInput{
		UserID:     req.UserID,
		WalletName: req.WalletName,
		WalletType: req.WalletType,
	}

	err := h.svc.CreateWallet(r.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create wallet",
			zap.Error(err),
			zap.Uint64("UserId", req.UserID),
			zap.String("WalletName", req.WalletName))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Wallet created successfully",
		zap.Uint64("UserId", req.UserID),
		zap.String("WalletName", req.WalletName))
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
		h.logger.Warn("Search wallets validation failed: missing user_id")
		WriteJSON(w, http.StatusBadRequest, "user_id is required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid user_id format for wallets search",
			zap.Error(err),
			zap.String("user_id", userIDStr))
		WriteJSON(w, http.StatusBadRequest, "Invalid user_id format", nil)
		return
	}

	h.logger.Debug("Searching all wallets", zap.Uint64("UserId", userID))

	wallets, err := h.svc.SearchAllWallet(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to search wallets",
			zap.Error(err),
			zap.Uint64("UserId", userID))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responses := make([]WalletResponse, 0, len(wallets))
	for _, wlt := range wallets {
		responses = append(responses, toWalletResponse(wlt))
	}

	h.logger.Debug("Wallets retrieved successfully",
		zap.Uint64("UserId", userID),
		zap.Int("count", len(responses)))
	WriteJSON(w, http.StatusOK, "Wallets retrieved successfully", responses)
}

func (h *WalletHandler) GetWallets(w http.ResponseWriter, r *http.Request) {
	h.SearchAllWallets(w, r)
}

func (h *WalletHandler) SearchWalletsByID(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized search wallet by ID: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	if idStr == "" {
		h.logger.Warn("Wallet ID parameter is required")
		WriteJSON(w, http.StatusBadRequest, "Wallet ID is required", nil)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid wallet ID format",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid wallet ID format", nil)
		return
	}

	h.logger.Debug("Searching wallet by ID",
		zap.Uint64("WalletId", id),
		zap.Uint64("UserId", userId))

	walletData, err := h.svc.SearchWalletByID(r.Context(), id, userId)
	if err != nil {
		h.logger.Warn("Wallet not found",
			zap.Error(err),
			zap.Uint64("WalletId", id),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusNotFound, "Wallet not found", nil)
		return
	}

	h.logger.Debug("Wallet retrieved successfully",
		zap.Uint64("WalletId", id),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "Wallet retrieved successfully", toWalletResponse(walletData))
}

func (h *WalletHandler) GetWalletByID(w http.ResponseWriter, r *http.Request) {
	h.SearchWalletsByID(w, r)
}

func (h *WalletHandler) UpdateWallet(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized update wallet: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid wallet ID format for update",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid wallet ID format", nil)
		return
	}

	var req UpdateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode update wallet request",
			zap.Error(err),
			zap.Uint64("WalletId", id),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.UserID == 0 {
		req.UserID = userId
	}

	h.logger.Debug("Updating wallet",
		zap.Uint64("WalletId", id),
		zap.Uint64("UserId", userId))

	input := service.WalletUpdateInput{
		ID:         id,
		UserID:     req.UserID,
		WalletName: req.WalletName,
		WalletType: req.WalletType,
		Balance:    req.Balance,
	}

	err = h.svc.UpdateWallet(r.Context(), input, userId)
	if err != nil {
		h.logger.Error("Failed to update wallet",
			zap.Error(err),
			zap.Uint64("WalletId", id),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Wallet updated successfully",
		zap.Uint64("WalletId", id),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "Wallet updated successfully", nil)
}

func (h *WalletHandler) DeleteWallet(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized delete wallet: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid wallet ID format for deletion",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
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

	h.logger.Debug("Deleting wallet",
		zap.Uint64("WalletId", id),
		zap.Uint64("UserId", userId))

	input := service.WalletUpdateInput{
		ID:         id,
		UserID:     userID,
		WalletName: req.WalletName,
		WalletType: req.WalletType,
		Balance:    req.Balance,
	}

	err = h.svc.DeleteWallet(r.Context(), input, userId)
	if err != nil {
		h.logger.Error("Failed to delete wallet",
			zap.Error(err),
			zap.Uint64("WalletId", id),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Wallet deleted successfully",
		zap.Uint64("WalletId", id),
		zap.Uint64("UserId", userId))
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
		h.logger.Warn("Missing user ID parameter for highest balance wallet")
		WriteJSON(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid user ID format for highest balance",
			zap.Error(err),
			zap.String("userId", userIDStr))
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	h.logger.Debug("Searching highest balance wallet", zap.Uint64("UserId", userID))

	walletData, err := h.svc.SearchHighestBalanceWallet(r.Context(), wallet.UserID(userID))
	if err != nil {
		h.logger.Warn("Highest balance wallet not found",
			zap.Error(err),
			zap.Uint64("UserId", userID))
		WriteJSON(w, http.StatusNotFound, "Highest balance wallet not found", nil)
		return
	}

	h.logger.Debug("Highest balance wallet retrieved successfully", zap.Uint64("UserId", userID))
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
		h.logger.Warn("Missing user ID parameter for most active wallet")
		WriteJSON(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid user ID format for most active wallet",
			zap.Error(err),
			zap.String("userId", userIDStr))
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	h.logger.Debug("Searching most active wallet", zap.Uint64("UserId", userID))

	walletData, err := h.svc.SearchMostActiveWallet(r.Context(), wallet.UserID(userID))
	if err != nil {
		h.logger.Warn("Most active wallet not found",
			zap.Error(err),
			zap.Uint64("UserId", userID))
		WriteJSON(w, http.StatusNotFound, "Most active wallet not found", nil)
		return
	}

	h.logger.Debug("Most active wallet retrieved successfully", zap.Uint64("UserId", userID))
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
		h.logger.Warn("Missing user ID parameter for total balance")
		WriteJSON(w, http.StatusBadRequest, "User ID is required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid user ID format for total balance",
			zap.Error(err),
			zap.String("userId", userIDStr))
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	h.logger.Debug("Calculating total balance", zap.Uint64("UserId", userID))

	totalBalance, err := h.svc.SearchTotalBalanceWallet(r.Context(), wallet.UserID(userID))
	if err != nil {
		h.logger.Error("Failed to calculate total balance",
			zap.Error(err),
			zap.Uint64("UserId", userID))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	response := map[string]uint64{
		"total_balance": totalBalance,
	}

	h.logger.Debug("Total balance calculated successfully",
		zap.Uint64("UserId", userID),
		zap.Uint64("total_balance", totalBalance))
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
