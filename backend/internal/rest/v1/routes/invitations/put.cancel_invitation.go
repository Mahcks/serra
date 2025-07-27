package invitations

import (
	"log/slog"
	"strconv"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/utils"
)

func (rg *RouteGroup) CancelInvitation(ctx *respond.Ctx) error {
	// Parse claims to get the requesting user
	user := ctx.ParseClaims()
	if user == nil {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has permission to cancel invitations (admin only for now)
	if !user.IsAdmin {
		return apiErrors.ErrForbidden().SetDetail("You don't have permission to cancel invitations")
	}

	// Get invitation ID from params
	idParam := ctx.Params("id")
	invitationID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid invitation ID")
	}

	// Cancel the invitation
	invitation, err := rg.gctx.Crate().Sqlite.Query().CancelInvitation(ctx.Context(), invitationID)
	if err != nil {
		utils.LogErrorWithStack("Failed to cancel invitation", err,
			"invitation_id", invitationID,
			"user_id", user.ID)
		return apiErrors.ErrInternalServerError()
	}

	// Audit log for invitation cancellation
	slog.Info("AUDIT: Invitation cancelled",
		"action", "invitation.cancel",
		"invitation_id", invitationID,
		"email", invitation.Email,
		"cancelled_by", user.ID,
		"cancelled_by_username", user.Username,
		"client_ip", ctx.IP(),
		"user_agent", ctx.Get("User-Agent"))

	return ctx.JSON(map[string]interface{}{
		"message": "Invitation cancelled successfully",
		"invitation": map[string]interface{}{
			"id":     invitation.ID,
			"email":  invitation.Email,
			"status": invitation.Status,
		},
	})
}

func (rg *RouteGroup) DeleteInvitation(ctx *respond.Ctx) error {
	// Parse claims to get the requesting user
	user := ctx.ParseClaims()
	if user == nil {
		return apiErrors.ErrUnauthorized()
	}

	// Check if user has permission to delete invitations (admin only for now)
	if !user.IsAdmin {
		return apiErrors.ErrForbidden().SetDetail("You don't have permission to delete invitations")
	}

	// Get invitation ID from params
	idParam := ctx.Params("id")
	invitationID, err := strconv.ParseInt(idParam, 10, 64)
	if err != nil {
		return apiErrors.ErrBadRequest().SetDetail("Invalid invitation ID")
	}

	// Get invitation details before deletion for audit logging
	invitation, err := rg.gctx.Crate().Sqlite.Query().GetInvitationByID(ctx.Context(), invitationID)
	if err != nil {
		utils.LogErrorWithStack("Failed to get invitation for deletion", err,
			"invitation_id", invitationID,
			"user_id", user.ID)
		// Continue with deletion even if we can't get details
	}

	// Delete the invitation
	err = rg.gctx.Crate().Sqlite.Query().DeleteInvitation(ctx.Context(), invitationID)
	if err != nil {
		utils.LogErrorWithStack("Failed to delete invitation", err,
			"invitation_id", invitationID,
			"user_id", user.ID)
		return apiErrors.ErrInternalServerError()
	}

	// Audit log for invitation deletion
	email := "unknown"
	if invitation.Email != "" {
		email = invitation.Email
	}
	slog.Info("AUDIT: Invitation deleted",
		"action", "invitation.delete",
		"invitation_id", invitationID,
		"email", email,
		"deleted_by", user.ID,
		"deleted_by_username", user.Username,
		"client_ip", ctx.IP(),
		"user_agent", ctx.Get("User-Agent"))

	return ctx.JSON(map[string]interface{}{
		"message": "Invitation deleted successfully",
	})
}