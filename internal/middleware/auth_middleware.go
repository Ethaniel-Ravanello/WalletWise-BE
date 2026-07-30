package middleware

import (
	"context"
	"net/http"
	"strings"

	"walletwise/pkg/jwt"
)

type contextKey string

const UserIdKey contextKey = "user_id"

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}

		splitBearer := strings.Split(authHeader, " ")
		if len(splitBearer) != 2 {
			http.Error(w, "Invalid Or Unauthorize Token", http.StatusUnauthorized)
		}

		tokenString := splitBearer[1]

		claims, err := jwt.ValidateToken(tokenString)
		if err != nil {
			http.Error(w, http.StatusText(http.StatusUnauthorized), http.StatusUnauthorized)
		}

		ctxId := context.WithValue(r.Context(), UserIdKey, claims.UserId)
		rWithCtxId := r.WithContext(ctxId)

		next.ServeHTTP(w, rWithCtxId)
	})
}
