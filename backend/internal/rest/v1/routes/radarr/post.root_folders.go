package radarr

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
)

type getRootFoldersRequest struct {
	BaseURL string `json:"base_url"`
	APIKey  string `json:"api_key"`
}

type rootFolder struct {
	Path            string           `json:"path"`
	Accessible      bool             `json:"accessible"`
	FreeSpace       int64            `json:"free_space"`
	UnmappedFolders []unmappedFolder `json:"unmapped_folders,omitempty"`
}

type unmappedFolder struct {
	Name         string `json:"name"`
	Path         string `json:"path"`
	RelativePath string `json:"relative_path"`
}

func (rg *RouteGroup) GetRootFolders(ctx *respond.Ctx) error {
	var req getRootFoldersRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("failed to parse request body")
	}

	client := &http.Client{Timeout: 30 * time.Second}
	url := fmt.Sprintf("%s/api/v3/rootFolder", req.BaseURL)
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

	var folders []rootFolder
	if err := json.NewDecoder(resp.Body).Decode(&folders); err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("failed to decode Radarr response")
	}

	var result []structures.RadarrRootFolder
	for _, folder := range folders {
		rf := structures.RadarrRootFolder{
			Path:       folder.Path,
			Accessible: folder.Accessible,
			FreeSpace:  folder.FreeSpace,
		}

		if len(folder.UnmappedFolders) > 0 {
			umappedFolders := make([]structures.RadarrUnmappedFolder, len(folder.UnmappedFolders))
			for i, unmapped := range folder.UnmappedFolders {
				umappedFolders[i] = structures.RadarrUnmappedFolder{
					Name:         unmapped.Name,
					Path:         unmapped.Path,
					RelativePath: unmapped.RelativePath,
				}
			}
			rf.UnmappedFolders = umappedFolders
		}

		result = append(result, rf)
	}

	return ctx.JSON(result)
}
