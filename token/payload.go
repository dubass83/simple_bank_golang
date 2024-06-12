package token

import (
	"errors"
	"time"

	"github.com/gofrs/uuid/v5"
)

// Different types of error returned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is unverifiable: error while executing keyfunc: unexpected signing method: none")
	ErrExpiredToken = errors.New("token has invalid claims: token is expired")
)

// Payload contain a data of the token
type Payload struct {
	ID        uuid.UUID `json:"id"`
	Username  string    `json:"username"`
	Role      string    `json:"role"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload get username and duration and create new token payload
func NewPayload(username string, role string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	payload := &Payload{
		ID:        tokenID,
		Username:  username,
		Role:      role,
		IssuedAt:  time.Now(),
		ExpiredAt: time.Now().Add(duration),
	}

	return payload, nil
}

func (payload Payload) Valid() error {
	if time.Now().After(payload.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}
