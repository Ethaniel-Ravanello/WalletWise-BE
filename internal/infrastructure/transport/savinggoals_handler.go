package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"walletwise/internal/application/saving_goal"
	"walletwise/internal/domain/saving_goals"
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

type SavingGoalsHandler struct {
	svc *saving_goal.Service
}

func NewSavingGoalsHandler(svc *saving_goal.Service) *SavingGoalsHandler {
	return &SavingGoalsHandler{svc: svc}
}

func (h *SavingGoalsHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	if req.UserID == 0 || req.Name == "" {
		WriteJSON(w, http.StatusBadRequest, "user_id and name are required", nil)
		return
	}

	input := &saving_goal.SgInput{
		UserID:        saving_goals.UserID(req.UserID),
		Name:          req.Name,
		TargetAmount:  saving_goals.TargetAmount(req.TargetAmount),
		CurrentAmount: saving_goals.CurrentAmount(req.CurrentAmount),
		Deadline:      req.Deadline,
		GoalStatus:    saving_goals.GoalStatus(req.GoalStatus),
		Description:   req.Description,
	}

	sg, err := h.svc.CreateGoal(r.Context(), input)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusCreated, "Saving goal created successfully", toGoalResponse(sg))
}

func (h *SavingGoalsHandler) GetAllGoals(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("userId")
	}
	if userIDStr == "" {
		WriteJSON(w, http.StatusBadRequest, "user_id query parameter is required", nil)
		return
	}

	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid user_id format", nil)
		return
	}

	goals, err := h.svc.GetAllGoals(r.Context(), saving_goals.UserID(userID))
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

func (h *SavingGoalsHandler) GetGoalByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("userId")
	}

	id, err1 := strconv.ParseUint(idStr, 10, 64)
	userID, err2 := strconv.ParseUint(userIDStr, 10, 64)
	if err1 != nil || err2 != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid goal ID or user ID format", nil)
		return
	}

	sg, err := h.svc.GetGoalByID(r.Context(), saving_goals.SavingGoalsID(id), saving_goals.UserID(userID))
	if err != nil {
		WriteJSON(w, http.StatusNotFound, "Saving goal not found", nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Saving goal retrieved successfully", toGoalResponse(sg))
}

func (h *SavingGoalsHandler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	userIDStr := r.URL.Query().Get("user_id")
	if userIDStr == "" {
		userIDStr = r.URL.Query().Get("userId")
	}

	id, err1 := strconv.ParseUint(idStr, 10, 64)
	userID, err2 := strconv.ParseUint(userIDStr, 10, 64)
	if err1 != nil || err2 != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid goal ID or user ID format", nil)
		return
	}

	var req UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid request payload", nil)
		return
	}

	input := &saving_goal.SgUpdate{
		GoalID:        saving_goals.SavingGoalsID(id),
		UserID:        saving_goals.UserID(userID),
		Name:          req.Name,
		TargetAmount:  saving_goals.TargetAmount(req.TargetAmount),
		CurrentAmount: saving_goals.CurrentAmount(req.CurrentAmount),
		Deadline:      req.Deadline,
		GoalStatus:    saving_goals.GoalStatus(req.GoalStatus),
		Description:   req.Description,
	}

	sg, err := h.svc.UpdateGoal(r.Context(), input)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Saving goal updated successfully", toGoalResponse(sg))
}

func (h *SavingGoalsHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "Invalid goal ID format", nil)
		return
	}

	err = h.svc.DeleteGoal(r.Context(), saving_goals.SavingGoalsID(id))
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJSON(w, http.StatusOK, "Saving goal deleted successfully", nil)
}

func toGoalResponse(sg *saving_goals.SavingGoals) GoalResponse {
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


