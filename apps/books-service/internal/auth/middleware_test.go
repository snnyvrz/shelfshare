package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func TestPublicReads(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "test-shared-secret"
	sign := func(method jwt.SigningMethod, key string, exp *jwt.NumericDate, sub, email string) string {
		token, err := jwt.NewWithClaims(method, Claims{Email: email, RegisteredClaims: jwt.RegisteredClaims{Subject: sub, ExpiresAt: exp}}).SignedString([]byte(key))
		if err != nil {
			t.Fatal(err)
		}
		return token
	}
	future := jwt.NewNumericDate(time.Now().Add(time.Hour))
	valid := sign(jwt.SigningMethodHS256, secret, future, "user-id", "reader@example.com")
	tests := []struct {
		name, method, header string
		status               int
	}{
		{"public books", "GET", "", 200},
		{"missing token", "POST", "", 401},
		{"malformed token", "PATCH", "Bearer garbage", 401},
		{"wrong scheme", "DELETE", "Basic " + valid, 401},
		{"valid create", "POST", "Bearer " + valid, 200},
		{"valid update", "PATCH", "Bearer " + valid, 200},
		{"valid delete", "DELETE", "Bearer " + valid, 200},
		{"wrong secret", "POST", "Bearer " + sign(jwt.SigningMethodHS256, "other", future, "user-id", "reader@example.com"), 401},
		{"wrong algorithm", "POST", "Bearer " + sign(jwt.SigningMethodHS384, secret, future, "user-id", "reader@example.com"), 401},
		{"expired", "POST", "Bearer " + sign(jwt.SigningMethodHS256, secret, jwt.NewNumericDate(time.Now().Add(-time.Hour)), "user-id", "reader@example.com"), 401},
		{"no expiration", "POST", "Bearer " + sign(jwt.SigningMethodHS256, secret, nil, "user-id", "reader@example.com"), 401},
		{"no subject", "POST", "Bearer " + sign(jwt.SigningMethodHS256, secret, future, "", "reader@example.com"), 401},
		{"no email", "POST", "Bearer " + sign(jwt.SigningMethodHS256, secret, future, "user-id", ""), 401},
	}
	for _, path := range []string{"/api/books", "/api/authors"} {
		for _, tt := range tests {
			t.Run(path+"/"+tt.name, func(t *testing.T) {
				router := gin.New()
				api := router.Group("/api", PublicReads(secret))
				api.Any(path[len("/api"):], func(c *gin.Context) {
					if c.Request.Method != "GET" && c.GetString("user_id") != "user-id" {
						t.Error("missing authenticated user")
					}
					c.Status(http.StatusOK)
				})
				request := httptest.NewRequest(tt.method, path, nil)
				request.Header.Set("Authorization", tt.header)
				response := httptest.NewRecorder()
				router.ServeHTTP(response, request)
				if response.Code != tt.status {
					t.Fatalf("got %d, want %d: %s", response.Code, tt.status, response.Body.String())
				}
			})
		}
	}
}
