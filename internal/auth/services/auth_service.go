package services

import (
	db "api--sigacore-gateway/internal/db/sqlc"
	token2 "api--sigacore-gateway/internal/token"
	"api--sigacore-gateway/internal/util"
	"context"
	"database/sql"
	"fmt"
	"time"

	"api--sigacore-gateway/internal/auth/models"

	"github.com/gin-gonic/gin"
)

type AuthService interface {
	CreateUser(ctx context.Context, req models.CreateUserRequest) (db.User, error)
	LoginUser(ctx *gin.Context, req models.LoginUserRequest) (models.LoginUserResponse, error)
	GetUser(ctx context.Context, username string) (db.User, error)
	RenewAccessToken(ctx context.Context, payload *token2.Payload, refreshToken string) (models.RenewAccessTokenResponse, error)
}

type authService struct {
	store      db.Store
	tokenMaker token2.Maker
	config     util.Config
}

func NewAuthService(store db.Store, tokenMaker token2.Maker, config util.Config) AuthService {
	return &authService{
		store:      store,
		tokenMaker: tokenMaker,
		config:     config,
	}
}

func (s *authService) CreateUser(ctx context.Context, req models.CreateUserRequest) (db.User, error) {
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return db.User{}, err
	}

	arg := db.CreateUserParams{
		Username:       req.Username,
		HashedPassword: hashedPassword,
		FullName:       req.FullName,
		Email:          req.Email,
		IsWhitelisted:  req.IsWhitelisted,
	}

	return s.store.CreateUser(ctx, arg)
}

func (s *authService) LoginUser(ctx *gin.Context, req models.LoginUserRequest) (models.LoginUserResponse, error) {
	user, err := s.store.GetUser(ctx, req.Username)
	if err != nil {
		return models.LoginUserResponse{}, err
	}

	// Verificar se o usuário está na whitelist
	if !user.IsWhitelisted {
		return models.LoginUserResponse{}, fmt.Errorf("user not whitelisted")
	}

	err = util.VerifyPassword(req.Password, user.HashedPassword)
	if err != nil {
		return models.LoginUserResponse{}, fmt.Errorf("invalid credentials")
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(
		user.Username,
		s.config.AccessTokenDuration,
	)
	if err != nil {
		return models.LoginUserResponse{}, err
	}

	refreshToken, refreshPayload, err := s.tokenMaker.CreateToken(
		user.Username,
		s.config.RefreshTokenDuration,
	)
	if err != nil {
		return models.LoginUserResponse{}, err
	}

	session, err := s.store.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshPayload.ID,
		Username:     user.Username,
		RefreshToken: refreshToken,
		UserAgent:    ctx.Request.UserAgent(),
		ClientIp:     ctx.ClientIP(),
		IsBlocked:    false,
		ExpiresAt:    refreshPayload.ExpiredAt,
	})
	if err != nil {
		return models.LoginUserResponse{}, err
	}

	return models.LoginUserResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		AccessTokenExpiresAt:  accessPayload.ExpiredAt,
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: refreshPayload.ExpiredAt,
		User:                  models.NewUserResponse(user),
	}, nil
}

func (s *authService) GetUser(ctx context.Context, username string) (db.User, error) {
	return s.store.GetUser(ctx, username)
}

func (s *authService) RenewAccessToken(ctx context.Context, payload *token2.Payload, refreshToken string) (models.RenewAccessTokenResponse, error) {
	session, err := s.store.GetSession(ctx, payload.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			return models.RenewAccessTokenResponse{}, fmt.Errorf("session not found")
		}
		return models.RenewAccessTokenResponse{}, err
	}

	if session.IsBlocked {
		return models.RenewAccessTokenResponse{}, fmt.Errorf("session blocked")
	}

	if session.Username != payload.Username {
		return models.RenewAccessTokenResponse{}, fmt.Errorf("incorrect session")
	}

	if session.RefreshToken != refreshToken {
		return models.RenewAccessTokenResponse{}, fmt.Errorf("incorrect session")
	}

	if time.Now().After(session.ExpiresAt) {
		return models.RenewAccessTokenResponse{}, fmt.Errorf("session expired")
	}

	accessToken, accessPayload, err := s.tokenMaker.CreateToken(payload.Username, s.config.AccessTokenDuration)
	if err != nil {
		return models.RenewAccessTokenResponse{}, err
	}

	return models.RenewAccessTokenResponse{
		AccessToken:          accessToken,
		AccessTokenExpiresAt: accessPayload.ExpiredAt,
	}, nil
}
