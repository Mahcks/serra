package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) Me(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	// Get user from database to get avatar_url
	dbUser, err := rg.gctx.Crate().Sqlite.Query().GetUserByID(ctx.Context(), user.ID)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to get user from database")
	}

	var isAdmin bool

	// Handle different user types
	if dbUser.UserType == "local" {
		// Local users: check if they have owner permission
		hasOwnerPermission, err := rg.gctx.Crate().Sqlite.Query().CheckUserPermission(ctx.Context(), repository.CheckUserPermissionParams{
			UserID:       user.ID,
			PermissionID: "owner",
		})
		if err == nil && hasOwnerPermission {
			isAdmin = true
		}
	} else {
		// Media server users: check with media server
		req, _ := http.NewRequest("GET", rg.Config().MediaServer.URL.String()+"/Users/"+user.ID, nil)

		// Set authorization header based on media server type
		version := rg.gctx.Bootstrap().Version
		utils.SetMediaServerAuthHeader(req, string(rg.Config().MediaServer.Type), version, user.AccessToken)

		client := &http.Client{Timeout: 5 * time.Second}
		res, err := client.Do(req)
		if err != nil || res.StatusCode != 200 {
			return apiErrors.ErrInternalServerError().SetDetail("failed to fetch user info")
		}
		defer res.Body.Close()

		var mediaUser struct {
			Policy struct {
				IsAdministrator bool `json:"IsAdministrator"`
			} `json:"Policy"`
		}
		json.NewDecoder(res.Body).Decode(&mediaUser)
		isAdmin = mediaUser.Policy.IsAdministrator
	}

	// Build response with avatar
	response := fiber.Map{
		"id":       user.ID,
		"username": user.Username,
		"is_admin": isAdmin,
	}

	// Add avatar_url if available
	if dbUser.AvatarUrl.Valid && dbUser.AvatarUrl.String != "" {
		response["avatar_url"] = dbUser.AvatarUrl.String
	}

	return ctx.JSON(response)
}
