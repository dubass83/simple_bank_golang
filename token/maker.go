package token

import (
	"time"
)

type Maker interface {
	// CreateToken create new token for username and duration
	CreateToken(username string, role string, duration time.Duration) (string, *Payload, error)
	// VerifyToken check is token valid or not
	VerifyToken(token string) (*Payload, error)
}
