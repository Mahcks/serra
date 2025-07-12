package discover

import (
	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
)

func (r *RouteGroup) SearchCompanies(ctx *respond.Ctx) error {
	query := ctx.Query("query")
	if query == "" {
		return apiErrors.ErrBadRequest().SetDetail("Query parameter is required")
	}
	
	page := ctx.Query("page", "1")
	
	// Search for companies using TMDB
	companies, err := r.integrations.TMDB.SearchCompanies(query, page)
	if err != nil {
		return apiErrors.ErrInternalServerError().SetDetail("Failed to search companies: " + err.Error())
	}
	
	return ctx.JSON(companies)
}