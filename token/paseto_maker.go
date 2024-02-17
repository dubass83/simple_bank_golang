package token

import (
	"fmt"
	"time"

	"github.com/aead/chacha20poly1305"
	"github.com/o1egl/paseto"
)

// PasetoMaker is PASETO token maker
type PasetoMaker struct {
	secreKey []byte
	paseto   *paseto.V2
}

// NewPasetoMaker function that get secret key and creates new PasetoMaker
func NewPasetoMaker(secretKey string) (Maker, error) {
	if len(secretKey) != chacha20poly1305.KeySize {
		return nil, fmt.Errorf("invalid key size: %d must be exactly %d", len(secretKey), chacha20poly1305.KeySize)
	}
	maker := &PasetoMaker{
		secreKey: []byte(secretKey),
		paseto:   paseto.NewV2(),
	}

	return maker, nil
}

// CreateToken create new token for username and duration
func (maker PasetoMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}

	// Encrypt data
	token, err := maker.paseto.Encrypt(maker.secreKey, payload, nil)
	if err != nil {
		return "", payload, err
	}
	return token, payload, nil
}

// VerifyToken check is token valid or not
func (maker PasetoMaker) VerifyToken(token string) (*Payload, error) {
	// Decrypt data
	var payload Payload
	err := maker.paseto.Decrypt(token, maker.secreKey, &payload, nil)
	if err != nil {
		return nil, ErrInvalidToken
	}

	err = payload.Valid()
	if err != nil {
		return nil, err
	}

	return &payload, nil
}
