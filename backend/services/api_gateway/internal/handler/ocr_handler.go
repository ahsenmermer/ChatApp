package handler

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func NewOCRHandler(target string) http.Handler {
	u, _ := url.Parse(target)
	proxy := httputil.NewSingleHostReverseProxy(u)

	proxy.Director = func(req *http.Request) {
		req.URL.Scheme = u.Scheme
		req.URL.Host = u.Host

		// /api/upload -> /api/upload (p<ath'i koru)
		// OCR servisi /api/upload dinliyor, bu yüzden path'i değiştirmeye gerek yok
		req.URL.Path = req.URL.Path

		req.Header.Set("X-Forwarded-Host", req.Host)
		req.Host = u.Host
	}

	return proxy
}
