package v1

import (
	"errors"
	"fmt"

	jwtware "github.com/gofiber/contrib/jwt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"

	"github.com/mahcks/serra/internal/global"
	"github.com/mahcks/serra/internal/integrations"
	"github.com/mahcks/serra/internal/rest/v1/middleware"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/rest/v1/routes"
	authRoutes "github.com/mahcks/serra/internal/rest/v1/routes/auth"
	"github.com/mahcks/serra/internal/rest/v1/routes/calendar"
	"github.com/mahcks/serra/internal/rest/v1/routes/discover"
	downloadclients "github.com/mahcks/serra/internal/rest/v1/routes/download_clients"
	"github.com/mahcks/serra/internal/rest/v1/routes/downloads"
	"github.com/mahcks/serra/internal/rest/v1/routes/emby"
	"github.com/mahcks/serra/internal/rest/v1/routes/mounted_drives"
	"github.com/mahcks/serra/internal/rest/v1/routes/permissions"
	"github.com/mahcks/serra/internal/rest/v1/routes/radarr"
	"github.com/mahcks/serra/internal/rest/v1/routes/requests"
	"github.com/mahcks/serra/internal/rest/v1/routes/settings"
	"github.com/mahcks/serra/internal/rest/v1/routes/setup"
	"github.com/mahcks/serra/internal/rest/v1/routes/sonarr"
	"github.com/mahcks/serra/internal/rest/v1/routes/users"
	"github.com/mahcks/serra/internal/services/auth"
	"github.com/mahcks/serra/internal/websocket"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	permissionConstants "github.com/mahcks/serra/pkg/permissions"
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

	// Test endpoint for WebSocket debugging
	router.Get("/test/websocket", ctx(indexRoute.TestWebSocket))

	authRoutes := authRoutes.NewRouteGroup(gctx)
	router.Post("/auth/login", ctx(authRoutes.AuthenticateLocal)) // Updated to support both local and media server users
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

	downloadClientsRoutes := downloadclients.NewRouteGroup(gctx)
	router.Post("/downloadclient/test", ctx(downloadClientsRoutes.TestConnection))

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
	router.Get("/test", ctx(indexRoute.TestRoute))
	router.Post("/auth/logout", ctx(authRoutes.Logout))

	// Discover routes - protected by JWT middleware
	discoverRoutes := discover.NewRouteGroup(gctx, integrations)
	router.Get("/discover/trending", ctx(discoverRoutes.GetTrending))

	// Movie routes
	router.Get("/discover/movie/popular", ctx(discoverRoutes.GetPopularMovies))
	router.Get("/discover/movie/upcoming", ctx(discoverRoutes.GetUpcomingMovies))
	router.Get("/discover/movie", ctx(discoverRoutes.GetDiscoverMovie))
	router.Get("/discover/search/movie", ctx(discoverRoutes.SearchMovie))
	router.Get("/discover/search/company", ctx(discoverRoutes.SearchCompanies))
	router.Get("/discover/movie/:movie_id/watch/providers", ctx(discoverRoutes.GetMovieWatchProviders))
	router.Get("/discover/movie/:movie_id/recommendations", ctx(discoverRoutes.GetMovieRecommendations))
	router.Get("/discover/movie/:movie_id/similar", ctx(discoverRoutes.GetMovieSimilar))
	router.Get("/discover/movie/:movie_id/release-dates", ctx(discoverRoutes.GetMovieReleaseDates))

	// TV routes
	router.Get("/discover/tv/popular", ctx(discoverRoutes.GetPopularTV))
	router.Get("/discover/tv/upcoming", ctx(discoverRoutes.GetUpcomingTV))
	router.Get("/discover/tv", ctx(discoverRoutes.GetDiscoverTV))
	router.Get("/discover/search/tv", ctx(discoverRoutes.GetTVSearch))
	router.Get("/discover/tv/:series_id/recommendations", ctx(discoverRoutes.GetTVRecommendations))
	router.Get("/discover/tv/:series_id/similar", ctx(discoverRoutes.GetTVSimilar))
	router.Get("/discover/tv/:series_id/season/:season_number", ctx(discoverRoutes.GetSeasonDetails))

	// Media details route
	router.Get("/discover/media/details/:id", ctx(discoverRoutes.GetMediaDetails))

	// Watch providers routes
	router.Get("/discover/watch/providers", ctx(discoverRoutes.GetWatchProviders))
	router.Get("/discover/watch/regions", ctx(discoverRoutes.GetWatchProviderRegions))

	// Collection routes
	router.Get("/discover/collection/:collection_id", ctx(discoverRoutes.GetCollection))

	// Person routes
	router.Get("/discover/person/:person_id", ctx(discoverRoutes.GetPerson))

	// Season availability routes
	router.Get("/discover/season-availability/:id", ctx(discoverRoutes.GetSeasonAvailability))
	router.Post("/discover/season-availability/:id/sync", ctx(discoverRoutes.SyncSeasonAvailability))

	// Media ratings routes
	router.Get("/discover/media/:tmdb_id/ratings", ctx(discoverRoutes.GetMediaRatings))

	downloadsRoutes := downloads.NewRouteGroup(gctx)
	router.Get("/downloads", ctx(downloadsRoutes.GetDownloads))

	embyRoutes := emby.NewRouteGroup(gctx, integrations)
	// Generic media server routes (supports both Emby and Jellyfin)
	router.Get("/media/latest", ctx(embyRoutes.GetLatestMedia))
	// Keep legacy route for backwards compatibility
	router.Get("/emby/latest-media", ctx(embyRoutes.GetLatestMedia))

	settingsRoutes := settings.NewRouteGroup(gctx)
	router.Get("/settings", ctx(settingsRoutes.GetSettings))

	mountedDrivesRoutes := mounted_drives.NewRouteGroup(gctx, integrations)
	router.Get("/mounted-drives", ctx(mountedDrivesRoutes.GetMountedDrives))
	router.Post("/mounted-drives", ctx(mountedDrivesRoutes.CreateMountedDrive))
	router.Get("/mounted-drives/:id", ctx(mountedDrivesRoutes.GetMountedDrive))
	router.Put("/mounted-drives/:id", ctx(mountedDrivesRoutes.UpdateMountedDrive))
	router.Delete("/mounted-drives/:id", ctx(mountedDrivesRoutes.DeleteMountedDrive))
	router.Get("/mounted-drives/system/available", ctx(mountedDrivesRoutes.GetSystemDrives))

	permissionsRoutes := permissions.NewRouteGroup(gctx)
	// Permission management routes - admin only
	router.Get("/permissions", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.AdminUsers), ctx(permissionsRoutes.GetAllPermissions))
	router.Get("/permissions/categories", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.AdminUsers), ctx(permissionsRoutes.GetPermissionsByCategory))

	// User permission routes - admin only
	router.Get("/users/:id/permissions", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.AdminUsers), ctx(permissionsRoutes.GetUserPermissions))
	router.Post("/users/:id/permissions", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.AdminUsers), ctx(permissionsRoutes.AssignUserPermission))
	router.Delete("/users/:id/permissions/:permission", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.AdminUsers), ctx(permissionsRoutes.RevokeUserPermission))
	router.Put("/users/:id/permissions", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.AdminUsers), ctx(permissionsRoutes.BulkUpdateUserPermissions))

	usersRoutes := users.NewRouteGroup(gctx)
	// User management routes - admin only
	router.Get("/users", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.AdminUsers), ctx(usersRoutes.GetAllUsers))
	router.Get("/users/:id", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.AdminUsers), ctx(usersRoutes.GetUser))
	router.Post("/users/local", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.AdminUsers), ctx(authRoutes.RegisterLocalUser))
	// Password change route - accessible to users with owner/admin.users permission or self
	router.Put("/users/:id/password", ctx(authRoutes.ChangeLocalUserPassword))
	// Avatar route - accessible to all authenticated users
	router.Get("/users/:id/avatar", ctx(usersRoutes.GetUserAvatar))

	// Request routes - users can view/create requests, admins can manage them
	requestsRoutes := requests.NewRouteGroup(gctx, integrations)
	// Create request - requires appropriate permission based on media type
	router.Post("/requests", ctx(requestsRoutes.CreateRequest))
	// Get user's own requests - all authenticated users
	router.Get("/requests/me", ctx(requestsRoutes.GetUserRequests))
	// Get all requests - admin only
	router.Get("/requests", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.RequestsView), ctx(requestsRoutes.GetAllRequests))
	// Get pending requests - admin only
	router.Get("/requests/pending", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.RequestsView), ctx(requestsRoutes.GetPendingRequests))
	// Get request statistics - admin only
	router.Get("/requests/statistics", middleware.RequirePermission(gctx.Crate().Sqlite.Query(), permissionConstants.RequestsView), ctx(requestsRoutes.GetRequestStatistics))

	// Get/Update/Delete specific request by ID
	router.Get("/requests/:id", ctx(requestsRoutes.GetRequestByID))
	router.Put("/requests/:id", ctx(requestsRoutes.UpdateRequest))
	router.Delete("/requests/:id", ctx(requestsRoutes.DeleteRequest))
}
