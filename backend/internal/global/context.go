package global

import (
	"context"
	"time"

	"github.com/mahcks/serra/config"
	"github.com/mahcks/serra/internal/services"
)

type Metadata struct {
	Version   string
	Timestamp string
}

type Context interface {
	context.Context
	Metadata() Metadata
	Bootstrap() *config.Bootstrap
	Crate() *services.Crate
}

type gCtx struct {
	context.Context
	metadata  Metadata
	bootstrap *config.Bootstrap
	cfg       *config.Config
	crate     *services.Crate
}

func (g *gCtx) Metadata() Metadata {
	return g.metadata
}

func (g *gCtx) Bootstrap() *config.Bootstrap {
	return g.bootstrap
}

func (g *gCtx) Crate() *services.Crate {
	return g.crate
}

func New(
	ctx context.Context,
	bootstrap *config.Bootstrap,
	Version string,
	Timestamp string,
) Context {
	return &gCtx{
		bootstrap: bootstrap,
		Context:   ctx,
		metadata: Metadata{
			Version:   Version,
			Timestamp: Timestamp,
		},
		crate: &services.Crate{},
	}
}

func WithCancel(ctx Context) (Context, context.CancelFunc) {
	metadata := ctx.Metadata()
	bootstrap := ctx.Bootstrap()
	crate := ctx.Crate()

	c, cancel := context.WithCancel(ctx)

	return &gCtx{
		Context:   c,
		bootstrap: bootstrap,
		metadata:  metadata,
		crate:     crate,
	}, cancel
}

func WithDeadline(ctx Context, deadline time.Time) (Context, context.CancelFunc) {
	metadata := ctx.Metadata()
	bootstrap := ctx.Bootstrap()
	crate := ctx.Crate()

	c, cancel := context.WithDeadline(ctx, deadline)

	return &gCtx{
		Context:   c,
		metadata:  metadata,
		bootstrap: bootstrap,
		crate:     crate,
	}, cancel
}

func WithValue(ctx Context, key interface{}, value interface{}) Context {
	metadata := ctx.Metadata()
	bootstrap := ctx.Bootstrap()
	crate := ctx.Crate()

	return &gCtx{
		Context:   context.WithValue(ctx, key, value),
		metadata:  metadata,
		bootstrap: bootstrap,
		crate:     crate,
	}
}

func WithTimeout(ctx Context, timeout time.Duration) (Context, context.CancelFunc) {
	metadata := ctx.Metadata()
	bootstrap := ctx.Bootstrap()
	crate := ctx.Crate()

	c, cancel := context.WithTimeout(ctx, timeout)

	return &gCtx{
		Context:   c,
		metadata:  metadata,
		bootstrap: bootstrap,
		crate:     crate,
	}, cancel
}
