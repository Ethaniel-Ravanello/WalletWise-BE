package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	service "walletwise/internal/application/wallet"
	"walletwise/internal/domain/wallet"
)

type WalletHandler struct {
	service *service.Service
}

func NewWalletHandler(service *service.Service) *WalletHandler {
	return &WalletHandler{service: service}
}

func (wh *WalletHandler) CreateWallets(w http.ResponseWriter, r *http.Request) {
	var walletInput service.WalletInput

	if err := json.NewEncoder(w).Encode(walletInput); err != nil {
		WriteJson(w, http.StatusBadRequest, "Format Input Tidak Valid", nil)
		return
	}
	err := wh.service.CreateWallet(r.Context(), walletInput)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Error", nil)
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
	userIdInt, err := strconv.Atoi(userIdStr)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing UserID", nil)
		return
	}
	wallets, err := wh.service.SearchAllWallet(r.Context(), uint64(userIdInt))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "Success Getting All Wallets", wallets)
}

func (wh *WalletHandler) SearchWalletsByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		WriteJson(w, http.StatusBadRequest, "ID is required", nil)
		return
	}
	idInt, err := strconv.Atoi(idStr)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing ID", nil)
		return
	}
	wallet, err := wh.service.SearchWalletByID(r.Context(), uint64(idInt))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "Success Getting Wallet", wallet)

}

func (wh *WalletHandler) UpdateWallet(w http.ResponseWriter, r *http.Request) {
	var walletUpdateInput service.WalletUpdateInput
	if err := json.NewDecoder(r.Body).Decode(&walletUpdateInput); err != nil {
		WriteJson(w, http.StatusBadRequest, "Format Input Tidak Valid", nil)
		return
	}
	err := wh.service.UpdateWallet(r.Context(), walletUpdateInput)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "Updated Wallet Successfully", nil)
}

func (wh *WalletHandler) DeleteWallet(w http.ResponseWriter, r *http.Request) {
	var walletInput service.WalletInput
	if err := json.NewDecoder(r.Body).Decode(&walletInput); err != nil {
		WriteJson(w, http.StatusBadRequest, "Format Input Tidak Valid", nil)
		return
	}

	err := wh.service.DeleteWallet(r.Context(), walletInput)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Error", nil)
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
	userIdInt, err := strconv.Atoi(userIdStr)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing UserID", nil)
	}
	wallets, err := wh.service.SearchHighestBalanceWallet(r.Context(), wallet.UserID(userIdInt))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Error", nil)
	}
	WriteJson(w, http.StatusOK, "Success Getting Highest Balance Wallets", wallets)
}

func (wh *WalletHandler) SearchMostActive(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.PathValue("userId")
	if userIdStr == "" {
		WriteJson(w, http.StatusBadRequest, "User Id is required", nil)
		return
	}
	userIdInt, err := strconv.Atoi(userIdStr)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing UserID", nil)
		return
	}
	wallet, err := wh.service.SearchMostActiveWallet(r.Context(), wallet.UserID(userIdInt))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "Success Getting Most Active Wallets", wallet)
}

func (wh *WalletHandler) SearchTotalBalance(w http.ResponseWriter, r *http.Request) {
	userIdStr := r.PathValue("userId")
	if userIdStr == "" {
		WriteJson(w, http.StatusBadRequest, "User Id is required", nil)
		return
	}
	userIdInt, err := strconv.Atoi(userIdStr)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Error Parsing UserID", nil)
		return
	}
	wallets, err := wh.service.SearchTotalBalanceWallet(r.Context(), wallet.UserID(userIdInt))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "Success Getting Total Balance Wallets", wallets)
}
