package invitations

import (
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) GetInvitationLink(ctx *respond.Ctx) error {
	// Parse claims to get the requesting user
	user := ctx.ParseClaims()
	if user == nil {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has permission to view invitations (admin only)
	if !user.IsAdmin {
		return apiErrors.ErrForbidden().SetDetail("You don't have permission to access invitation links")
	}

	// Get invitation ID from params
	invitationIDStr := ctx.Params("id")
	if invitationIDStr == "" {
		return apiErrors.ErrBadRequest().SetDetail("Invitation ID is required")
	}

	// Convert invitation ID to int64
	invitationID, err := strconv.ParseInt(invitationIDStr, 10, 64)
	if err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid invitation ID")
	}

	// Get invitation by ID
	invitation, err := rg.gctx.Crate().Sqlite.Query().GetInvitationByID(ctx.Context(), invitationID)
	if err != nil {
		utils.LogErrorWithStack("Failed to get invitation by ID", err, "invitation_id", invitationID, "user_id", user.ID)
		return apiErrors.ErrNotFound().SetDetail("Invitation not found")
	}

	// Only allow link generation for pending invitations
	if invitation.Status.String != "pending" {
		return apiErrors.ErrBadRequest().SetDetail("Can only generate links for pending invitations")
	}

	// Get app URL from settings
	appURL := rg.getSettingWithDefault("app_url", "http://localhost:3000")
	
	// Generate the invitation link
	inviteURL := appURL + "/invite/accept/" + invitation.Token

	return ctx.JSON(map[string]string{
		"invite_url": inviteURL,
	})
}