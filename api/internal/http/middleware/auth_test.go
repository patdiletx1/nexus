package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

func TestRequireAuthMissingToken(t *testing.T) {
	mw := RequireAuth(AuthConfig{JWTSecret: "test-secret"})
	handler := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Fatal("handler should not have been called")
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/protected", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}
}

func TestRequireAuthValidToken(t *testing.T) {
	secret := "test-secret"
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":        "user-123",
		"company_id": "company-999",
		"role":       "member",
		"exp":        time.Now().Add(10 * time.Minute).Unix(),
	})
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		t.Fatalf("failed creating token: %v", err)
	}

	called := false
	mw := RequireAuth(AuthConfig{JWTSecret: secret})
	handler := mw(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))

	req := httptest.NewRequest(http.MethodGet, "/v1/protected", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if !called {
		t.Fatal("expected wrapped handler to be called")
	}
}
