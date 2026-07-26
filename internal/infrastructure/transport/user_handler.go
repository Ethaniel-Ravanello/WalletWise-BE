package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	service "walletwise/internal/application/user"
	"walletwise/internal/domain/users"
)

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

type UserHandler struct {
	svc *service.Service
}

func NewUserHandler(svc *service.Service) *UserHandler {
	return &UserHandler{svc: svc}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		WriteJSON(w, http.StatusBadRequest, "username, email, and password are required", nil)
		return
	}

	input := service.UserInput{
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		MonthlyLimit: req.MonthlyLimit,
		IsActive:     req.IsActive,
	}

	user, err := h.svc.CreateUser(r.Context(), input)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "User created successfully", toUserResponse(user))
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		WriteJSON(w, http.StatusBadRequest, "User ID parameter is required", nil)
		return
	}

	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	user, err := h.svc.SearchUserById(r.Context(), users.UserID(userID))
	if err != nil {
		WriteJSON(w, http.StatusNotFound, "User not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "User retrieved successfully", toUserResponse(user))
}

func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.PathValue("email")
	if email == "" {
		WriteJSON(w, http.StatusBadRequest, "Email parameter is required", nil)
		return
	}

	user, err := h.svc.SearchUserByEmail(r.Context(), email)
	if err != nil {
		WriteJSON(w, http.StatusNotFound, "User not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "User retrieved successfully", toUserResponse(user))
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := &service.UserUpdateInput{
		ID:           userID,
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		MonthlyLimit: req.MonthlyLimit,
		IsActive:     req.IsActive,
	}

	if err := h.svc.UpdateUser(r.Context(), input); err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "User updated successfully", nil)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	input := service.UserUpdateInput{
		ID: userID,
	}

	if err := h.svc.DeleteUser(r.Context(), input); err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "User deleted successfully", nil)
}

func toUserResponse(user *users.User) UserResponse {
	if user == nil {
		return UserResponse{}
	}
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


