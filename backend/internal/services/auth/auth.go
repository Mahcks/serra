package auth

import (
	"time"

	"github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
)

type Authmen interface {
	SignJWT(secret string, claim jwt.Claims) (string, error)
	CreateAccessToken(id, username, accessToken string, isAdmin bool) (string, time.Time, error)
	ValidateJWT(tokenStr string) (*JWTClaimUser, error)

	Cookie(key, token string, duration time.Duration) *fiber.Cookie
}

type authmen struct {
	// Secret key used for signing.
	JWTSecret string
	// Domain for the cookie.
	Domain string
	// If cookie should be secure or not
	Secure bool
}

const (
	CookieAuth = "serra_token"
)

func New(jwtSecret, domain string, secure bool) Authmen {
	a := &authmen{
		JWTSecret: jwtSecret,
		Domain:    domain,
		Secure:    secure,
	}

	return a
}

// CreateAccessToken creates a new access token which represents a user.
func (a *authmen) CreateAccessToken(id, username, accessToken string, isAdmin bool) (string, time.Time, error) {
	expireAt := time.Now().Add(time.Hour * 2) // Make the actual JWT expire in 2 hours

	token, err := a.SignJWT(a.JWTSecret, &JWTClaimUser{
		UserID:      id,
		Username:    username,
		AccessToken: accessToken,
		IsAdmin:     isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "serra-dashboard",
			ExpiresAt: &jwt.NumericDate{Time: expireAt},
			NotBefore: &jwt.NumericDate{Time: time.Now()},
			IssuedAt:  &jwt.NumericDate{Time: time.Now()},
		},
	})
	if err != nil {
		return "", time.Time{}, err
	}

	return token, expireAt, nil
}

func (a *authmen) Cookie(key, token string, duration time.Duration) *fiber.Cookie {
	cookie := &fiber.Cookie{}
	cookie.Name = key
	cookie.Value = token
	cookie.Expires = time.Now().Add(duration)
	cookie.HTTPOnly = true
	cookie.Domain = a.Domain
	cookie.Path = "/"
	cookie.SameSite = fiber.CookieSameSiteLaxMode // Use Lax for better compatibility
	cookie.Secure = a.Secure // Use the configured secure setting

	return cookie
}
