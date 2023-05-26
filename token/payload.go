package token

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Different types of error returned by the VerifyToken function
var (
	ErrInvalidToken = errors.New("token is invalid")
	ErrExpiredToken = errors.New("token has expired")
)

// Payload contains the payload data of the token
type Payload struct {
	jwt.RegisteredClaims

	ID       uuid.UUID `json:"id"`
	Username string    `json:"username"`
	// IssuedAt  time.Time `json:"issued_at"`
	// ExpiredAt time.Time `json:"expired_at"`
}

// NewPayload creates a new token payload with a specific username and duration
func NewPayload(username string, duration time.Duration) (*Payload, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	payload := &Payload{
		RegisteredClaims: jwt.RegisteredClaims{
			// Also fixed dates can be used for the NumericDate
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
		ID:       tokenID,
		Username: username,
		// IssuedAt:  time.Now(),
		// ExpiredAt: time.Now().Add(duration),

	}
	return payload, nil
}

// Valid checks if the token payload is valid or not
func (payload *Payload) Validate() error {
	if time.Now().After(payload.ExpiresAt.Time) {
		return ErrExpiredToken
	}
	return nil
}
