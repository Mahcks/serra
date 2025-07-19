package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (rg *RouteGroup) GetPerson(ctx *respond.Ctx) error {
	personID := ctx.Params("person_id")

	tmdbResp, err := rg.integrations.TMDB.GetPerson(personID)
	if err != nil {
		return apiErrors.ErrInternalServerError()
	}

	return ctx.JSON(tmdbResp)
}