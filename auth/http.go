package auth

import (
	"net/http"
	"strings"
)

func Token(r *http.Request) string {
	token := r.Header.Get("Authorization")
	parts := strings.Split(token, " ")
	if len(parts) < 2 {
		return ""
	}
	bearer, token := parts[0], parts[1]
	if bearer != "Bearer" {
		return ""
	}
	return token
}
