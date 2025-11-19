package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewChatDataProxy(target string) http.Handler {
	u, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		if req.Header.Get("X-Forwarded-Host") == "" {
			req.Header.Set("X-Forwarded-Host", req.Host)
		}
		req.Host = u.Host
	}
	return proxy
}
