package token

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var ErrExpiredToken = errors.New("expired token")

type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expires"`
}

func (p *Payload) GetExpirationTime() (*jwt.NumericDate, error) {
	//TODO implement me
	return &jwt.NumericDate{Time: p.ExpiredAt}, nil
}

func (p *Payload) GetIssuedAt() (*jwt.NumericDate, error) {
	//TODO implement me
	return &jwt.NumericDate{Time: p.IssuedAt}, nil
}

func (p *Payload) GetNotBefore() (*jwt.NumericDate, error) {
	//TODO implement me
	return &jwt.NumericDate{Time: p.IssuedAt}, nil
}

func (p *Payload) GetIssuer() (string, error) {
	//TODO implement me
	return "", nil
}

func (p *Payload) GetSubject() (string, error) {
	//TODO implement me
	return "", nil
}

func (p *Payload) GetAudience() (jwt.ClaimStrings, error) {
	//TODO implement me
	return jwt.ClaimStrings{}, nil
}

func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewUUID()
	if err != nil {
		return nil, fmt.Errorf("NewPayload: %v", err)
	}
	return &Payload{
		ID:        tokenID,
		Username:  username,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}, nil
}

func (p *Payload) Valid() error {
	if time.Now().After(p.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
