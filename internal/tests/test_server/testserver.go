package test_server

import (
	"context"
	"fmt"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/tests/agriroutertestcontainer"
)

var _ StrictServerInterface = (*Server)(nil)

type Server struct {
	events chan struct {
		Data      string
		EventType string
	}
}

// ReceiveMessages implements StrictServerInterface.
func (s *Server) ReceiveMessages(ctx context.Context, request ReceiveMessagesRequestObject) (ReceiveMessagesResponseObject, error) {
	panic("unimplemented")
}

// SendMessages implements StrictServerInterface.
func (s *Server) SendMessages(ctx context.Context, request SendMessagesRequestObject) (SendMessagesResponseObject, error) {
	panic("unimplemented")
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
	}
}
