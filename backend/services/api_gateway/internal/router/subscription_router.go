package router

import (
	"api_gateway/internal/handler"
	"net/http"
)

func RegisterSubscriptionRoutes(mux *http.ServeMux, subProxy *handler.ProxyHandler) {
	// Tüm /api/subscription/* isteklerini Subscription Service'e yönlendir
	mux.HandleFunc("/api/subscription/", subProxy.ServeHTTP)
}
