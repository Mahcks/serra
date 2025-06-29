package v1

import (
	"errors"
	"fmt"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/rest/v1/routes"
	authRoutes "github.com/mahcks/serra/internal/rest/v1/routes/auth"
	"github.com/mahcks/serra/internal/rest/v1/routes/calendar"
	"github.com/mahcks/serra/internal/rest/v1/routes/downloads"
	"github.com/mahcks/serra/internal/rest/v1/routes/radarr"
	"github.com/mahcks/serra/internal/rest/v1/routes/settings"
	"github.com/mahcks/serra/internal/rest/v1/routes/setup"
	"github.com/mahcks/serra/internal/rest/v1/routes/sonarr"
	"github.com/mahcks/serra/internal/services/auth"
	"github.com/mahcks/serra/internal/websocket"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func ctx(fn func(*respond.Ctx) error) fiber.Handler {
	return func(c *fiber.Ctx) error {
		newCtx := &respond.Ctx{Ctx: c}
		return fn(newCtx)
	}
}

func New(gctx global.Context, integrations *integrations.Integration, router fiber.Router) {
	indexRoute := routes.NewRouteGroup(gctx, integrations)
	router.Get("/", ctx(indexRoute.Index))

	setupGroup := setup.NewRouteGroup(gctx)
	router.Post("/setup", ctx(setupGroup.Initialize))
	router.Get("/setup/status", ctx(indexRoute.SetupStatus))

	authRoutes := authRoutes.NewRouteGroup(gctx)
	router.Post("/auth/login", ctx(authRoutes.Authenticate))
	router.Post("/auth/refresh", ctx(authRoutes.RefreshToken))

	// WebSocket routes - register before JWT middleware
	// WebSocket handles its own authentication via cookies
	websocket.RegisterRoutes(gctx, router)

	radarrRoutes := radarr.NewRouteGroup(gctx)
	router.Post("/radarr/test", ctx(radarrRoutes.TestRadarr))
	router.Post("/radarr/qualityprofiles", ctx(radarrRoutes.GetProfiles))
	router.Post("/radarr/rootfolders", ctx(radarrRoutes.GetRootFolders))

	sonarrRoutes := sonarr.NewRouteGroup(gctx)
	router.Post("/sonarr/test", ctx(sonarrRoutes.TestSonarr))
	router.Post("/sonarr/qualityprofiles", ctx(sonarrRoutes.GetProfiles))
	router.Post("/sonarr/rootfolders", ctx(sonarrRoutes.GetSonarrRootFolders))

	calendarRoutes := calendar.NewRouteGroup(gctx, integrations)
	router.Get("/calendar/upcoming", ctx(calendarRoutes.GetUpcomingMedia))

	// JWT middleware for protected routes
	router.Use(jwtware.New(jwtware.Config{
		ContextKey:  "_serrauser",
		TokenLookup: "cookie:serra_token",
		SigningKey:  jwtware.SigningKey{Key: []byte(gctx.Bootstrap().Credentials.JwtSecret)},
		Claims:      &auth.JWTClaimUser{},
		KeyFunc: func(t *jwt.Token) (interface{}, error) {
			if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
			}

			issuer, err := t.Claims.GetIssuer()
			if err != nil || issuer != "serra-dashboard" {
				return nil, fmt.Errorf("unexpected jwt issuer=%v", issuer)
			}

			return []byte(gctx.Bootstrap().Credentials.JwtSecret), nil
		},
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			if err != nil {
				// Use errors.Is to compare the error to specific JWT error variables
				if errors.Is(err, jwt.ErrTokenExpired) {
					return apiErrors.ErrTokenExpired().SetDetail("Your session has expired. Please refresh your token.")
				} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
					return apiErrors.ErrInvalidToken().SetDetail("Token is not valid yet.")
				} else if errors.Is(err, jwt.ErrTokenMalformed) {
					return apiErrors.ErrInvalidToken().SetDetail("Malformed token.")
				} else {
					return apiErrors.ErrInvalidToken().SetDetail("Invalid token.")
				}
			}
			// Default unauthorized response
			return apiErrors.ErrUnauthorized()
		},
	}))

	router.Get("/me", ctx(indexRoute.Me))

	downloadsRoutes := downloads.NewRouteGroup(gctx)
	router.Get("/downloads", ctx(downloadsRoutes.GetDownloads))

	settingsRoutes := settings.NewRouteGroup(gctx)
	router.Get("/settings", ctx(settingsRoutes.GetSettings))
}
