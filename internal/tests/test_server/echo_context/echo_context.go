package echo_context

import (
	"context"

	"github.com/labstack/echo/v4"
)

type contextKey struct{}

var contextKeyInstance = contextKey{}

func withEchoContext(ctx context.Context, eCtx echo.Context) context.Context {
	return context.WithValue(ctx, contextKeyInstance, eCtx)
}

func Middleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := withEchoContext(c.Request().Context(), c)
		c.SetRequest(c.Request().WithContext(ctx))
		return next(c)
	}
}

func GetFromGoContext(ctx context.Context) echo.Context {
	val := ctx.Value(contextKeyInstance)
	if val == nil {
		return nil
	}
	return val.(echo.Context)
}
