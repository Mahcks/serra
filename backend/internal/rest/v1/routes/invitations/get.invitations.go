package invitations

import (
	"encoding/json"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/pkg/structures"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) GetInvitations(ctx *respond.Ctx) error {
	// Parse claims to get the requesting user
	user := ctx.ParseClaims()
	if user == nil {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has permission to view invitations (admin only for now)
	if !user.IsAdmin {
		return apiErrors.ErrForbidden().SetDetail("You don't have permission to view invitations")
	}

	// Get all invitations
	invitations, err := rg.gctx.Crate().Sqlite.Query().GetAllInvitations(ctx.Context())
	if err != nil {
		utils.LogErrorWithStack("Failed to get invitations", err, "user_id", user.ID)
		return apiErrors.ErrInternalServerError()
	}

	// Convert to API response format
	var apiInvitations []structures.Invitation
	for _, inv := range invitations {
		// Parse permissions
		var permissions []string
		if inv.Permissions.Valid && inv.Permissions.String != "" {
			if err := json.Unmarshal([]byte(inv.Permissions.String), &permissions); err != nil {
				permissions = []string{}
			}
		}

		apiInv := structures.Invitation{
			ID:              inv.ID,
			Email:           inv.Email,
			Username:        inv.Username,
			// Don't include token in list response
			InvitedBy:       inv.InvitedBy,
			InviterUsername: inv.InviterUsername,
			Permissions:     permissions,
			CreateMediaUser: inv.CreateMediaUser.Bool,
			Status:          inv.Status.String,
			ExpiresAt:       inv.ExpiresAt.Format("2006-01-02T15:04:05Z"),
			CreatedAt:       inv.CreatedAt.Time.Format("2006-01-02T15:04:05Z"),
			UpdatedAt:       inv.UpdatedAt.Time.Format("2006-01-02T15:04:05Z"),
		}

		if inv.AcceptedAt.Valid {
			apiInv.AcceptedAt = inv.AcceptedAt.Time.Format("2006-01-02T15:04:05Z")
		}

		apiInvitations = append(apiInvitations, apiInv)
	}

	return ctx.JSON(apiInvitations)
}

func (rg *RouteGroup) GetInvitationStats(ctx *respond.Ctx) error {
	// Parse claims to get the requesting user
	user := ctx.ParseClaims()
	if user == nil {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has permission to view stats (admin only for now)
	if !user.IsAdmin {
		return apiErrors.ErrForbidden().SetDetail("You don't have permission to view invitation statistics")
	}

	// Get invitation statistics
	stats, err := rg.gctx.Crate().Sqlite.Query().GetInvitationStats(ctx.Context())
	if err != nil {
		utils.LogErrorWithStack("Failed to get invitation stats", err, "user_id", user.ID)
		return apiErrors.ErrInternalServerError()
	}

	apiStats := structures.InvitationStats{
		PendingCount:   stats.PendingCount,
		AcceptedCount:  stats.AcceptedCount,
		ExpiredCount:   stats.ExpiredCount,
		CancelledCount: stats.CancelledCount,
		TotalCount:     stats.TotalCount,
	}

	return ctx.JSON(apiStats)
}