package respond

import (
	"errors"
	
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

// ParseClaims is a helper function for use with standard fiber.Ctx (like in middleware)
func ParseClaims(c *fiber.Ctx) (*structures.User, error) {
	user, ok := c.Locals(_ctxKey).(*jwt.Token)
	if !ok {
		return nil, errors.New("no JWT token found in context")
	}

	claims, ok := user.Claims.(*auth.JWTClaimUser)
	if !ok {
		return nil, errors.New("invalid JWT claims type")
	}

	return &structures.User{
		ID:          claims.UserID,
		Username:    claims.Username,
		AccessToken: claims.AccessToken,
		IsAdmin:     claims.IsAdmin,
	}, nil
}
