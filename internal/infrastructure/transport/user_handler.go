package transport

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"walletwise/internal/middleware"

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

type UserLoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserLoginResponse struct {
	Token string `json:"token"`
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

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input UserLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if input.Email == "" || input.Password == "" {
		WriteJSON(w, http.StatusBadRequest, "email, and password are required", nil)
	}

	jwtToken, err := h.svc.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		WriteJSON(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}
	fmt.Println(jwtToken)
	WriteJSON(w, http.StatusOK, "Login Successful", LoginResponse(jwtToken))
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

	user, err := h.svc.SearchUserById(r.Context(), userID)
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
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
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

	if err := h.svc.UpdateUser(r.Context(), input, userId); err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "User updated successfully", nil)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}

	idStr := r.PathValue("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	input := service.UserUpdateInput{
		ID: userID,
	}

	if err := h.svc.DeleteUser(r.Context(), input, userId); err != nil {
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
		Username:     user.Username(),
		Email:        user.Email(),
		MonthlyLimit: uint64(user.MonthlyLimit()),
		IsActive:     user.IsActive(),
		CreatedAt:    user.CreatedAt(),
		UpdatedAt:    user.UpdatedAt(),
	}
}

func LoginResponse(token string) UserLoginResponse {
	if token == "" {
		return UserLoginResponse{}
	}
	return UserLoginResponse{
		Token: token,
	}
}
