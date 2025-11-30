// internal/handler/base_handler.go
package handler

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// ProxyHandler provides reverse proxy functionality for a service
type ProxyHandler struct {
	proxy       *httputil.ReverseProxy
	targetURL   string
	serviceName string
}

// NewProxyHandler creates a new proxy handler for a backend service
func NewProxyHandler(targetURL, serviceName string) *ProxyHandler {
	u, err := url.Parse(targetURL)
	if err != nil {
		log.Fatalf("❌ Invalid target URL for %s: %v", serviceName, err)
	}

	proxy := httputil.NewSingleHostReverseProxy(u)

	// Custom director for headers
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Header.Set("X-Forwarded-Proto", "http")
		req.Header.Set("X-Forwarded-For", req.RemoteAddr)
		req.Host = u.Host
	}

	// Error handler for proxy failures
	proxy.ErrorHandler = func(w http.ResponseWriter, r *http.Request, err error) {
		log.Printf("❌ [%s] Proxy error: %v", serviceName, err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error":"service unavailable"}`))
	}

	return &ProxyHandler{
		proxy:       proxy,
		targetURL:   targetURL,
		serviceName: serviceName,
	}
}

// ServeHTTP forwards the request as-is (implements http.Handler)
func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	log.Printf("🔀 [%s] %s %s", h.serviceName, r.Method, r.URL.Path)
	h.proxy.ServeHTTP(w, r)
}

// Forward forwards the request with custom path
func (h *ProxyHandler) Forward(targetPath string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔀 [%s] %s %s → %s", h.serviceName, r.Method, r.URL.Path, targetPath)

		// Change path temporarily
		originalPath := r.URL.Path
		r.URL.Path = targetPath

		h.proxy.ServeHTTP(w, r)

		// Restore path
		r.URL.Path = originalPath
	}
}
