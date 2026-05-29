package handlers

import (
	"context"
	paste "github.com/infinage/pastebin/pkg"
)

type Application struct {
	st  *paste.Store
	ctx context.Context
}

func NewApplication(ctx context.Context) *Application {
	return &Application{st: paste.NewEmptyStore(), ctx: ctx}
}
