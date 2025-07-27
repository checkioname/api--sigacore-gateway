package handlers

import (
	"net/http"

	"api--sigacore-gateway/internal/auth/models"

	"github.com/gin-gonic/gin"
)

// Handler para renovar access token
func (h *AuthHandler) RenewAccessToken(c *gin.Context) {
	var req models.RenewAccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	payload, err := h.token.VerifyToken(req.RefreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	response, err := h.authService.RenewAccessToken(c, payload, req.RefreshToken)
	if err != nil {
		statusCode := http.StatusInternalServerError
		switch err.Error() {
		case "session not found":
			statusCode = http.StatusNotFound
		case "session blocked", "incorrect session", "session expired":
			statusCode = http.StatusUnauthorized
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, response)
}
