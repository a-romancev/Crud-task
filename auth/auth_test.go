package auth

import (
	"encoding/base64"
	"encoding/json"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	// nolint: gosec
	secretKey = `-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIB8fmVWhMdAo/UkDNN4UGo8PYwKxz/lN7nilmYa2KEkboAoGCCqGSM49
AwEHoUQDQgAETrMd0Br7GOpE7US1jJ7LbL0L8vIi3NxRxnXhOxDWaAhd4MxdF17f
AY5OGjJpPdWJ8TDMQH7Es98SAB9pVRVZhg==
-----END EC PRIVATE KEY-----`
	// nolint: gosec
	publicKey = `-----BEGIN PUBLIC KEY-----
MFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAETrMd0Br7GOpE7US1jJ7LbL0L8vIi
3NxRxnXhOxDWaAhd4MxdF17fAY5OGjJpPdWJ8TDMQH7Es98SAB9pVRVZhg==
-----END PUBLIC KEY-----`
)

func TestAPIToken(t *testing.T) {
	t.Parallel()

	sk, err := NewSecretKey(secretKey)
	require.NoError(t, err)
	pk, err := NewPublicKey(publicKey)
	require.NoError(t, err)

	t.Run("Happy path", func(t *testing.T) {
		userID := uuid.New()
		token, err := sk.Sign(NewAPIClaims(userID))
		require.NoError(t, err)
		var claims APIClaims
		err = pk.Verify(token, &claims)
		require.NoError(t, err)
		assert.Equal(t, userID, claims.UserID)
		assert.NotZero(t, claims.ExpiresAt)
	})

	t.Run("Invalid signature returns error", func(t *testing.T) {
		userID := uuid.New()
		token, err := sk.Sign(NewAPIClaims(userID))
		require.NoError(t, err)
		parts := strings.Split(token, ".")

		// Replacing the email in the token
		body, _ := base64.StdEncoding.DecodeString(parts[1])
		var claims APIClaims
		_ = json.Unmarshal(body, &claims)
		claims.UserID = uuid.New()
		body, _ = json.Marshal(claims)
		parts[2] = base64.StdEncoding.EncodeToString(body)
		token = strings.Join(parts, ".")

		err = pk.Verify(token, nil)
		assert.Error(t, err)
	})
}
