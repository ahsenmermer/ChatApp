package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"subscription_service/internal/services"

	"github.com/google/uuid"
)

type SubscriptionHandler struct {
	service *services.UserSubscriptionService
}

func NewSubscriptionHandler(service *services.UserSubscriptionService) *SubscriptionHandler {
	return &SubscriptionHandler{service: service}
}

// AssignFreePlan assigns Free plan to a user
func (h *SubscriptionHandler) AssignFreePlan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	if err := h.service.AssignFreePlanToUser(uid); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success"}`))
}

// AssignPlan assigns a specific plan to a user
func (h *SubscriptionHandler) AssignPlan(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Plan   string `json:"plan"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	uid, err := uuid.Parse(req.UserID)
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	if err := h.service.AssignPlanToUserByName(uid, req.Plan); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte(`{"status":"success"}`))
}

// GetUserQuota returns remaining quota
func (h *SubscriptionHandler) GetUserQuota(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 {
		http.Error(w, "user_id required", http.StatusBadRequest)
		return
	}

	uid, err := uuid.Parse(parts[4])
	if err != nil {
		http.Error(w, "invalid user_id", http.StatusBadRequest)
		return
	}

	quota, err := h.service.GetUserQuota(uid)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(map[string]int{"quota": quota})
}

// ListPlans lists all subscription plans
func (h *SubscriptionHandler) ListPlans(w http.ResponseWriter, r *http.Request) {
	plans, err := h.service.ListAllPlans()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode(plans)
}

// UserEvent logs user event and decrements quota
func (h *SubscriptionHandler) UserEvent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID string `json:"user_id"`
		Event  string `json:"event"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	if err := h.service.LogEventAndDecrementQuota(req.UserID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"ok"}`))
}
