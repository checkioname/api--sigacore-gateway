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

	// Configurar CORS
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// proxies
	router.Any("/*users", gin.HandlerFunc(func(ctx *gin.Context) {
		target, _ := url.Parse("http://localhost:8081")
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}))
	fmt.Println("✅ Rota POST /users registrada")

	// // Health check
	// router.GET("/health", gin.HandlerFunc(func(ctx *gin.Context) {
	// 	ctx.JSON(200, gin.H{"status": "Gateway service is healthy"})
	// }))

	return router
}
