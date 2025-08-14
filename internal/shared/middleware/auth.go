package middleware

import (
	"api--sigacore-gateway/internal/token"
	"errors"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	_authHeaderKey  = "Authorization"
	_authTypeBearer = "bearer"
	_authPayloadKey = "authorization_payload"
)

var (
	_errMissingAuthHeader   = errors.New("authorization header is required")
	_errInvalidAuthFormat   = errors.New("invalid authorization header format")
	_errUnsupportedAuthType = errors.New("unsupported authorization type")
)

// AuthMiddleware creates an authentication middleware for the specified framework.
func AuthMiddleware(tokenMaker token.Maker) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader(_authHeaderKey)
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized,
				gin.H{"error": _errMissingAuthHeader.Error()})
			return
		}

		fields := strings.Fields(authHeader)
		if len(fields) < 2 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": _errInvalidAuthFormat.Error()})
			return
		}

		authorizationType := strings.ToLower(fields[0])
		if authorizationType != _authTypeBearer {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": _errUnsupportedAuthType.Error()})
			return
		}

		accessToken := fields[1]
		payload, err := tokenMaker.VerifyToken(accessToken)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}

		c.Set(_authPayloadKey, payload)
		c.Next()
	}
}
