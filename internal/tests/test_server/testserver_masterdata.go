package test_server

import (
	"context"
	"strings"
)

// Master-data (AgmaSync) stubs. The integration tests do not exercise the
// master-data surface; these implementations only satisfy StrictServerInterface
// so the test server compiles and serves the rest of the API. Each returns the
// operation's success response with a zero-valued body.

// ---------------------------------------------------------------------- parties

func (s *Server) PutParty(_ context.Context, _ PutPartyRequestObject) (PutPartyResponseObject, error) {
	return PutParty200JSONResponse{}, nil
}

func (s *Server) BindPartyMapping(_ context.Context, _ BindPartyMappingRequestObject) (BindPartyMappingResponseObject, error) {
	return BindPartyMapping204Response{}, nil
}

func (s *Server) UnbindPartyMapping(_ context.Context, _ UnbindPartyMappingRequestObject) (UnbindPartyMappingResponseObject, error) {
	return UnbindPartyMapping204Response{}, nil
}

func (s *Server) DeactivateParty(_ context.Context, _ DeactivatePartyRequestObject) (DeactivatePartyResponseObject, error) {
	return DeactivateParty200JSONResponse{}, nil
}

func (s *Server) RequestParty(_ context.Context, _ RequestPartyRequestObject) (RequestPartyResponseObject, error) {
	return RequestParty202Response{}, nil
}

// ------------------------------------------------------------------------ farms

func (s *Server) PutFarm(_ context.Context, _ PutFarmRequestObject) (PutFarmResponseObject, error) {
	return PutFarm200JSONResponse{}, nil
}

func (s *Server) BindFarmMapping(_ context.Context, _ BindFarmMappingRequestObject) (BindFarmMappingResponseObject, error) {
	return BindFarmMapping204Response{}, nil
}

func (s *Server) UnbindFarmMapping(_ context.Context, _ UnbindFarmMappingRequestObject) (UnbindFarmMappingResponseObject, error) {
	return UnbindFarmMapping204Response{}, nil
}

func (s *Server) DeactivateFarm(_ context.Context, _ DeactivateFarmRequestObject) (DeactivateFarmResponseObject, error) {
	return DeactivateFarm200JSONResponse{}, nil
}

func (s *Server) RequestFarm(_ context.Context, _ RequestFarmRequestObject) (RequestFarmResponseObject, error) {
	return RequestFarm202Response{}, nil
}

// ----------------------------------------------------------------------- fields

func (s *Server) PutField(_ context.Context, _ PutFieldRequestObject) (PutFieldResponseObject, error) {
	return PutField200JSONResponse{}, nil
}

func (s *Server) BindFieldMapping(_ context.Context, _ BindFieldMappingRequestObject) (BindFieldMappingResponseObject, error) {
	return BindFieldMapping204Response{}, nil
}

func (s *Server) UnbindFieldMapping(_ context.Context, _ UnbindFieldMappingRequestObject) (UnbindFieldMappingResponseObject, error) {
	return UnbindFieldMapping204Response{}, nil
}

func (s *Server) DeactivateField(_ context.Context, _ DeactivateFieldRequestObject) (DeactivateFieldResponseObject, error) {
	return DeactivateField200JSONResponse{}, nil
}

func (s *Server) RequestField(_ context.Context, _ RequestFieldRequestObject) (RequestFieldResponseObject, error) {
	return RequestField202Response{}, nil
}

// -------------------------------------------------------------- field boundaries

func (s *Server) PutFieldBoundary(_ context.Context, _ PutFieldBoundaryRequestObject) (PutFieldBoundaryResponseObject, error) {
	return PutFieldBoundary200JSONResponse{}, nil
}

func (s *Server) BindFieldBoundaryMapping(_ context.Context, _ BindFieldBoundaryMappingRequestObject) (BindFieldBoundaryMappingResponseObject, error) {
	return BindFieldBoundaryMapping204Response{}, nil
}

func (s *Server) UnbindFieldBoundaryMapping(_ context.Context, _ UnbindFieldBoundaryMappingRequestObject) (UnbindFieldBoundaryMappingResponseObject, error) {
	return UnbindFieldBoundaryMapping204Response{}, nil
}

func (s *Server) DeactivateFieldBoundary(_ context.Context, _ DeactivateFieldBoundaryRequestObject) (DeactivateFieldBoundaryResponseObject, error) {
	return DeactivateFieldBoundary200JSONResponse{}, nil
}

func (s *Server) RequestFieldBoundary(_ context.Context, _ RequestFieldBoundaryRequestObject) (RequestFieldBoundaryResponseObject, error) {
	return RequestFieldBoundary202Response{}, nil
}

// ---------------------------------------------------------------- initial load

func (s *Server) GetInitialLoadStatus(_ context.Context, _ GetInitialLoadStatusRequestObject) (GetInitialLoadStatusResponseObject, error) {
	return GetInitialLoadStatus200JSONResponse{}, nil
}

func (s *Server) ReportUserAttention(_ context.Context, _ ReportUserAttentionRequestObject) (ReportUserAttentionResponseObject, error) {
	return ReportUserAttention200JSONResponse{}, nil
}

func (s *Server) SetInitialLoadState(_ context.Context, _ SetInitialLoadStateRequestObject) (SetInitialLoadStateResponseObject, error) {
	return SetInitialLoadState200JSONResponse{}, nil
}

// --------------------------------------------------------------- event streams

func (s *Server) StreamMasterdataEvents(_ context.Context, _ StreamMasterdataEventsRequestObject) (StreamMasterdataEventsResponseObject, error) {
	return StreamMasterdataEvents200TexteventStreamResponse{Body: strings.NewReader("")}, nil
}

func (s *Server) StreamInitialLoadEvents(_ context.Context, _ StreamInitialLoadEventsRequestObject) (StreamInitialLoadEventsResponseObject, error) {
	return StreamInitialLoadEvents200TexteventStreamResponse{Body: strings.NewReader("")}, nil
}
