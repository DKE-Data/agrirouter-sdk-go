package test_server

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/url"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/tests/agriroutertestcontainer"
	"github.com/DKE-Data/agrirouter-sdk-go/internal/tests/test_server/echo_context"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/tmaxmax/go-sse"
)

var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	events chan struct {
		Data      string
		EventType string
	}
	sentMessagesTestEvents chan *SendMessagesTestEventData
}

func (s *Server) GetMessagePayload(
	ctx context.Context,
	request GetMessagePayloadRequestObject,
) (GetMessagePayloadResponseObject, error) {
	// even though this is part of the spec, test server does not need to implement this
	// because it uses _testPayloads routes to serve payloads and correctly behaving clients
	// are expected to use uris provided in events, rather than calling this resource
	panic("should not be implemented")
}

type SendMessagesTestEventData struct {
	EndpointID   uuid.UUID `json:"endpointId"`
	Payload      string    `json:"payload"` // base64-encoded payload
	MessageType  string    `json:"messageType"`
	AppMessageId string    `json:"appMessageId"`
	Filename     *string   `json:"filename,omitempty"`
}

func (s *Server) ReceiveEvents(ctx context.Context, request ReceiveEventsRequestObject) (ReceiveEventsResponseObject, error) {
	sseServer := &sse.Server{}
	eCtx := echo_context.GetFromGoContext(ctx)
	receivedMessageType, err := sse.NewType(string(MESSAGERECEIVED))
	if err != nil {
		return nil, err
	}
	fileReceivedType, err := sse.NewType(string(FILERECEIVED))
	if err != nil {
		return nil, err
	}
	go func() {
		for {
			select {
			case <-ctx.Done():
				slog.Info("Context done, stopping receiving events")
				return
			case messageSentTestEvent := <-s.sentMessagesTestEvents:
				messageId := uuid.New()
				payloadPath := fmt.Sprintf("/_testPayloads/%s/2025-09-18", messageId.String())

				eCtx.Echo().GET(payloadPath, func(c echo.Context) error {
					payloadBytes, err := base64.StdEncoding.DecodeString(messageSentTestEvent.Payload)
					if err != nil {
						slog.Error("Error decoding base64 payload", "error", err)
						return c.NoContent(500)
					}
					return c.Blob(200, "application/octet-stream", payloadBytes)
				})

				payloadUri := url.URL{
					Scheme: eCtx.Scheme(),
					Host:   eCtx.Request().Host,
					Path:   payloadPath,
				}

				payloadUriStr := payloadUri.String()

				if isFileMessageType(messageSentTestEvent.MessageType) {
					sseMessage := &sse.Message{
						Type: fileReceivedType,
					}
					// compute size from base64 payload
					decoded, decErr := base64.StdEncoding.DecodeString(messageSentTestEvent.Payload)
					var size int64
					if decErr == nil {
						size = int64(len(decoded))
					}
					eventData := FileReceivedEventData{
						EventType:           string(FILERECEIVED),
						MessageType:         messageSentTestEvent.MessageType,
						ReceivingEndpointId: messageSentTestEvent.EndpointID,
						PayloadUri:          &payloadUriStr,
						Filename:            messageSentTestEvent.Filename,
						Size:                size,
					}
					marshalledEventData, err := json.Marshal(eventData)
					if err != nil {
						slog.Error("Error marshaling FileReceivedEventData", "error", err)
						continue
					}
					sseMessage.AppendData(string(marshalledEventData))
					publishErr := sseServer.Publish(sseMessage)
					if publishErr != nil {
						slog.Error("Error publishing SSE message", "error", publishErr)
					} else {
						slog.Info("Server sent FileReceived event", "data", string(marshalledEventData))
					}
				} else {
					sseMessage := &sse.Message{
						Type: receivedMessageType,
					}
					eventData := MessageReceivedEventData{
						AppMessageId:        messageSentTestEvent.AppMessageId,
						EventType:           string(MESSAGERECEIVED),
						PayloadUri:          &payloadUriStr,
						MessageType:         messageSentTestEvent.MessageType,
						Id:                  messageId,
						ReceivingEndpointId: messageSentTestEvent.EndpointID,
					}
					marshalledEventData, err := json.Marshal(eventData)
					if err != nil {
						slog.Error("Error marshaling MessageReceivedEventData", "error", err)
						continue
					}
					sseMessage.AppendData(string(marshalledEventData))
					publishErr := sseServer.Publish(sseMessage)
					if publishErr != nil {
						slog.Error("Error publishing SSE message", "error", publishErr)
					} else {
						slog.Info("Server sent MessageReceived event", "data", string(marshalledEventData))
					}
				}
			}
		}
	}()

	slog.Info("Client connected to receive events")

	sseServer.ServeHTTP(eCtx.Response(), eCtx.Request())
	return nil, nil
}

func (s *Server) SendMessages(ctx context.Context, request SendMessagesRequestObject) (SendMessagesResponseObject, error) {
	bodyBytes, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}
	bodyBase64 := base64.StdEncoding.EncodeToString(bodyBytes)
	var data SendMessagesTestEventData = SendMessagesTestEventData{
		EndpointID:   request.Params.XAgrirouterEndpointId,
		Payload:      bodyBase64,
		MessageType:  request.Params.XAgrirouterMessageType,
		AppMessageId: request.Params.XAgrirouterContextId + "-0",
		Filename:     request.Params.XAgrirouterFilename,
	}
	s.sentMessagesTestEvents <- &data

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	s.events <- struct {
		Data      string
		EventType string
	}{
		Data:      string(dataBytes),
		EventType: agriroutertestcontainer.SendMessagesTestEvent,
	}

	return SendMessages200Response{}, nil
}

func (s *Server) ConfirmMessages(_ context.Context, request ConfirmMessagesRequestObject) (ConfirmMessagesResponseObject, error) {
	dataBytes, err := json.Marshal(request.Body)
	if err != nil {
		return nil, err
	}
	s.events <- struct {
		Data      string
		EventType string
	}{
		Data:      string(dataBytes),
		EventType: agriroutertestcontainer.ConfirmMessagesTestEvent,
	}

	return ConfirmMessages202Response{}, nil
}

func (s *Server) PutEndpoint(ctx context.Context, request PutEndpointRequestObject) (PutEndpointResponseObject, error) {
	s.events <- struct {
		Data      string
		EventType string
	}{
		Data:      fmt.Sprintf(`{"externalId": "%s"}`, request.ExternalId),
		EventType: agriroutertestcontainer.PutEndpointTestEvent,
	}

	return PutEndpoint200JSONResponse{
		ExternalId: request.ExternalId,
	}, nil
}

func isFileMessageType(messageType string) bool {
	switch messageType {
	case "iso:11783:-10:taskdata:zip":
		return true
	case "shp:shape:zip":
		return true
	case "doc:pdf":
		return true
	case "img:jpeg":
		return true
	case "img:png":
		return true
	case "img:bmp":
		return true
	case "vid:avi":
		return true
	case "vid:mp4":
		return true
	case "vid:wmv":
		return true
	}
	return false
}

func (s *Server) GetTestEvents() <-chan struct {
	Data      string
	EventType string
} {
	return s.events
}

func NewServer() *Server {
	return &Server{
		events: make(chan struct {
			Data      string
			EventType string
		}),
		sentMessagesTestEvents: make(chan *SendMessagesTestEventData, 100),
	}
}
