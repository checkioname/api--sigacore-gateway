package router

import (
	"fmt"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"

	"api--sigacore-gateway/internal/util"
)

// func setupGatewayMiddlewares() error {
// 	r := gin.Default()
// 	// Middleware global de whitelist de IPs
// 	// ipWhitelistMiddleware := middleware.IPWhitelist(r.config.AllowedIPs...)

// 	// Middleware global de rate limiting
// 	rateLimiterMiddleware := middleware.RateLimiter(rate.Limit(5), 10) // 5 req/s, burst de 10

// 	// r.router.Use(ipWhitelistMiddleware)
// 	r.Use(rateLimiterMiddleware)

// 	return nil
// }

func SetupGatewayRoutes(cfg util.Config) *gin.Engine {
	router := gin.Default()

	// proxies
	router.Any("/*users", gin.HandlerFunc(func(ctx *gin.Context) {
		target, _ := url.Parse("http://localhost:8081")
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}))
	fmt.Println("✅ Rota POST /users registrada")

	// Health check
	// router.GET("/health", gin.HandlerFunc(func(ctx *gin.Context) {
	// 	ctx.JSON(200, gin.H{"status": "Gateway service is healthy"})
	// }))

	return router
}
