package proxy

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewReverseProxy cria um handler de proxy reverso para um serviço de backend.
func NewReverseProxy(targetURL string) (http.Handler, error) {
	target, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(target), nil
} 