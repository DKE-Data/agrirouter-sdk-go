package agrirouter

import (
	"errors"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/agmasync"
)

// ErrMasterdataCallFailed is returned when a master-data (AgmaSync) API call
// fails. Every master-data error matches it; the sentinels below say why.
var ErrMasterdataCallFailed = errors.New("master-data API call failed")

// Sentinel errors for the master-data responses a participant branches on.
// Compare with errors.Is; errors.As into *RevisionConflict, *MappingConflict,
// or *MasterdataAPIError for the detail.
var (
	// ErrRevisionConflict is a 412: the write's base revision cannot be
	// reconciled with the current one. RevisionConflict.CurrentRevision is the
	// revision to rebase onto.
	ErrRevisionConflict = agmasync.ErrRevisionConflict

	// ErrBaseRevisionRequired is a 428: an update was sent without a base
	// revision.
	ErrBaseRevisionRequired = agmasync.ErrBaseRevisionRequired

	// ErrMappingConflict is a 409 on a write or a binding: an identifier is
	// already bound to something else. MappingConflict.Rejection says which,
	// and why.
	ErrMappingConflict = agmasync.ErrMappingConflict

	// ErrMasterdataValidation is a 400: the payload was rejected, not repaired.
	ErrMasterdataValidation = agmasync.ErrValidation

	// ErrMasterdataForbidden is a 403: the acting endpoint is not entitled to
	// the object, or is not opted into its entity type.
	ErrMasterdataForbidden = agmasync.ErrForbidden

	// ErrMasterdataNotFound is a 404.
	ErrMasterdataNotFound = agmasync.ErrNotFound

	// ErrInitialLoadConflict is a 409 on the initial-load status resource: a
	// transition out of order, or user attention reported on a completed load.
	ErrInitialLoadConflict = agmasync.ErrInitialLoadConflict

	// ErrUnknownEntityType is raised locally for an entity type this SDK does
	// not know.
	ErrUnknownEntityType = agmasync.ErrUnknownEntityType

	// ErrLocalIDRequired is raised locally: an entity is sent under its local
	// id, so a Put with an empty localID cannot be addressed.
	ErrLocalIDRequired = agmasync.ErrLocalIDRequired

	// ErrEmptyResponse is a success status that carried no body to read.
	ErrEmptyResponse = agmasync.ErrEmptyResponse

	// ErrEntityTypeMismatch is an answer carrying an entity of another type
	// than the operation asked for.
	ErrEntityTypeMismatch = agmasync.ErrEntityTypeMismatch

	// ErrNotEventStream is a master-data stream answering 200 in a media type
	// other than text/event-stream, such as a proxy answering in agrirouter's
	// place.
	ErrNotEventStream = agmasync.ErrNotEventStream

	// ErrMasterdataStreamEnded is returned by StreamMasterdataEvents when the
	// live stream ends without the context being canceled: the connection
	// dropped or agrirouter closed it. Reconnect from the last durably applied
	// position.
	ErrMasterdataStreamEnded = errors.New("master-data stream ended")
)

// RevisionConflict is the error of a 412: a rejected write and the revision
// that stands.
type RevisionConflict = agmasync.RevisionConflict

// MappingConflict is the error of a mapping 409: a binding that could not be
// recorded, and why.
type MappingConflict = agmasync.MappingConflict

// MasterdataAPIError is any other non-success master-data response.
type MasterdataAPIError = agmasync.APIError

// Rejection reasons carried by MappingConflict.Rejection. The set is an
// extensible enumeration: tolerate a value not listed here.
const (
	ReasonLocalIDAlreadyBound      = agmasync.ReasonLocalIDAlreadyBound
	ReasonAgrirouterIDAlreadyBound = agmasync.ReasonAgrirouterIDAlreadyBound
	ReasonUnknownObject            = agmasync.ReasonUnknownObject
	ReasonDuplicateInRequest       = agmasync.ReasonDuplicateInRequest
)

// NeedsUser reports whether a rejection is one only a person can settle.
func NeedsUser(r IDMappingRejection) bool { return agmasync.NeedsUser(r) }
