package authentication

import (
	"context"
	"encoding/json"
	"net/http"
	"speakerbob/internal"
)

func AuthenticationMiddleware(rw http.ResponseWriter, r *http.Request, next http.HandlerFunc) {
	cookie, err := r.Cookie(AuthCookieName)

	// Fail if no cookie is found or the cookie value does not exist in redis
	if err != nil {
		var resp = UnauthenticatedResponse{"You must be authenticated to preform this action"}
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(resp)
	}

	userId := internal.GetRedisClient().Get(cookie.Value).Val()
	if userId == "" {
		var resp = UnauthenticatedResponse{"You must be authenticated to preform this action"}
		rw.WriteHeader(http.StatusForbidden)
		_ = json.NewEncoder(rw).Encode(resp)
	}

	ctx := context.WithValue(r.Context(), "user_id", userId)

	next(rw, r.WithContext(ctx))
}
