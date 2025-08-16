package auth

import (
	"errors"
	"log/slog"
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

	// Try to validate the current token
	claims, err := rg.gctx.Crate().AuthService.ValidateJWT(cookie)
	if err != nil {
		// If token is expired, try to refresh it
		if errors.Is(err, jwt.ErrTokenExpired) {
			return rg.refreshExpiredToken(ctx, cookie)
		}
		return apiErrors.ErrInvalidToken().SetDetail("invalid token")
	}

	// If token is still valid but expires soon (within 30 minutes), refresh it proactively
	if claims.ExpiresAt != nil && time.Until(claims.ExpiresAt.Time) < 30*time.Minute {
		slog.Debug("Token expires soon, refreshing proactively", "user_id", claims.UserID, "expires_at", claims.ExpiresAt.Time)
		return rg.refreshValidToken(ctx, claims)
	}

	// Token still valid and not expiring soon
	return ctx.JSON(fiber.Map{
		"message": "Token still valid, no refresh needed.",
	})
}

// refreshExpiredToken handles refreshing an expired token
func (rg *RouteGroup) refreshExpiredToken(ctx *respond.Ctx, cookie string) error {
	// Extract claims from expired token
	parsed, _, err := new(jwt.Parser).ParseUnverified(cookie, &auth.JWTClaimUser{})
	if err != nil {
		return apiErrors.ErrInvalidToken().SetDetail("failed to parse expired token")
	}

	expiredClaims, ok := parsed.Claims.(*auth.JWTClaimUser)
	if !ok {
		return apiErrors.ErrInvalidToken().SetDetail("invalid token claims")
	}

	// Validate that the user still exists and is active
	if err := rg.validateUserForRefresh(ctx, expiredClaims); err != nil {
		return err
	}

	// Create new token
	return rg.issueNewToken(ctx, expiredClaims)
}

// refreshValidToken handles refreshing a valid but soon-to-expire token
func (rg *RouteGroup) refreshValidToken(ctx *respond.Ctx, claims *auth.JWTClaimUser) error {
	// Validate that the user still exists and is active
	if err := rg.validateUserForRefresh(ctx, claims); err != nil {
		return err
	}

	// Create new token
	return rg.issueNewToken(ctx, claims)
}

// validateUserForRefresh ensures the user is still valid for token refresh
func (rg *RouteGroup) validateUserForRefresh(ctx *respond.Ctx, claims *auth.JWTClaimUser) error {
	// Check if user still exists in database
	dbUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), claims.UserID)
	if err != nil {
		slog.Warn("User not found during token refresh", "user_id", claims.UserID)
		// Clear the invalid cookie to stop refresh attempts
		rg.clearAuthCookie(ctx)
		return apiErrors.ErrUnauthorized().SetDetail("user no longer exists")
	}

	// For media server users, we could validate the access token is still valid
	// For now, we'll just check that the user exists
	// TODO: Add media server token validation if needed

	slog.Debug("User validated for token refresh", "user_id", dbUser.ID, "username", dbUser.Username, "user_type", dbUser.UserType)
	return nil
}

// issueNewToken creates and sets a new JWT token
func (rg *RouteGroup) issueNewToken(ctx *respond.Ctx, claims *auth.JWTClaimUser) error {
	// Create new token with same claims
	newToken, expireAt, err := rg.gctx.Crate().AuthService.CreateAccessToken(
		claims.UserID,
		claims.Username,
		claims.AccessToken,
		claims.IsAdmin,
	)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to create new token")
	}

	// Set cookie with consistent duration (24 hours like login)
	ctx.Cookie(rg.gctx.Crate().AuthService.Cookie(auth.CookieAuth, newToken, time.Hour*24*14))

	slog.Info("Token refreshed successfully", "user_id", claims.UserID, "username", claims.Username)

	return ctx.JSON(fiber.Map{
		"message":      "Token refreshed successfully",
		"expires_at":   expireAt,
		"access_token": newToken,
	})
}

// clearAuthCookie clears the authentication cookie by setting it to expire immediately
func (rg *RouteGroup) clearAuthCookie(ctx *respond.Ctx) {
	// Use the same cookie creation method as logout but with immediate expiration
	ctx.Cookie(rg.gctx.Crate().AuthService.Cookie(auth.CookieAuth, "", time.Second*-1))
	slog.Info("Cleared invalid authentication cookie")
}
