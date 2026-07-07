package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	service "walletwise/internal/application/wallet"
	"walletwise/internal/domain/wallet"
)

// --- DTO / Response Structs ---

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
}

// --- Handler ---

type WalletHandler struct {
	service *service.Service
}

func NewWalletHandler(service *service.Service) *WalletHandler {
	return &WalletHandler{service: service}
}

func (wh *WalletHandler) CreateWallets(w http.ResponseWriter, r *http.Request) {
	var req CreateWalletRequest

	// PERBAIKAN: Gunakan Decoder, bukan Encoder
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := service.WalletInput{
		UserID:     req.UserID,
		WalletName: req.WalletName,
		WalletType: req.WalletType,
		// Balance di-set 0 di service sesuai logika Anda
	}

	err := wh.service.CreateWallet(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusCreated, "Created Wallet Successfully", nil)
}

func (wh *WalletHandler) SearchAllWallets(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.URL.Query().Get("userId")
	if userIdStr == "" {
		WriteJson(w, http.StatusBadRequest, "User Id is required", nil)
		return
	}

	// PERBAIKAN: Gunakan ParseUint untuk tipe data uint64, bukan Atoi (int)
	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing UserID", nil)
		return
	}

	wallets, err := wh.service.SearchAllWallet(r.Context(), userId)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Mapping ke Response Struct agar JSON tidak kosong
	var responses []WalletResponse
	for _, w := range wallets {
		responses = append(responses, toWalletResponse(w))
	}
	if responses == nil {
		responses = []WalletResponse{}
	}

	WriteJson(w, http.StatusOK, "Success Getting All Wallets", responses)
}

func (wh *WalletHandler) SearchWalletsByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		WriteJson(w, http.StatusBadRequest, "ID is required", nil)
		return
	}

	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing ID", nil)
		return
	}

	wData, err := wh.service.SearchWalletByID(r.Context(), id)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success Getting Wallet", toWalletResponse(wData))
}

func (wh *WalletHandler) UpdateWallet(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req UpdateWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := service.WalletUpdateInput{
		ID:         id,
		UserID:     req.UserID,
		WalletName: req.WalletName,
		WalletType: req.WalletType,
	}

	err = wh.service.UpdateWallet(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Updated Wallet Successfully", nil)
}

func (wh *WalletHandler) DeleteWallet(w http.ResponseWriter, r *http.Request) {
	// Standar REST API: Ambil ID Wallet yang mau dihapus dari URL, bukan dari Body
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	// CATATAN: Service DeleteWallet Anda saat ini menggunakan WalletInput
	// dan menjadikan input.userId sebagai WalletID.
	// Saya memetakan variabel 'id' dari URL ke UserID agar kodenya bisa jalan.
	input := service.WalletInput{
		UserID: id,
	}

	err = wh.service.DeleteWallet(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Deleted Wallet Successfully", nil)
}

func (wh *WalletHandler) SearchHighestBalance(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.PathValue("userId")
	if userIdStr == "" {
		WriteJson(w, http.StatusBadRequest, "User Id is required", nil)
		return
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing UserID", nil)
		return
	}

	wData, err := wh.service.SearchHighestBalanceWallet(r.Context(), wallet.UserID(userId))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success Getting Highest Balance Wallet", toWalletResponse(wData))
}

func (wh *WalletHandler) SearchMostActive(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.PathValue("userId")
	if userIdStr == "" {
		WriteJson(w, http.StatusBadRequest, "User Id is required", nil)
		return
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing UserID", nil)
		return
	}

	wData, err := wh.service.SearchMostActiveWallet(r.Context(), wallet.UserID(userId))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success Getting Most Active Wallet", toWalletResponse(wData))
}

func (wh *WalletHandler) SearchTotalBalance(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.PathValue("userId")
	if userIdStr == "" {
		WriteJson(w, http.StatusBadRequest, "User Id is required", nil)
		return
	}

	userId, err := strconv.ParseUint(userIdStr, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing UserID", nil)
		return
	}

	totalBalance, err := wh.service.SearchTotalBalanceWallet(r.Context(), wallet.UserID(userId))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	// Return map sederhana untuk total balance
	response := map[string]uint64{
		"total_balance": totalBalance,
	}

	WriteJson(w, http.StatusOK, "Success Getting Total Balance", response)
}

// --- Helper Functions ---

// toWalletResponse memetakan Entity Wallet ke Struct Response API agar JSON terbaca
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
