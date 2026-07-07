package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	service "walletwise/internal/application/user"
	"walletwise/internal/domain/users"
)

// --- DTO / Response Structs ---

type UserResponse struct {
	ID           uint64    `json:"id"`
	Username     string    `json:"username"`
	Email        string    `json:"email"`
	MonthlyLimit uint64    `json:"monthly_limit"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	MonthlyLimit uint64 `json:"monthly_limit"`
	IsActive     bool   `json:"is_active"`
}

type UpdateUserRequest struct {
	Username     string `json:"username"`
	Email        string `json:"email"`
	Password     string `json:"password"`
	MonthlyLimit uint64 `json:"monthly_limit"`
	IsActive     bool   `json:"is_active"`
}

// --- Handler ---

type UserHandler struct {
	service *service.Service
}

func NewUserHandler(service *service.Service) *UserHandler {
	return &UserHandler{service: service}
}

func (u *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest

	// PERBAIKAN: Menggunakan Decoder untuk membaca Request Body, bukan Encoder
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := service.UserInput{
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		MonthlyLimit: req.MonthlyLimit,
		IsActive:     req.IsActive,
	}

	user, err := u.service.CreateUser(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusCreated, "User Created", toUserResponse(user))
}

func (u *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idString := r.PathValue("id")
	if idString == "" {
		WriteJson(w, http.StatusBadRequest, "ID parameter is required", nil)
		return
	}

	userId, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	user, err := u.service.SearchUserById(r.Context(), users.UserID(userId))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "User Found", toUserResponse(user))
}

func (u *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.PathValue("email")
	if email == "" {
		WriteJson(w, http.StatusBadRequest, "Email parameter is required", nil)
		return
	}

	userEntity, err := u.service.SearchUserByEmail(r.Context(), email)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "User Found", toUserResponse(userEntity))
}

func (u *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL (Standar REST: PUT /users/{id})
	idString := r.PathValue("id")
	userId, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := &service.UserUpdateInput{
		ID:           userId,
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		MonthlyLimit: req.MonthlyLimit,
		IsActive:     req.IsActive,
	}

	err = u.service.UpdateUser(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "User Updated", nil)
}

func (u *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	// Ambil ID dari URL (Standar REST: DELETE /users/{id})
	idString := r.PathValue("id")
	userId, err := strconv.ParseUint(idString, 10, 64)
	if err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid ID format", nil)
		return
	}

	input := service.UserUpdateInput{
		ID: userId,
	}

	err = u.service.DeleteUser(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "User Deleted", nil)
}

// --- Helper Functions ---

// toUserResponse memetakan Entity User ke Struct Response API
func toUserResponse(user *users.User) UserResponse {
	return UserResponse{
		ID:           uint64(user.UserID()),
		Username:     user.Username(),
		Email:        user.Email(),
		MonthlyLimit: uint64(user.MonthlyLimit()),
		IsActive:     user.IsActive(),
		CreatedAt:    user.CreatedAt(),
		UpdatedAt:    user.UpdatedAt(),
	}
}
