package hwhkit

import (
	"github.com/hwhkit/hwhkit-go/core"
	"github.com/hwhkit/hwhkit-go/core/appctx"
)

// Type aliases re-exported for ergonomic use of `hwhkit.Application`, `hwhkit.AppContext`, etc.
type (
	Application         = core.Application
	IntegrationProvider = core.IntegrationProvider
	BuiltApplication    = core.BuiltApplication
	BaseProvider        = core.BaseProvider
	AppContext          = appctx.Context
)
