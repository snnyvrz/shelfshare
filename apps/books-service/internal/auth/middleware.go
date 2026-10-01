package auth

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

type Claims struct {
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// PublicReads permits catalog browsing but requires a signed session for mutations.
func PublicReads(secret string) gin.HandlerFunc {
	return middleware(secret, true)
}

// Required authenticates reads as well as mutations on private resources.
func Required(secret string) gin.HandlerFunc {
	return middleware(secret, false)
}

func middleware(secret string, public bool) gin.HandlerFunc {
	if secret == "" {
		panic("JWT_SECRET environment variable is required")
	}
	return func(c *gin.Context) {
		if public && (c.Request.Method == http.MethodGet || c.Request.Method == http.MethodHead || c.Request.Method == http.MethodOptions) {
			c.Next()
			return
		}
		parts := strings.Fields(c.GetHeader("Authorization"))
		if len(parts) != 2 || parts[0] != "Bearer" {
			unauthorized(c)
			return
		}
		claims := &Claims{}
		token, err := jwt.ParseWithClaims(parts[1], claims, func(_ *jwt.Token) (any, error) {
			return []byte(secret), nil
		}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithExpirationRequired())
		if err != nil || !token.Valid || claims.Subject == "" || claims.Email == "" {
			unauthorized(c)
			return
		}
		c.Set("user_id", claims.Subject)
		c.Set("user_email", claims.Email)
		c.Set("expires_at", claims.ExpiresAt.Time)
		c.Next()
	}
}

func unauthorized(c *gin.Context) {
	c.Header("WWW-Authenticate", "Bearer")
	c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
		"code": "UNAUTHORIZED", "message": "A valid login is required",
	})
}
