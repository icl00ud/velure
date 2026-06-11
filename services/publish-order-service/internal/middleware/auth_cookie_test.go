package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/golang-jwt/jwt/v5"
)

func signTestToken(t *testing.T, secret, subject string) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.RegisteredClaims{Subject: subject})
	s, err := tok.SignedString([]byte(secret))
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestAuth_AcceptsAccessTokenCookie(t *testing.T) {
	const secret = "test-secret"
	token := signTestToken(t, secret, "user-7")

	var gotUserID string
	handler := Auth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with cookie auth, got %d", w.Code)
	}
	if gotUserID != "user-7" {
		t.Fatalf("expected user-7 from cookie token, got %q", gotUserID)
	}
}

func TestSSEAuth_AcceptsAccessTokenCookie(t *testing.T) {
	const secret = "test-secret"
	token := signTestToken(t, secret, "user-8")

	var gotUserID string
	handler := SSEAuth(secret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID = GetUserID(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/events", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 with cookie auth, got %d", w.Code)
	}
	if gotUserID != "user-8" {
		t.Fatalf("expected user-8 from cookie token, got %q", gotUserID)
	}
}
