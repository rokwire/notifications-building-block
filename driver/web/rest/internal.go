package rest

import "notifications/core"

// InternalApisHandler handles the rest Admin APIs implementation
type InternalApisHandler struct {
	app *core.Application
}

// NewInternalApisHandler creates new rest Handler instance
func NewInternalApisHandler(app *core.Application) *InternalApisHandler {
	return &InternalApisHandler{app: app}
}
