package routes

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) Me(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	req, _ := http.NewRequest("GET", rg.Config().MediaServer.URL.String()+"/Users/"+user.ID, nil)
	mediaServerHeader := utils.Ternary[string](rg.Config().MediaServer.Type == structures.ProviderEmby, "X-Emby-Token", "X-Jellyfin-Token")
	req.Header.Set(mediaServerHeader, user.AccessToken)

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

	return ctx.JSON(fiber.Map{
		"id":       user.ID,
		"username": user.Username,
		"is_admin": mediaUser.Policy.IsAdministrator,
	})
}
