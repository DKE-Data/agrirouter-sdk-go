package test_server

import (
	"context"
	"strings"
)

// Master-data (AgmaSync) stubs. The integration tests do not exercise the
// master-data surface; these implementations only satisfy StrictServerInterface
// so the test server compiles and serves the rest of the API. Each returns the
// operation's success response with a zero-valued body.

// ---------------------------------------------------------------- organizations

func (s *Server) PutOrganization(_ context.Context, _ PutOrganizationRequestObject) (PutOrganizationResponseObject, error) {
	return PutOrganization200JSONResponse{}, nil
}

func (s *Server) BindOrganizationMapping(_ context.Context, _ BindOrganizationMappingRequestObject) (BindOrganizationMappingResponseObject, error) {
	return BindOrganizationMapping204Response{}, nil
}

func (s *Server) UnbindOrganizationMapping(_ context.Context, _ UnbindOrganizationMappingRequestObject) (UnbindOrganizationMappingResponseObject, error) {
	return UnbindOrganizationMapping204Response{}, nil
}

func (s *Server) DeactivateOrganization(_ context.Context, _ DeactivateOrganizationRequestObject) (DeactivateOrganizationResponseObject, error) {
	return DeactivateOrganization200JSONResponse{}, nil
}

func (s *Server) RequestOrganization(_ context.Context, _ RequestOrganizationRequestObject) (RequestOrganizationResponseObject, error) {
	return RequestOrganization202Response{}, nil
}

// ---------------------------------------------------------------------- persons

func (s *Server) PutPerson(_ context.Context, _ PutPersonRequestObject) (PutPersonResponseObject, error) {
	return PutPerson200JSONResponse{}, nil
}

func (s *Server) BindPersonMapping(_ context.Context, _ BindPersonMappingRequestObject) (BindPersonMappingResponseObject, error) {
	return BindPersonMapping204Response{}, nil
}

func (s *Server) UnbindPersonMapping(_ context.Context, _ UnbindPersonMappingRequestObject) (UnbindPersonMappingResponseObject, error) {
	return UnbindPersonMapping204Response{}, nil
}

func (s *Server) DeactivatePerson(_ context.Context, _ DeactivatePersonRequestObject) (DeactivatePersonResponseObject, error) {
	return DeactivatePerson200JSONResponse{}, nil
}

func (s *Server) RequestPerson(_ context.Context, _ RequestPersonRequestObject) (RequestPersonResponseObject, error) {
	return RequestPerson202Response{}, nil
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
