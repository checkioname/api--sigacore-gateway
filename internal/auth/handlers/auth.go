package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
	
	"api--sigacore-gateway/internal/auth/models"
	"api--sigacore-gateway/internal/auth/services"
	db "api--sigacore-gateway/internal/db/sqlc"
	token2 "api--sigacore-gateway/internal/token"
	"api--sigacore-gateway/internal/util"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/time/rate"
)

type AuthHandler struct {
	s           db.Store
	authService services.AuthService
	token       token2.Maker
	config      util.Config
}

func NewAuthHandler(authService services.AuthService, tokenMaker token2.Maker, connPool *pgxpool.Pool, config util.Config) *AuthHandler {
	return &AuthHandler{
		s:           db.NewStore(connPool),
		authService: authService,
		token:       tokenMaker,
		config:      config,
	}
}

func (h *AuthHandler) CreateUser(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 30*time.Second)
	defer cancel()

	var req models.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	hashed, err := util.HashPassword(req.Password)
	if err != nil {
		log.Printf("createUser: %v", err)
		c.JSON(http.StatusInternalServerError, err)
		return
	}

	args := db.CreateUserParams{
		Username:       req.Username,
		FullName:       req.FullName,
		Email:          req.Email,
		HashedPassword: hashed,
	}

	user, err := h.s.CreateUser(ctx, args)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.NewUserResponse(user))
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type loginResponse struct {
	SessionID             uuid.UUID `json:"session_id"`
	AccessToken           string    `json:"access_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshToken          string    `json:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at"`	User                  string    `json:"user"`
}

func (h *AuthHandler) LoginUser(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errResponse(c, http.StatusBadRequest, err)
		return
	}

	user, err := h.s.GetUser(c, req.Username)
	if err != nil {
		log.Printf("loginUser: %v", err)
		if err == sql.ErrNoRows {
			errResponse(c, http.StatusNotFound, err)
			return
		}
		errResponse(c, http.StatusInternalServerError, err)
		return
	}

	err = util.VerifyPassword(req.Password, user.HashedPassword)
	if err != nil {
		errResponse(c, http.StatusUnauthorized, err)
		return
	}

	accessToken, payload, err := h.token.CreateToken(user.Username, h.config.AccessTokenDuration)
	fmt.Println(payload)
	if err != nil {
		errResponse(c, http.StatusInternalServerError, err)
		return
	}

	refreshToken, refreshPayload, err := h.token.CreateToken(user.Username, h.config.AccessTokenDuration)
	fmt.Println(payload)
	if err != nil {
		errResponse(c, http.StatusInternalServerError, err)
		return
	}

	sessionParams := db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    c.Request.UserAgent(),
		ClientIp:     c.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	}
	session, err := h.s.CreateSession(c.Request.Context(), sessionParams)
	if err != nil {
		errResponse(c, http.StatusInternalServerError, err)
		return
	}

	rsp := loginResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  payload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  user.Username,
	}

	c.JSON(http.StatusOK, rsp)
}

// Handler para obter usuário (rota protegida)
func (h *AuthHandler) GetUser(c *gin.Context) {
	username := c.Param("username")

	// Verificar autorização
	authPayload := c.MustGet("authorization_payload").(*token2.Payload)
	if username != authPayload.Username {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "account doesn't belong to the authenticated user"})
		return
	}

	user, err := h.authService.GetUser(c, username)
	if err != nil {
		statusCode := http.StatusInternalServerError
		if err == sql.ErrNoRows {
			statusCode = http.StatusNotFound
		}
		c.JSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, models.NewUserResponse(user))
}
func errResponse(c *gin.Context, statusCode int, err error) {
	c.AbortWithStatusJSON(statusCode, gin.H{"error": err.Error()})
}


