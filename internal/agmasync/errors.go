package agmasync

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	oapi "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
)

// Sentinel errors for conditions a participant branches on. Compare with
// [errors.Is]; the concrete types below carry the detail.
var (
	// ErrRevisionConflict is a 412. The write was made from a base revision
	// that cannot be reconciled with the current one, or from a base
	// agrirouter never issued, or against a localId that resolves to nothing.
	// See [RevisionConflict].
	ErrRevisionConflict = errors.New("revision conflict")

	// ErrBaseRevisionRequired is a 428: an update was sent with no base
	// revision. Omitting the header would opt the participant out of
	// concurrency control, so agrirouter refuses rather than assuming.
	ErrBaseRevisionRequired = errors.New("base revision required")

	// ErrMappingConflict is a 409: an identifier in the request is already
	// bound to something else. See [MappingConflict].
	ErrMappingConflict = errors.New("identifier mapping conflict")

	// ErrValidation is a 400. agrirouter rejects an invalid payload rather
	// than repairing it.
	ErrValidation = errors.New("validation failed")

	// ErrForbidden is a 403: the acting endpoint is not entitled to the
	// object, or is not opted into its entity type.
	ErrForbidden = errors.New("forbidden")

	// ErrNotFound is a 404.
	ErrNotFound = errors.New("not found")

	// ErrInitialLoadConflict is a 409 on the initial-load status resource: a
	// transition out of order. Repeating the endpoint's current state
	// succeeds.
	ErrInitialLoadConflict = errors.New("initial load transition out of order")

	// ErrLocalIDRequired is raised locally: an entity is sent under its
	// local_id, so a write without one cannot be addressed.
	ErrLocalIDRequired = errors.New("local_id is required on send")

	// ErrEmptyResponse is a success status that carried no body to read.
	ErrEmptyResponse = errors.New("empty response")

	// ErrUnknownEntityType is raised locally, not by agrirouter, when an
	// entity carries a type this version does not define.
	ErrUnknownEntityType = errors.New("unknown entity type")

	// ErrEntityTypeMismatch is an answer carrying an entity of another type
	// than the operation asked for.
	ErrEntityTypeMismatch = errors.New("entity type mismatch")

	// ErrNotEventStream is a stream answering 200 in a media type other than
	// text/event-stream, such as a proxy or an error page answering in
	// agrirouter's place.
	ErrNotEventStream = errors.New("not an event stream")
)

// RevisionConflict reports a rejected write and the revision that stands.
//
// To recover, obtain the current object, rebase the change onto it, and write
// again.
type RevisionConflict struct {
	// CurrentRevision is the canonical object's revision now and the one to
	// rebase onto. It is not necessarily base + 1: other participants may have
	// written more than once since.
	CurrentRevision int
	Message         string
}

func (e *RevisionConflict) Error() string {
	return fmt.Sprintf("agmasync: revision conflict, current revision is %d: %s",
		e.CurrentRevision, e.Message)
}

// Is makes [errors.Is] match [ErrRevisionConflict].
func (e *RevisionConflict) Is(target error) bool { return errors.Is(target, ErrRevisionConflict) }

// MappingConflict reports a binding that could not be recorded.
//
// Handling differs by cause: some causes need a user, others must never reach
// one (see [NeedsUser]). Rejection names the taken identifier and, where one
// exists, the conflicting mapping. Both ends of that mapping belong to the
// rejected participant, so it discloses nothing the participant does not
// already hold.
type MappingConflict struct {
	Rejection oapi.IdMappingRejection
	Message   string
}

func (e *MappingConflict) Error() string {
	return fmt.Sprintf("agmasync: mapping conflict (%s) for localId %q / agrirouterId %s: %s",
		e.Rejection.Reason, e.Rejection.LocalId, e.Rejection.AgrirouterId, e.Message)
}

// Is makes [errors.Is] match [ErrMappingConflict].
func (e *MappingConflict) Is(target error) bool { return errors.Is(target, ErrMappingConflict) }

// Rejection reasons. agrirouter may add reasons, so these are plain strings
// rather than a closed enum, and a participant must handle an unknown value
// in a default branch.
const (
	// ReasonLocalIDAlreadyBound means this application already knows a
	// different canonical object in the tenant by that localId. Resolving it
	// needs a user: two of their records are being claimed to be one. What the
	// application has bound in another tenant never causes this: the mapping is
	// keyed by the application and the tenant.
	ReasonLocalIDAlreadyBound = "LOCAL_ID_ALREADY_BOUND"

	// ReasonAgrirouterIDAlreadyBound means this application already knows that
	// canonical object under a different localId. The participant can usually
	// resolve it without a user.
	ReasonAgrirouterIDAlreadyBound = "AGRIROUTER_ID_ALREADY_BOUND"

	// ReasonUnknownObject means no such canonical object, or the endpoint is
	// not entitled to it.
	ReasonUnknownObject = "UNKNOWN_OBJECT"

	// ReasonDuplicateInRequest means one bulk confirmation named the same
	// identifier in more than one pair. Every pair involved is rejected and
	// none applied, since the outcome would otherwise depend on ordering.
	ReasonDuplicateInRequest = "DUPLICATE_IN_REQUEST"
)

// NeedsUser reports whether a rejection is one only a person can settle.
//
// A localId already bound to a different canonical object means two systems
// disagree on granularity (the n:1 case), which the protocol does not merge
// automatically. The participant fixes the other reasons itself: it already
// holds the object under another identifier, or it sent a malformed batch.
func NeedsUser(r oapi.IdMappingRejection) bool {
	return r.Reason == ReasonLocalIDAlreadyBound
}

// APIError is any other non-success response.
type APIError struct {
	StatusCode int
	Message    string
	sentinel   error
}

func (e *APIError) Error() string {
	return fmt.Sprintf("agmasync: %s (HTTP %d)", e.Message, e.StatusCode)
}

// Is matches the sentinel for the status code, where there is one.
func (e *APIError) Is(target error) bool {
	return e.sentinel != nil && errors.Is(target, e.sentinel)
}

// writeResult is the common shape of a write response. The generated response
// types differ per operation and entity type but are structurally identical, so
// each wrapper fills this in and shares the handling below.
type writeResult struct {
	statusCode int
	validation *oapi.MasterdataError
	forbidden  *oapi.MasterdataError
	notFound   *oapi.MasterdataError
	conflict   *oapi.MappingConflictError
	precond    *oapi.RevisionConflictError
	// required is the 428 body. It is shaped like the 412 body and carries the
	// revision the participant should have sent.
	required *oapi.RevisionConflictError
	body     []byte
}

// err converts a response into the typed error the participant branches on, or
// nil where the response was a success.
func (r writeResult) err() error {
	switch r.statusCode {
	case http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent:
		return nil
	case http.StatusBadRequest:
		return &APIError{r.statusCode, r.message(r.validation, "validation failed"), ErrValidation}
	case http.StatusForbidden:
		return &APIError{r.statusCode, r.message(r.forbidden, "forbidden"), ErrForbidden}
	case http.StatusNotFound:
		return &APIError{r.statusCode, r.message(r.notFound, "not found"), ErrNotFound}
	case http.StatusConflict:
		if r.conflict != nil {
			return &MappingConflict{Rejection: r.conflict.Rejection, Message: r.conflict.Message}
		}
		return &APIError{r.statusCode, "conflict", ErrMappingConflict}
	case http.StatusPreconditionFailed:
		if r.precond != nil {
			return &RevisionConflict{
				CurrentRevision: r.precond.CurrentRevision,
				Message:         r.precond.Message,
			}
		}
		return &APIError{r.statusCode, "precondition failed", ErrRevisionConflict}
	case http.StatusPreconditionRequired:
		msg := "base revision required"
		if r.required != nil && r.required.Message != "" {
			msg = r.required.Message
		}
		return &APIError{r.statusCode, msg, ErrBaseRevisionRequired}
	default:
		return &APIError{r.statusCode, unexpected(r.body), nil}
	}
}

// message is the error message agrirouter sent, or fallback where it sent none.
//
// e is the body as the generated client decoded it. It is nil for the streams,
// which bypass generated decoding, and where the operation declares no body for
// the status: only puts declare a 400, though agrirouter answers any malformed
// request with one. In those cases the raw body is decoded instead.
func (r writeResult) message(e *oapi.MasterdataError, fallback string) string {
	if e == nil {
		var raw oapi.MasterdataError
		if json.Unmarshal(r.body, &raw) == nil {
			e = &raw
		}
	}
	if e == nil || e.Message == "" {
		return fallback
	}
	return e.Message
}

func unexpected(body []byte) string {
	if len(body) == 0 {
		return "unexpected response"
	}
	const maxLen = 200
	if len(body) > maxLen {
		body = body[:maxLen]
	}
	return "unexpected response: " + string(body)
}

// mappingResult is the [writeResult] counterpart for the binding operations,
// which answer 204 and carry no entity.
type mappingResult struct {
	statusCode int
	forbidden  *oapi.MasterdataError
	notFound   *oapi.MasterdataError
	conflict   *oapi.MappingConflictError
	body       []byte
}

func (r mappingResult) err() error {
	return writeResult{
		statusCode: r.statusCode,
		forbidden:  r.forbidden,
		notFound:   r.notFound,
		conflict:   r.conflict,
		body:       r.body,
	}.err()
}

func transportErr(err error) error {
	return fmt.Errorf("agmasync: request failed: %w", err)
}
