package auth

import (
	"fmt"
	"log/slog"
	"time"

	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/mahcks/serra/utils"
)

// SignJWT signs a JWT token and returns the token string
func (a *authmen) SignJWT(secret string, claim jwt.Claims) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claim)

	// Sign it
	tokenStr, err := token.SignedString(utils.S2B(secret))

	return tokenStr, err
}

// ValidateJWT validates a JWT token and returns a JWTClaimUser if the token is valid
func (a *authmen) ValidateJWT(tokenStr string) (*JWTClaimUser, error) {
	claims := &JWTClaimUser{}

	token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return []byte(a.JWTSecret), nil
	})
	if err != nil {
		// Add debug logging for JWT parsing errors
		slog.Debug("JWT parsing failed", "error", err, "current_time", time.Now().Unix())
		if claims.ExpiresAt != nil {
			slog.Debug("Token expiration info", "expires_at", claims.ExpiresAt.Time, "expires_unix", claims.ExpiresAt.Unix())
		}
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	if !token.Valid {
		slog.Debug("Token marked as invalid", "valid", token.Valid)
		return nil, fmt.Errorf("token is invalid")
	}

	return claims, nil
}

type JWTClaimUser struct {
	UserID      string `json:"id"`
	Username    string `json:"username"`
	AccessToken string `json:"access_token"`
	IsAdmin     bool   `json:"is_admin"`
	ServerType  string `json:"server_type"`

	jwt.RegisteredClaims
}

type JWTClaimOAuth2CSRF struct {
	State     string    `json:"s"`
	CreatedAt time.Time `json:"at"`
	Bind      string    `json:"bind"`

	jwt.RegisteredClaims
}
