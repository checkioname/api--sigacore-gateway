package token

import (
	"fmt"
	"time"

	"github.com/o1egl/paseto"
	"golang.org/x/crypto/chacha20"
)

type PasetoMaker struct {
	paseto       *paseto.V2
	symmetricKey []byte
}

func NewPasetoMaker(symmetricKey string) (Maker, error) {
	if len(symmetricKey) != chacha20.KeySize {
		return nil, fmt.Errorf("NewPasetoMaker: invalid key size")
	}
	return &PasetoMaker{
		paseto:       paseto.NewV2(),
		symmetricKey: []byte(symmetricKey),
	}, nil
}

func (pm *PasetoMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", nil, fmt.Errorf("CreateToken: %w", err)
	}

	token, err := pm.paseto.Encrypt(pm.symmetricKey, payload, nil)
	if err != nil {
		return "", nil, fmt.Errorf("CreateToken: %w", err)
	}
	return token, payload, nil
}

func (pm *PasetoMaker) VerifyToken(token string) (*Payload, error) {
	payload := &Payload{}
	err := pm.paseto.Decrypt(token, pm.symmetricKey, payload, nil)
	if err != nil {
		return nil, fmt.Errorf("VerifyToken: %w", err)
	}

	err = payload.Valid()
	if err != nil {
		return nil, fmt.Errorf("VerifyToken: %w", err)
	}

	return payload, nil
}
