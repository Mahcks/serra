package respond

import (
	"github.com/gofiber/fiber/v2"
	jwt "github.com/golang-jwt/jwt/v5"
	"github.com/mahcks/serra/internal/services/auth"
	"github.com/mahcks/serra/pkg/structures"
)

const (
	_ctxKey string = "_serrauser"
)

type Ctx struct {
	*fiber.Ctx
}

func (ctx *Ctx) ParseClaims() *structures.User {
	user, ok := ctx.Locals(_ctxKey).(*jwt.Token)
	if !ok {
		return nil
	}

	claims, ok := user.Claims.(*auth.JWTClaimUser)
	if !ok {
		return nil
	}

	return &structures.User{
		ID:          claims.UserID,
		Username:    claims.Username,
		AccessToken: claims.AccessToken,
		IsAdmin:     claims.IsAdmin,
	}
}
