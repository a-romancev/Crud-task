package main

import (
	"net/http"
)

type Handler struct {
	router http.Handler
}

func NewHandler() *Handler {
	r := http.NewServeMux()

	r.HandleFunc("/health", health)
	return &Handler{router: r}
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	defer func() {
		if err := recover(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}()
	h.router.ServeHTTP(w, r.WithContext(ctx))
}

func health(w http.ResponseWriter, _ *http.Request) {
	_, _ = w.Write([]byte("ok"))
}
