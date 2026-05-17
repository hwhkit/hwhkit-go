// Package core defines the bootstrap pipeline, runtime traits, and built application type for hwhkit-go.
package core

import (
	"context"
	"net/http"

	"github.com/hwhkit/hwhkit-go/config"
	"github.com/hwhkit/hwhkit-go/core/appctx"
)

type Application interface {
	BuildRouter(ctx context.Context, app *appctx.Context, cfg *config.AppConfig) (http.Handler, error)
}
