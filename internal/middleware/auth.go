package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/Gemini8532/gitwapp/internal/config"
	"github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return config.JWTSecret, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok {
			ctx := context.WithValue(r.Context(), "user_id", claims["sub"])
			next.ServeHTTP(w, r.WithContext(ctx))
		} else {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
		}
	})
}

// LocalhostOnlyMiddleware ensures that requests only come from localhost
func LocalhostOnlyMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// This is a naive check. For production, consider checking X-Forwarded-For if behind proxy,
		// but since this is for internal CLI usage on the same machine, RemoteAddr is usually sufficient.
		// IPv4: 127.0.0.1:port, IPv6: [::1]:port
		remoteAddr := r.RemoteAddr
		if !strings.HasPrefix(remoteAddr, "127.0.0.1:") && !strings.HasPrefix(remoteAddr, "[::1]:") {
			http.Error(w, "Forbidden: Localhost only", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}
