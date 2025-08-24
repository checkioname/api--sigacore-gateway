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

	// Configurar CORS - mais específico para evitar problemas com preflight
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "http://localhost:3000")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400") // Cache preflight por 24h

		// Responder imediatamente para OPTIONS (preflight)
		if c.Request.Method == "OPTIONS" {
			fmt.Printf("🎯 CORS: OPTIONS preflight para %s\n", c.Request.URL.Path)
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// proxies
	router.Any("/users/*path", gin.HandlerFunc(func(ctx *gin.Context) {
		target, _ := url.Parse("http://localhost:8081")
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}))

	// client service
	router.Any("/clientes/*path", gin.HandlerFunc(func(ctx *gin.Context) {
		// O prefixo do destino deve ser a rota correta.
		targetPath := "http://localhost:8082"

		target, _ := url.Parse(targetPath)
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}))

	router.GET("/clientes", gin.HandlerFunc(func(ctx *gin.Context) {
		targetPath := `http://localhost:8082`

		target, _ := url.Parse(targetPath)
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}))

	// mapping apenas p criacao de doc (ja que eh outra api)
	router.Any("/docs/*path", gin.HandlerFunc(func(ctx *gin.Context) {
		target, _ := url.Parse("http://localhost:8083")
		proxy := httputil.NewSingleHostReverseProxy(target)

		proxy.ServeHTTP(ctx.Writer, ctx.Request)
	}))

	// Health check
	// router.GET("/health", gin.HandlerFunc(func(ctx *gin.Context) {
	// 	ctx.JSON(200, gin.H{"status": "Gateway service is healthy"})
	// }))

	return router
}
