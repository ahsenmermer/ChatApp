package router

import (
	"net/http"
	"subscription_service/internal/handler"
)

func SetupSubscriptionRoutes(mux *http.ServeMux, h *handler.SubscriptionHandler) {
	// Free plan ata
	mux.HandleFunc("/api/subscription/assign_free", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.AssignFreePlan(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Belirli plan ata
	mux.HandleFunc("/api/subscription/assign_subscription", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.AssignPlan(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Kullanıcı quota
	mux.HandleFunc("/api/subscription/quota/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.GetUserQuota(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Tüm planlar
	mux.HandleFunc("/api/subscription/plans", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			h.ListPlans(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})

	// Kullanıcı event
	mux.HandleFunc("/api/subscription/event", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			h.UserEvent(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	})
}
