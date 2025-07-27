package users

import (
	"database/sql"
	"log/slog"

	"github.com/mahcks/serra/internal/rest/v1/respond"
	apiErrors "github.com/mahcks/serra/pkg/api_errors"
	"github.com/mahcks/serra/internal/db/repository"
)

type UpdateUserRequest struct {
	Username string  `json:"username"`
	Email    *string `json:"email"`
}

// UpdateUser updates basic user information (username, email)
func (rg *RouteGroup) UpdateUser(ctx *respond.Ctx) error {
	user := ctx.ParseClaims()
	if user == nil || user.ID == "" {
		return apiErrors.ErrUnauthorized()
	}

	userID := ctx.Params("id")
	if userID == "" {
		return apiErrors.ErrBadRequest().SetDetail("user ID is required")
	}

	var req UpdateUserRequest
	if err := ctx.BodyParser(&req); err != nil {
		return apiErrors.ErrBadRequest().SetDetail("invalid request body")
	}

	// Basic validation
	if req.Username == "" {
		return apiErrors.ErrBadRequest().SetDetail("username is required")
	}

	// Check if user exists
	exists, err := rg.gctx.Crate().Sqlite.Query().UserExists(ctx.Context(), userID)
	if err != nil {
		slog.Error("Failed to check if user exists", "error", err, "user_id", userID)
		return apiErrors.ErrInternalServerError().SetDetail("failed to check user existence")
	}

	if exists == 0 {
		return apiErrors.ErrNotFound().SetDetail("user not found")
	}

	// Prepare update parameters
	var email sql.NullString
	if req.Email != nil && *req.Email != "" {
		email = sql.NullString{String: *req.Email, Valid: true}
	}

	updateParams := repository.UpdateUserParams{
		Username: req.Username,
		Email:    email,
		ID:       userID,
	}

	// Update the user
	err = rg.gctx.Crate().Sqlite.Query().UpdateUser(ctx.Context(), updateParams)
	if err != nil {
		slog.Error("Failed to update user", "error", err, "user_id", userID)
		return apiErrors.ErrInternalServerError().SetDetail("failed to update user")
	}

	slog.Info("User updated successfully", "user_id", userID, "updated_by", user.ID)

	return ctx.JSON(map[string]interface{}{
		"message": "User updated successfully",
		"user_id": userID,
	})
}