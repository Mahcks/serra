package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetCollection(ctx *respond.Ctx) error {
	collectionID := ctx.Params("collection_id")

	tmdbResp, err := rg.integrations.TMDB.GetCollection(collectionID)
	if err != nil {
		return apiErrors.ErrInternalServerError()
	}

	return ctx.JSON(tmdbResp)
}