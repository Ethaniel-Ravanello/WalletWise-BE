package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	service "walletwise/internal/application/user"
	"walletwise/internal/domain/users"
)

type UserHandler struct {
	service *service.Service
}

func NewUserHandler(service *service.Service) *UserHandler { return &UserHandler{service: service} }

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var userinput service.UserInput

	if err := json.NewEncoder(w).Encode(userinput); err != nil {
		WriteJson(w, http.StatusBadRequest, "Format JSON Tidak Valid", nil)
		return
	}
	user, err := u.service.CreateUser(r.Context(), userinput)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "User Created", user)
}

func (u *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	// Tangkap {id} dari URL /users/{id}
	idString := r.PathValue("id")
	if idString == "" {
		WriteJson(w, http.StatusBadRequest, "ID parameter is required", nil)
		return
	}

	// Convert string ke angka
	userId, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	user, err := u.service.SearchUserById(r.Context(), users.UserID(userId))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "User Found", user)
}

func (u *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.PathValue("email")
	if email == "" {
		WriteJson(w, http.StatusBadRequest, "Email parameter is required", nil)
		return
	}

	userEntity, err := u.service.SearchUserByEmail(r.Context(), email)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "User Not Found", nil)
		return
	}

	WriteJson(w, http.StatusOK, "User Found", userEntity)
}

func (u *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	var userInput service.UserUpdateInput

	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		WriteJson(w, http.StatusBadRequest, "Format JSON Tidak Valid", nil)
		return
	}

	err := u.service.UpdateUser(r.Context(), &userInput)

	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "User Updated", nil)
}

func (u *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	var userInput service.UserUpdateInput

	if err := json.NewDecoder(r.Body).Decode(&userInput); err != nil {
		WriteJson(w, http.StatusBadRequest, "Format JSON Tidak Valid", nil)
		return
	}

	err := u.service.DeleteUser(r.Context(), userInput)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, "Internal Server Error", nil)
		return
	}
	WriteJson(w, http.StatusOK, "User Deleted", nil)
}
