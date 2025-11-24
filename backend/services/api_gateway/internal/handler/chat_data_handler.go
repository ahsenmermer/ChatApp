package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewChatDataHandler(target string) http.Handler {
	u, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(u)
	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host
		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Host = u.Host
	}
	return proxy
}
