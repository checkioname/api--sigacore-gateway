package api

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ErrSessionBlocked   = errors.New("session blocked")
	ErrIncorrectSession = errors.New("incorrect session")
)

type renewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token" binding:"required"`
}

type renewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

func (s *Server) renewAccessToken(c *gin.Context) {
	var req renewAccessTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResponse(c, http.StatusBadRequest, err)
		return
	}

	payload, err := s.token.VerifyToken(req.RefreshToken)
	if err != nil {
		errResponse(c, http.StatusUnauthorized, err)
		return
	}

	session, err := s.store.GetSession(c, payload.ID)
	if err != nil {
		log.Printf("renewAccessTokenUser: %v", err)
		if err == sql.ErrNoRows {
			errResponse(c, http.StatusNotFound, err)
			return
		}
		errResponse(c, http.StatusInternalServerError, err)
		return
	}

	if session.IsBlocked {
		errResponse(c, http.StatusUnauthorized, ErrSessionBlocked)
		return
	}

	if session.Username != payload.Username {
		errResponse(c, http.StatusUnauthorized, ErrIncorrectSession)
		return
	}

	if session.RefreshToken != req.RefreshToken {
		errResponse(c, http.StatusUnauthorized, ErrIncorrectSession)
		return
	}

	if time.Now().After(session.ExpiresAt) {
		err := fmt.Errorf("session expired")
		errResponse(c, http.StatusUnauthorized, err)
		return
	}

	accessToken, payload, err := s.token.CreateToken(payload.Username, s.config.AccessTokenDuration)
	fmt.Println(payload)
	if err != nil {
		errResponse(c, http.StatusInternalServerError, err)
		return
	}

	rsp := renewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: payload.ExpiredAt,
	}

	c.JSON(http.StatusOK, rsp)
}
