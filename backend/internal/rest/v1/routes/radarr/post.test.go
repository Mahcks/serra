package radarr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

type testRadarrRequest struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

func (rg *RouteGroup) TestRadarr(ctx *respond.Ctx) error {
	var req testRadarrRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("failed to parse request body")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("%s/api/v3/system/status", req.BaseURL)
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to create request")
	}
	httpReq.Header.Set("X-Api-Key", req.APIKey)

	resp, err := client.Do(httpReq)
	if err != nil {
		return apiErrors.ErrBadGateway().SetDetail("failed to contact Radarr server")
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return apiErrors.ErrInternalServerError().SetDetail("Radarr server returned an error: " + resp.Status)
	}

	var statusResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&statusResp); err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to decode Radarr response")
	}

	return ctx.JSON(map[string]interface{}{
		"success": true,
		"status":  statusResp,
	})
}
