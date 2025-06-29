package auth

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services/auth"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) RefreshToken(ctx *respond.Ctx) error {
	cookie := ctx.Cookies(auth.CookieAuth)
	if cookie == "" {
		return apiErrors.ErrUnauthorized().SetDetail("missing auth cookie")
	}

	_, err := rg.gctx.Crate().AuthService.ValidateJWT(cookie)
	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			// Extract the expired token's claims to re-issue
			parsed, _, err := new(jwt.Parser).ParseUnverified(cookie, &auth.JWTClaimUser{})
			if err != nil {
				return apiErrors.ErrInvalidToken().SetDetail("failed to parse token")
			}

			expiredClaims, ok := parsed.Claims.(*auth.JWTClaimUser)
			if !ok {
				return apiErrors.ErrInvalidToken().SetDetail("invalid token claims")
			}

			// Re-issue a new token with the same claims
			newToken, expireAt, err := rg.gctx.Crate().AuthService.CreateAccessToken(
				expiredClaims.UserID,
				expiredClaims.Username,
				expiredClaims.AccessToken,
				expiredClaims.IsAdmin,
			)

			ctx.Cookie(rg.gctx.Crate().AuthService.Cookie(auth.CookieAuth, newToken, time.Hour))

			return ctx.JSON(fiber.Map{
				"message":      "Token refreshed successfully",
				"expires_at":   expireAt,
				"access_token": newToken,
			})
		}

		return apiErrors.ErrInvalidToken().SetDetail("Invalid token, please log in again.")
	}

	// Token still valid → no refresh needed
	return ctx.JSON(fiber.Map{
		"message": "Token still valid, no refresh needed.",
	})
}
