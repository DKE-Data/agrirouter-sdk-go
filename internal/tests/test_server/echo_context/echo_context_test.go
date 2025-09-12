package echo_context

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestGetFromGoContextNil(t *testing.T) {
	assert.Nil(t, GetFromGoContext(context.Background()))
}

func TestMiddlewareSetsContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", http.NoBody)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	called := false
	next := func(ec echo.Context) error {
		called = true
		got := GetFromGoContext(ec.Request().Context())
		assert.NotNil(t, got)
		assert.Equal(t, ec, got)
		return nil
	}

	handler := Middleware(next)
	err := handler(c)
	assert.NoError(t, err)
	assert.True(t, called)
}
