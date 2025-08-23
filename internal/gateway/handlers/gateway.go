package handlers

import (
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/gin-gonic/gin"

	"api--sigacore-gateway/internal/util"
)

type GatewayHandler struct {
	config           util.Config
	authServiceProxy *httputil.ReverseProxy
	userServiceProxy *httputil.ReverseProxy
}

func NewGatewayHandler(cfg util.Config) *GatewayHandler {
	authServiceURL := "http://" + cfg.AuthServerAddress
	authTarget, _ := url.Parse(authServiceURL)
	authServiceProxy := httputil.NewSingleHostReverseProxy(authTarget)

	userTarget, _ := url.Parse(cfg.UserServiceAddress)
	userServiceProxy := httputil.NewSingleHostReverseProxy(userTarget)

	return &GatewayHandler{
		config:           cfg,
		authServiceProxy: authServiceProxy,
		userServiceProxy: userServiceProxy,
	}
}

func (h *GatewayHandler) ProxyToAuthService(c *gin.Context) {
	// Remove o prefixo /auth da URL
	c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/auth")
	h.authServiceProxy.ServeHTTP(c.Writer, c.Request)
}

func (h *GatewayHandler) ProxyToUserService(c *gin.Context) {
	// Remove o prefixo /users da URL
	c.Request.URL.Path = strings.TrimPrefix(c.Request.URL.Path, "/users")
	h.userServiceProxy.ServeHTTP(c.Writer, c.Request)
}
