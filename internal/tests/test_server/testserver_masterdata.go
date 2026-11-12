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

func (s *Server) CorrectPartyMappings(_ context.Context, _ CorrectPartyMappingsRequestObject) (CorrectPartyMappingsResponseObject, error) {
	return CorrectPartyMappings204Response{}, nil
}

func (s *Server) ArchiveParty(_ context.Context, _ ArchivePartyRequestObject) (ArchivePartyResponseObject, error) {
	return ArchiveParty200JSONResponse{}, nil
}

func (s *Server) UnarchiveParty(_ context.Context, _ UnarchivePartyRequestObject) (UnarchivePartyResponseObject, error) {
	return UnarchiveParty200JSONResponse{}, nil
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

func (s *Server) CorrectFarmMappings(_ context.Context, _ CorrectFarmMappingsRequestObject) (CorrectFarmMappingsResponseObject, error) {
	return CorrectFarmMappings204Response{}, nil
}

func (s *Server) ArchiveFarm(_ context.Context, _ ArchiveFarmRequestObject) (ArchiveFarmResponseObject, error) {
	return ArchiveFarm200JSONResponse{}, nil
}

func (s *Server) UnarchiveFarm(_ context.Context, _ UnarchiveFarmRequestObject) (UnarchiveFarmResponseObject, error) {
	return UnarchiveFarm200JSONResponse{}, nil
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

func (s *Server) CorrectFieldMappings(_ context.Context, _ CorrectFieldMappingsRequestObject) (CorrectFieldMappingsResponseObject, error) {
	return CorrectFieldMappings204Response{}, nil
}

func (s *Server) ArchiveField(_ context.Context, _ ArchiveFieldRequestObject) (ArchiveFieldResponseObject, error) {
	return ArchiveField200JSONResponse{}, nil
}

func (s *Server) UnarchiveField(_ context.Context, _ UnarchiveFieldRequestObject) (UnarchiveFieldResponseObject, error) {
	return UnarchiveField200JSONResponse{}, nil
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

func (s *Server) CorrectFieldBoundaryMappings(_ context.Context, _ CorrectFieldBoundaryMappingsRequestObject) (CorrectFieldBoundaryMappingsResponseObject, error) {
	return CorrectFieldBoundaryMappings204Response{}, nil
}

func (s *Server) ArchiveFieldBoundary(_ context.Context, _ ArchiveFieldBoundaryRequestObject) (ArchiveFieldBoundaryResponseObject, error) {
	return ArchiveFieldBoundary200JSONResponse{}, nil
}

func (s *Server) UnarchiveFieldBoundary(_ context.Context, _ UnarchiveFieldBoundaryRequestObject) (UnarchiveFieldBoundaryResponseObject, error) {
	return UnarchiveFieldBoundary200JSONResponse{}, nil
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
