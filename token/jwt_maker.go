package token

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/golang-jwt/jwt/v5"
)

const minSecretKey = 32

type JwtMaker struct {
	secretKey string
}

func NewJwtMaker(secretKey string) (Maker, error) {
	if len(secretKey) < minSecretKey {
		return nil, fmt.Errorf("secret key length less the %d", minSecretKey)
	}
	return &JwtMaker{secretKey}, nil
}

// CreateToken create new token for username and duration
func (maker JwtMaker) CreateToken(username string, duration time.Duration) (string, *Payload, error) {
	payload, err := NewPayload(username, duration)
	if err != nil {
		return "", payload, err
	}
	// prepare claim for jwt from payload struct
	claim := &jwt.RegisteredClaims{
		ID:        payload.ID.String(),
		Subject:   payload.Username,
		IssuedAt:  jwt.NewNumericDate(payload.IssuedAt),
		ExpiresAt: jwt.NewNumericDate(payload.ExpiredAt),
	}

	jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)
	ss, err := jwtToken.SignedString([]byte(maker.secretKey))
	if err != nil {
		return "", payload, err
	}
	return ss, payload, nil
}

// VerifyToken check is token valid or not
func (maker JwtMaker) VerifyToken(tokenString string) (*Payload, error) {
	// Parse takes the token string and a function for looking up the key. The latter is especially
	// useful if you use multiple keys for your application.  The standard is to use 'kid' in the
	// head of the token to identify which key to use, but the parsed token (head and claims) is provided
	// to the callback, providing flexibility.
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		// hmacSampleSecret is a []byte containing your secret, e.g. []byte("my_secret_key")
		return []byte(maker.secretKey), nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("can not cast token.Claims to jwt.MapClaims struct %v", token.Claims)
	}
	// convert claims["iat"] to golang time.Time type
	var tiat time.Time
	switch iat := claims["iat"].(type) {
	case float64:
		tiat = time.Unix(int64(iat), 0)
	case json.Number:
		v, _ := iat.Int64()
		tiat = time.Unix(v, 0)
	}
	// convert claims["exp"] to golang time.Time type
	var texp time.Time
	switch exp := claims["exp"].(type) {
	case float64:
		texp = time.Unix(int64(exp), 0)
	case json.Number:
		v, _ := exp.Int64()
		texp = time.Unix(v, 0)
	}

	return &Payload{
		ID:        uuid.Must(uuid.FromString(claims["jti"].(string))),
		Username:  claims["sub"].(string),
		IssuedAt:  tiat,
		ExpiredAt: texp,
	}, nil
}
