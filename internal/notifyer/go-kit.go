package alertmanager

import (
	"context"
	"log/slog"
)

type GoKitAdapter struct {
	Ctx      context.Context
	Logger   *slog.Logger
	LogLevel slog.Level
	Message  string
}

func (a *GoKitAdapter) Log(keyvals ...interface{}) error {
	a.Logger.Log(a.Ctx, a.LogLevel, a.Message, keyvals...)
	return nil
}
