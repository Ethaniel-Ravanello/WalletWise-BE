package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	service "walletwise/internal/application/user"
	"walletwise/internal/domain/user"
	"walletwise/internal/middleware"
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
	svc    *service.Service
	logger *zap.Logger
}

func NewUserHandler(svc *service.Service) *UserHandler {
	return &UserHandler{
		svc:    svc,
		logger: zap.L(),
	}
}

func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode create user request", zap.Error(err))
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.Username == "" || req.Email == "" || req.Password == "" {
		h.logger.Warn("Create user validation failed: missing required fields",
			zap.String("Username", req.Username),
			zap.String("Email", req.Email))
		WriteJSON(w, http.StatusBadRequest, "username, email, and password are required", nil)
		return
	}

	h.logger.Debug("Creating user",
		zap.String("Username", req.Username),
		zap.String("Email", req.Email))

	input := service.UserInput{
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		MonthlyLimit: req.MonthlyLimit,
		IsActive:     req.IsActive,
	}

	user, err := h.svc.CreateUser(r.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create user",
			zap.Error(err),
			zap.String("Username", req.Username),
			zap.String("Email", req.Email))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("User created successfully",
		zap.Uint64("UserId", uint64(user.ID())),
		zap.String("Username", user.Username()),
		zap.String("Email", user.Email()))
	WriteJSON(w, http.StatusCreated, "User created successfully", toUserResponse(user))
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var input UserLoginRequest

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		h.logger.Warn("Failed to decode login request", zap.Error(err))
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if input.Email == "" || input.Password == "" {
		h.logger.Warn("Login validation failed: missing email or password", zap.String("Email", input.Email))
		WriteJSON(w, http.StatusBadRequest, "email, and password are required", nil)
		return
	}

	h.logger.Debug("Attempting user login", zap.String("Email", input.Email))

	jwtToken, err := h.svc.Login(r.Context(), input.Email, input.Password)
	if err != nil {
		h.logger.Warn("User login failed",
			zap.Error(err),
			zap.String("Email", input.Email))
		WriteJSON(w, http.StatusUnauthorized, err.Error(), nil)
		return
	}

	h.logger.Info("User login successful", zap.String("Email", input.Email))
	WriteJSON(w, http.StatusOK, "Login Successful", LoginResponse(jwtToken))
}

func (h *UserHandler) GetUserByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	if idStr == "" {
		h.logger.Warn("User ID parameter is required")
		WriteJSON(w, http.StatusBadRequest, "User ID parameter is required", nil)
		return
	}

	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid user ID format",
			zap.Error(err),
			zap.String("id", idStr))
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	h.logger.Debug("Fetching user by ID", zap.Uint64("UserId", userID))

	user, err := h.svc.SearchUserById(r.Context(), userID)
	if err != nil {
		h.logger.Warn("User not found by ID",
			zap.Error(err),
			zap.Uint64("UserId", userID))
		WriteJSON(w, http.StatusNotFound, "User not found", nil)
		return
	}

	h.logger.Debug("User retrieved successfully by ID", zap.Uint64("UserId", userID))
	WriteJSON(w, http.StatusOK, "User retrieved successfully", toUserResponse(user))
}

func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := r.PathValue("email")
	if email == "" {
		h.logger.Warn("Email parameter is required")
		WriteJSON(w, http.StatusBadRequest, "Email parameter is required", nil)
		return
	}

	h.logger.Debug("Fetching user by email", zap.String("Email", email))

	user, err := h.svc.SearchUserByEmail(r.Context(), email)
	if err != nil {
		h.logger.Warn("User not found by email",
			zap.Error(err),
			zap.String("Email", email))
		WriteJSON(w, http.StatusNotFound, "User not found", nil)
		return
	}

	h.logger.Debug("User retrieved successfully by email", zap.String("Email", email))
	WriteJSON(w, http.StatusOK, "User retrieved successfully", toUserResponse(user))
}

func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized user update: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode update user request",
			zap.Error(err),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	h.logger.Debug("Updating user", zap.Uint64("UserId", userId))

	input := &service.UserUpdateInput{
		ID:           userId,
		Username:     req.Username,
		Email:        req.Email,
		Password:     req.Password,
		MonthlyLimit: req.MonthlyLimit,
		IsActive:     req.IsActive,
	}

	if err := h.svc.UpdateUser(r.Context(), input, userId); err != nil {
		h.logger.Error("Failed to update user",
			zap.Error(err),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("User updated successfully", zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "User updated successfully", nil)
}

func (h *UserHandler) DeleteUser(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized user deletion: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	userID, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid user ID format for deletion",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid user ID format", nil)
		return
	}

	h.logger.Debug("Deleting user",
		zap.Uint64("TargetUserId", userID),
		zap.Uint64("UserId", userId))

	input := service.UserUpdateInput{
		ID: userID,
	}

	if err := h.svc.DeleteUser(r.Context(), input, userId); err != nil {
		h.logger.Error("Failed to delete user",
			zap.Error(err),
			zap.Uint64("TargetUserId", userID),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("User deleted successfully",
		zap.Uint64("TargetUserId", userID),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "User deleted successfully", nil)
}

func toUserResponse(u *user.User) UserResponse {
	if u == nil {
		return UserResponse{}
	}
	return UserResponse{
		ID:           uint64(u.ID()),
		Username:     u.Username(),
		Email:        u.Email(),
		MonthlyLimit: uint64(u.MonthlyLimit()),
		IsActive:     u.IsActive(),
		CreatedAt:    u.CreatedAt(),
		UpdatedAt:    u.UpdatedAt(),
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
