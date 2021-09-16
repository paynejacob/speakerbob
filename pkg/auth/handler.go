package auth

import "net/http"

type Handler struct {
	h           http.Handler
	AuthService *Service
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if _, ok := h.AuthService.VerifyRequest(r); !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	h.h.ServeHTTP(w, r)
}
