package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"
	"walletwise/internal/middleware"

	savingGoalService "walletwise/internal/application/saving_goal"
	savingGoalDomain "walletwise/internal/domain/saving_goal"
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
	svc *savingGoalService.Service
}

type SavingGoalsHandler = SavingGoalHandler

func NewSavingGoalHandler(svc *savingGoalService.Service) *SavingGoalHandler {
	return &SavingGoalHandler{svc: svc}
}

func (h *SavingGoalHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.UserID == 0 || req.Name == "" {
		WriteJSON(w, http.StatusBadRequest, "user_id and name are required", nil)
		return
	}

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
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "Saving goal created successfully", toGoalResponse(sg))
}

func (h *SavingGoalHandler) GetAllGoals(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}

	goals, err := h.svc.GetAllGoals(r.Context(), savingGoalDomain.UserID(userId))
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	responses := make([]GoalResponse, 0, len(goals))
	for _, sg := range goals {
		responses = append(responses, toGoalResponse(sg))
	}

	WriteJSON(w, http.StatusOK, "Saving goals retrieved successfully", responses)
}

func (h *SavingGoalHandler) GetGoalByID(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}
	idStr := r.PathValue("id")

	id, err1 := strconv.ParseUint(idStr, 10, 64)

	if err1 != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid goal ID or user ID format", nil)
		return
	}

	sg, err := h.svc.GetGoalByID(r.Context(), savingGoalDomain.SavingGoalID(id), savingGoalDomain.UserID(userId))
	if err != nil {
		WriteJSON(w, http.StatusNotFound, "Saving goal not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Saving goal retrieved successfully", toGoalResponse(sg))
}

func (h *SavingGoalHandler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}

	idStr := r.PathValue("id")
	id, err1 := strconv.ParseUint(idStr, 10, 64)
	if err1 != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid goal ID or user ID format", nil)
		return
	}

	var req UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

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
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Saving goal updated successfully", toGoalResponse(sg))
}

func (h *SavingGoalHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	userIdCtx := r.Context().Value(middleware.UserIdKey)
	userId, ok := userIdCtx.(uint64)
	if !ok {
		WriteJSON(w, http.StatusBadRequest, "Unauthorized: Invalid user session", nil)
	}

	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid goal ID format", nil)
		return
	}

	err = h.svc.DeleteGoal(r.Context(), savingGoalDomain.SavingGoalID(id), savingGoalDomain.UserID(userId))
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

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
