package middleware

import (
	"api--sigacore-gateway/internal/token"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	authHeader     = "Authorization"
	authType       = "bearer"
	authPayloadKey = "authorization_payload"
)

var (
	ErrInvalidHeader   = errors.New("invalid authorization header")
	ErrInvalidAuthType = errors.New("invalid authorization type")
)

func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader(authHeader)
		if authHeader == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidHeader.Error()})
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidHeader.Error()})
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != authType {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": ErrInvalidAuthType.Error()})
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		ctx.Set(authPayloadKey, payload)
		ctx.Next()
	}
}
