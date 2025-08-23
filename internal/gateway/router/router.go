package router

import (
	"fmt"

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
	router.POST("/users", gin.HandlerFunc(func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "Auth service is healthy"})
		// fmt.Printf("DEBUG: Recebido %s %s\n", r.Method, r.URL.Path)

		// target, _ := url.Parse("http://localhost:8081")
		// proxy := httputil.NewSingleHostReverseProxy(target)

		// proxy.ServeHTTP(w, r)
	}))

	// 	func(w http.ResponseWriter, r *http.Request) {
	// }))

	fmt.Println("✅ Rota POST /users registrada")

	// Health check
	router.GET("/health", gin.HandlerFunc(func(ctx *gin.Context) {
		ctx.JSON(200, gin.H{"status": "Gateway service is healthy"})
	}))

	return router
}
