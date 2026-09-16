package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"go.uber.org/zap"

	savingGoalService "walletwise/internal/application/saving_goal"
	savingGoalDomain "walletwise/internal/domain/saving_goal"
	"walletwise/internal/middleware"
)

type CreateGoalRequest struct {
	UserID        uint64    `json:"user_id"`
	Name          string    `json:"name"`
	TargetAmount  int64     `json:"target_amount"`
	CurrentAmount int64     `json:"current_amount"`
	Deadline      time.Time `json:"dead_line"`
	GoalStatus    string    `json:"status"`
	Description   string    `json:"description"`
}

type UpdateGoalRequest struct {
	Name          string    `json:"name"`
	TargetAmount  int64     `json:"target_amount"`
	CurrentAmount int64     `json:"current_amount"`
	Deadline      time.Time `json:"dead_line"`
	GoalStatus    string    `json:"status"`
	Description   string    `json:"description"`
}

type GoalResponse struct {
	ID            uint64    `json:"id"`
	UserID        uint64    `json:"user_id"`
	Name          string    `json:"name"`
	TargetAmount  int64     `json:"target_amount"`
	CurrentAmount int64     `json:"current_amount"`
	Deadline      time.Time `json:"dead_line"`
	Status        string    `json:"status"`
	Description   string    `json:"description"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type SavingGoalHandler struct {
	svc    *savingGoalService.Service
	logger *zap.Logger
}

type SavingGoalsHandler = SavingGoalHandler

func NewSavingGoalHandler(svc *savingGoalService.Service) *SavingGoalHandler {
	return &SavingGoalHandler{
		svc:    svc,
		logger: zap.L(),
	}
}

func (h *SavingGoalHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode create saving goal request", zap.Error(err))
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.UserID == 0 {
		userIdCtx := r.Context().Value(middleware.UserIdKey)
		if userId, ok := userIdCtx.(uint64); ok && userId > 0 {
			req.UserID = userId
		}
	}

	if req.UserID == 0 || req.Name == "" {
		h.logger.Warn("Create saving goal validation failed: missing required fields",
			zap.Uint64("UserId", req.UserID),
			zap.String("name", req.Name))
		WriteJSON(w, http.StatusBadRequest, "user_id and name are required", nil)
		return
	}

	h.logger.Debug("Creating saving goal",
		zap.Uint64("UserId", req.UserID),
		zap.String("name", req.Name),
		zap.Int64("TargetAmount", req.TargetAmount))

	input := &savingGoalService.SgInput{
		UserID:        savingGoalDomain.UserID(req.UserID),
		Name:          req.Name,
		TargetAmount:  savingGoalDomain.TargetAmount(req.TargetAmount),
		CurrentAmount: savingGoalDomain.CurrentAmount(req.CurrentAmount),
		Deadline:      req.Deadline,
		GoalStatus:    savingGoalDomain.GoalStatus(req.GoalStatus),
		Description:   req.Description,
	}

	sg, err := h.svc.CreateGoal(r.Context(), input)
	if err != nil {
		h.logger.Error("Failed to create saving goal",
			zap.Error(err),
			zap.Uint64("UserId", req.UserID),
			zap.String("name", req.Name))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Saving goal created successfully",
		zap.Uint64("GoalId", uint64(sg.ID())),
		zap.Uint64("UserId", uint64(sg.UserID())),
		zap.String("name", sg.Name()))
	WriteJSON(w, http.StatusCreated, "Saving goal created successfully", toGoalResponse(sg))
}

func (h *SavingGoalHandler) GetAllGoals(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized get all goals: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	h.logger.Debug("Fetching all saving goals", zap.Uint64("UserId", userId))

	goals, err := h.svc.GetAllGoals(r.Context(), savingGoalDomain.UserID(userId))
	if err != nil {
		h.logger.Error("Failed to get saving goals",
			zap.Error(err),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responses := make([]GoalResponse, 0, len(goals))
	for _, sg := range goals {
		responses = append(responses, toGoalResponse(sg))
	}

	h.logger.Debug("Saving goals retrieved successfully",
		zap.Uint64("UserId", userId),
		zap.Int("count", len(responses)))
	WriteJSON(w, http.StatusOK, "Saving goals retrieved successfully", responses)
}

func (h *SavingGoalHandler) GetGoalByID(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized get goal by ID: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}
	idStr := r.PathValue("id")

	id, err1 := strconv.ParseUint(idStr, 10, 64)
	if err1 != nil {
		h.logger.Warn("Invalid goal ID format",
			zap.Error(err1),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid goal ID or user ID format", nil)
		return
	}

	h.logger.Debug("Fetching saving goal by ID",
		zap.Uint64("GoalId", id),
		zap.Uint64("UserId", userId))

	sg, err := h.svc.GetGoalByID(r.Context(), savingGoalDomain.SavingGoalID(id), savingGoalDomain.UserID(userId))
	if err != nil {
		h.logger.Warn("Saving goal not found",
			zap.Error(err),
			zap.Uint64("GoalId", id),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusNotFound, "Saving goal not found", nil)
		return
	}

	h.logger.Debug("Saving goal retrieved successfully",
		zap.Uint64("GoalId", id),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "Saving goal retrieved successfully", toGoalResponse(sg))
}

func (h *SavingGoalHandler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized update goal: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	id, err1 := strconv.ParseUint(idStr, 10, 64)
	if err1 != nil {
		h.logger.Warn("Invalid goal ID format for update",
			zap.Error(err1),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid goal ID or user ID format", nil)
		return
	}

	var req UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("Failed to decode update saving goal request",
			zap.Error(err),
			zap.Uint64("GoalId", id),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	h.logger.Debug("Updating saving goal",
		zap.Uint64("GoalId", id),
		zap.Uint64("UserId", userId))

	input := &savingGoalService.SgUpdate{
		GoalID:        savingGoalDomain.SavingGoalID(id),
		UserID:        savingGoalDomain.UserID(userId),
		Name:          req.Name,
		TargetAmount:  savingGoalDomain.TargetAmount(req.TargetAmount),
		CurrentAmount: savingGoalDomain.CurrentAmount(req.CurrentAmount),
		Deadline:      req.Deadline,
		GoalStatus:    savingGoalDomain.GoalStatus(req.GoalStatus),
		Description:   req.Description,
	}

	sg, err := h.svc.UpdateGoal(r.Context(), input, savingGoalDomain.UserID(userId))
	if err != nil {
		h.logger.Error("Failed to update saving goal",
			zap.Error(err),
			zap.Uint64("GoalId", id),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Saving goal updated successfully",
		zap.Uint64("GoalId", id),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "Saving goal updated successfully", toGoalResponse(sg))
}

func (h *SavingGoalHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		h.logger.Warn("Unauthorized delete goal: invalid user session")
		WriteJSON(w, http.StatusUnauthorized, "Unauthorized: Invalid user session", nil)
		return
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		h.logger.Warn("Invalid goal ID format for deletion",
			zap.Error(err),
			zap.String("id", idStr),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusBadRequest, "Invalid goal ID format", nil)
		return
	}

	h.logger.Debug("Deleting saving goal",
		zap.Uint64("GoalId", id),
		zap.Uint64("UserId", userId))

	err = h.svc.DeleteGoal(r.Context(), savingGoalDomain.SavingGoalID(id), savingGoalDomain.UserID(userId))
	if err != nil {
		h.logger.Error("Failed to delete saving goal",
			zap.Error(err),
			zap.Uint64("GoalId", id),
			zap.Uint64("UserId", userId))
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	h.logger.Info("Saving goal deleted successfully",
		zap.Uint64("GoalId", id),
		zap.Uint64("UserId", userId))
	WriteJSON(w, http.StatusOK, "Saving goal deleted successfully", nil)
}

func toGoalResponse(sg *savingGoalDomain.SavingGoal) GoalResponse {
	if sg == nil {
		return GoalResponse{}
	}
	return GoalResponse{
		ID:            uint64(sg.ID()),
		UserID:        uint64(sg.UserID()),
		Name:          sg.Name(),
		TargetAmount:  int64(sg.TargetAmount()),
		CurrentAmount: int64(sg.CurrentAmount()),
		Deadline:      sg.Deadline(),
		Status:        string(sg.Status()),
		Description:   sg.Description(),
		CreatedAt:     sg.CreatedAt(),
		UpdatedAt:     sg.UpdatedAt(),
	}
}
