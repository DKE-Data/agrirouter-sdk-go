package main

import (
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/tests/test_server"
	"github.com/DKE-Data/agrirouter-sdk-go/internal/tests/test_server/echo_context"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/tmaxmax/go-sse"
)

func main() {
	server := test_server.NewServer()

	sseServer := &sse.Server{}

	go func() {
		for event := range server.GetTestEvents() {
			eventType, err := sse.NewType(event.EventType)
			if err != nil {
				slog.Error("Error creating SSE type for test event", "error", err)
				continue
			}
			m := &sse.Message{
				Type: eventType,
			}
			m.AppendData(event.Data)
			err = sseServer.Publish(m)
			if err != nil {
				slog.Error("Error publishing SSE message with test event", "error", err)
				continue
			}
			slog.Info("Server sent test event", "eventType", event.EventType, "data", event.Data)
		}
	}()

	strict := test_server.NewStrictHandler(server, nil)

	e := echo.New()
	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))

	e.Use(echo_context.Middleware)
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus:   true,
		LogURI:      true,
		LogError:    true,
		HandleError: true,
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			if v.Error == nil {
				logger.LogAttrs(context.Background(), slog.LevelInfo, "REQUEST",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("method", v.Method),
				)
			} else {
				logger.LogAttrs(context.Background(), slog.LevelError, "REQUEST_ERROR",
					slog.String("uri", v.URI),
					slog.Int("status", v.Status),
					slog.String("err", v.Error.Error()),
					slog.String("method", v.Method),
				)
			}
			return nil
		},
	}))

	e.GET("/_testEvents", func(c echo.Context) error {
		slog.Info("Client connected to /_testEvents, starting SSE server")
		go func() {
			time.Sleep(100 * time.Millisecond)
			eventType, err := sse.NewType("ready")
			if err != nil {
				slog.Error("Error creating ready event type", "error", err)
				return
			}
			m := &sse.Message{Type: eventType}
			m.AppendData(`"ready"`)
			err = sseServer.Publish(m)
			if err != nil {
				slog.Error("Error publishing ready event", "error", err)
			} else {
				slog.Info("Published ready event")
			}
		}()
		sseServer.ServeHTTP(c.Response(), c.Request())
		return nil
	})
	test_server.RegisterHandlers(e, strict)

	err := e.Start(":8080")
	if err != nil {
		slog.Error("Error starting server", "error", err)
		os.Exit(1)
	}
}
