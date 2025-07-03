package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/mahcks/serra/internal/db/repository"
	"github.com/mahcks/serra/internal/rest/v1/respond"
	"github.com/mahcks/serra/internal/services/auth"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

type AuthRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type authResponse struct {
	User struct {
		ID       string `json:"Id"`
		Username string `json:"Name"`
	} `json:"User"`
	Accesstoken string `json:"AccessToken"`
}

func (rg *RouteGroup) Authenticate(ctx *respond.Ctx) error {
	var req AuthRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Failed to parse request body")
	}

	authPayload := map[string]string{
		"Username": req.Username,
		"Pw":       req.Password,
	}
	authPayloadBytes, _ := json.Marshal(authPayload)

	mediaServerURL := fmt.Sprintf("%s/Users/AuthenticateByName", rg.Config().MediaServer.URL)
	httpReq, err := http.NewRequest("POST", mediaServerURL, bytes.NewBuffer(authPayloadBytes))
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to create media server request")
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Accept", "application/json")
	authKeyHeader := utils.Ternary[string](rg.Config().MediaServer.Type == structures.ProviderEmby, "X-Emby-Authorization", "X-Jellyfin-Authorization")
	httpReq.Header.Set(authKeyHeader, `Emby Client="Serra", Device="Web", DeviceId="dash-123", Version="1.0.0"`)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to contact media server")
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return apiErrors.ErrUnauthorized().SetDetail("Media server rejected credentials: " + string(respBody))
	}

	var res authResponse
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to decode media server response")
	}

	// Store user in DB
	user, err := rg.gctx.Crate().Sqlite.Query().CreateUser(ctx.Context(), repository.CreateUserParams{
		ID:          res.User.ID,
		Username:    res.User.Username,
		AccessToken: utils.NewNullString(res.Accesstoken),
	})
	if err != nil {
		slog.Debug("failed to store user", "error", err)
		return apiErrors.ErrInternalServerError().SetDetail("failed to store user in database")
	}

	// Create JWT token
	token, _, err := rg.gctx.Crate().AuthService.CreateAccessToken(user.ID, user.Username, res.Accesstoken, false)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to create JWT token")
	}

	// Store the token in a cookie
	ctx.Cookie(rg.gctx.Crate().AuthService.Cookie(auth.CookieAuth, token, time.Hour*24*14))

	return ctx.SendStatus(fiber.StatusNoContent)
}
