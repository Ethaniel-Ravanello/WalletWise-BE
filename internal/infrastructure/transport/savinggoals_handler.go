package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"walletwise/internal/application/saving_goal"
	"walletwise/internal/domain/saving_goals"
)

// --- DTO Request & Response ---

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

// --- Handler / Controller ---

type SavingGoalsHandler struct {
	service *saving_goal.Service
}

func NewSavingGoalsHandler(service *saving_goal.Service) *SavingGoalsHandler {
	return &SavingGoalsHandler{service: service}
}

// 1. Create Goal (POST)
func (h *SavingGoalsHandler) CreateGoal(w http.ResponseWriter, r *http.Request) {
	var req CreateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid request payload", nil)
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

	sg, err := h.service.CreateTransaction(r.Context(), input) // Asumsi nama method di service masih CreateTransaction
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	res := GoalResponse{
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

	WriteJson(w, http.StatusCreated, "Success Create Goal", res)
}

// 2. Get All Goals (GET)
func (h *SavingGoalsHandler) GetAllGoals(w http.ResponseWriter, r *http.Request) {
	userIDStr := r.URL.Query().Get("user_id")
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)

	goals, err := h.service.GetAllGoals(r.Context(), saving_goals.UserID(userID))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	var responses []GoalResponse
	for _, sg := range goals {
		responses = append(responses, GoalResponse{
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
		})
	}

	if responses == nil {
		responses = []GoalResponse{}
	}

	WriteJson(w, http.StatusOK, "Success Get Goals List", responses)
}

// 3. Get Goal By ID (GET)
func (h *SavingGoalsHandler) GetGoalByID(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	userIDStr := r.URL.Query().Get("user_id")

	id, _ := strconv.ParseUint(idStr, 10, 64)
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)

	sg, err := h.service.GetGoalByID(r.Context(), saving_goals.SavingGoalsID(id), saving_goals.UserID(userID))
	if err != nil {
		WriteJson(w, http.StatusNotFound, "Goal not found", nil)
		return
	}

	res := GoalResponse{
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

	WriteJson(w, http.StatusOK, "Success Get Goal Detail", res)
}

// 4. Update Goal (PUT)
func (h *SavingGoalsHandler) UpdateGoal(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	userIDStr := r.URL.Query().Get("user_id")

	id, _ := strconv.ParseUint(idStr, 10, 64)
	userID, _ := strconv.ParseUint(userIDStr, 10, 64)

	var req UpdateGoalRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		WriteJson(w, http.StatusBadRequest, "Invalid request payload", nil)
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

	sg, err := h.service.UpdateGoal(r.Context(), input)
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	res := GoalResponse{
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

	WriteJson(w, http.StatusOK, "Success Update Goal", res)
}

// 5. Delete Goal (DELETE)
func (h *SavingGoalsHandler) DeleteGoal(w http.ResponseWriter, r *http.Request) {
	idStr := r.URL.Query().Get("id")
	id, _ := strconv.ParseUint(idStr, 10, 64)

	err := h.service.DeleteGoal(r.Context(), saving_goals.SavingGoalsID(id))
	if err != nil {
		WriteJson(w, http.StatusInternalServerError, err.Error(), nil)
		return
	}

	WriteJson(w, http.StatusOK, "Success Delete Goal", nil)
}
