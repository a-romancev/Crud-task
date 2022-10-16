package auth

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/google/uuid"
)

const (
	apiExpire = 15 * time.Minute
)

type APIClaims struct {
	jwt.StandardClaims
	UserID uuid.UUID `json:"uid"`
}

func NewAPIClaims(userID uuid.UUID) *APIClaims {
	return &APIClaims{
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(apiExpire).Unix(),
		},
		UserID: userID,
	}
}

func (c APIClaims) Valid() error {
	return c.StandardClaims.Valid()
}

func (c APIClaims) ExpiresIn() time.Duration {
	return apiExpire
}
