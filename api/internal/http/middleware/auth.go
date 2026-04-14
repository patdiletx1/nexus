package middleware

import (
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"

	"nexus/api/internal/http/handlers"
)

type AuthConfig struct {
	JWTSecret string
}

func RequireAuth(cfg AuthConfig) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if cfg.JWTSecret == "" {
				http.Error(w, "auth misconfigured", http.StatusInternalServerError)
				return
			}

			tokenString, err := bearerToken(r.Header.Get("Authorization"))
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
				if token.Method != jwt.SigningMethodHS256 {
					return nil, errors.New("unexpected signing method")
				}

				return []byte(cfg.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				http.Error(w, "invalid token claims", http.StatusUnauthorized)
				return
			}

			userID, _ := claims["sub"].(string)
			if userID == "" {
				http.Error(w, "missing subject", http.StatusUnauthorized)
				return
			}

			role, _ := claims["role"].(string)
			companyID := companyIDFromClaims(claims)

			ctx := handlers.WithAuthContext(r.Context(), userID, companyID, role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func bearerToken(header string) (string, error) {
	if header == "" {
		return "", errors.New("missing bearer token")
	}

	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") || parts[1] == "" {
		return "", errors.New("invalid bearer token format")
	}

	return parts[1], nil
}

func companyIDFromClaims(claims jwt.MapClaims) string {
	if companyID, ok := claims["company_id"].(string); ok {
		return companyID
	}

	if userMetadata, ok := claims["user_metadata"].(map[string]any); ok {
		if companyID, ok := userMetadata["company_id"].(string); ok {
			return companyID
		}
	}

	return ""
}
