package auth

import (
	"crypto/ecdsa"
	"errors"

	"github.com/dgrijalva/jwt-go"
)

const (
	algorithm = "ES256"
)

type SecretKey struct {
	key    *ecdsa.PrivateKey
	method jwt.SigningMethod
}

func NewSecretKey(ecdsaKey string) (*SecretKey, error) {
	key, err := jwt.ParseECPrivateKeyFromPEM([]byte(ecdsaKey))
	if err != nil {
		return nil, err
	}
	return &SecretKey{
		key:    key,
		method: jwt.GetSigningMethod(algorithm),
	}, nil
}

func (s *SecretKey) Sign(claims jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(s.method, claims)
	signed, err := token.SignedString(s.key)
	if err != nil {
		return "", err
	}
	return signed, nil
}

type PublicKey struct {
	key *ecdsa.PublicKey
}

func NewPublicKey(ecdsaKey string) (*PublicKey, error) {
	key, err := jwt.ParseECPublicKeyFromPEM([]byte(ecdsaKey))
	if err != nil {
		return nil, err
	}
	return &PublicKey{
		key: key,
	}, nil
}

func (v *PublicKey) Verify(token string, claims jwt.Claims) error {
	parsed, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return v.key, nil
	})
	if err != nil {
		return err
	}
	if !parsed.Valid {
		return errors.New("invalid token")
	}
	return nil
}
