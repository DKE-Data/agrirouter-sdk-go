package agmasync

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"mime"
	"net/http"

	oapiclient "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi"
	oapi "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
	"github.com/tmaxmax/go-sse"
)

// Event types carried on the streams.
const (
	// EventMasterdataChanged carries a canonical object.
	EventMasterdataChanged = "MASTERDATA_CHANGED"

	// EventMasterdataDeactivated carries a canonical object that has become
	// inactive. It is a lifecycle transition, not a removal: the object and
	// its identifier mapping are retained.
	EventMasterdataDeactivated = "MASTERDATA_DEACTIVATED"

	// EventCaughtUp marks the end of catch-up on the live stream. It carries
	// no entity and names no entity type: it covers the catch-up as a whole.
	EventCaughtUp = "CAUGHT_UP"

	// EventRouteChanged states which entity types the user has selected on one
	// endpoint. It carries an [oapi.RouteChangedEventData] rather than an
	// entity.
	EventRouteChanged = "ROUTE_CHANGED"

	// EventMasterdataReset states that the user wiped one tenant's master data:
	// every canonical object, identifier mapping, route and initial-load state
	// there is gone. It carries an [oapi.MasterdataResetEventData] rather than
	// an entity, and stands for an empty selection on every endpoint it lists.
	EventMasterdataReset = "RESET_MASTERDATA_SYNC"
)

// Event is one frame of a master-data stream.
type Event struct {
	// Type is one of the event constants above. It is a plain string because
	// a participant must tolerate a frame type it does not know rather than
	// fail on it.
	Type string

	// ID is the opaque delivery position as of this frame. Do not interpret,
	// compare, construct, or modify it; its structure may change between
	// versions. The same value may repeat across frames, as it does throughout
	// catch-up.
	//
	// It is safe to resume from once this frame and every earlier one have
	// been durably applied. It is empty on the initial-load stream, which has
	// no position.
	ID string

	// Object is the canonical object the frame carries. It is the zero value
	// on a frame that carries none, such as EventCaughtUp.
	Object

	// RouteChange is set on an [EventRouteChanged] frame and nil otherwise. It
	// holds the endpoint's whole selection after the change.
	RouteChange *oapi.RouteChangedEventData

	// Reset is set on an [EventMasterdataReset] frame and nil otherwise. On
	// receipt, a participant discards every binding it holds in the tenant,
	// keeps its local records, and durably stores this frame's position before
	// rejoining the tenant's initial load. Resuming from an earlier position
	// would deliver the reset again after the new bindings exist.
	Reset *oapi.MasterdataResetEventData
}

// HasEntity reports whether the frame carries a canonical object.
func (e Event) HasEntity() bool { return e.Envelope.Type != "" }

// Stream is an open master-data event stream.
//
// Close it when done; the underlying connection stays open until then.
type Stream struct {
	resp *http.Response

	// positioned is false for the initial-load stream, which delivers a fixed
	// set rather than a sequence of changes and therefore carries no position.
	positioned bool

	// MaxEventSize bounds one frame, in bytes; a larger frame ends the
	// iteration with an error. Zero means [DefaultMaxEventSize].
	MaxEventSize int
}

// Close releases the stream's connection.
func (s *Stream) Close() error {
	if s.resp == nil || s.resp.Body == nil {
		return nil
	}
	return s.resp.Body.Close()
}

// Events iterates the stream's frames until it ends, the context is canceled,
// or a frame cannot be decoded.
//
// On the live stream, the iteration ending means the connection dropped;
// reconnect from the last durably applied position. On the initial-load
// stream, it does not prove the set arrived; check [GetInitialLoadStatus].
func (s *Stream) Events() iter.Seq2[Event, error] {
	return func(yield func(Event, error) bool) {
		// sse.Read, not sse.Client as the SDK's other streams use: sse.Client
		// reconnects on its own from the last frame it read, but the live
		// stream must resume from the last frame the participant durably
		// applied, which only the participant knows.
		maxEventSize := s.MaxEventSize
		if maxEventSize <= 0 {
			maxEventSize = DefaultMaxEventSize
		}
		for raw, err := range sse.Read(s.resp.Body, &sse.ReadConfig{MaxEventSize: maxEventSize}) {
			if err != nil {
				if errIsStreamEnd(err) {
					return
				}
				yield(Event{}, fmt.Errorf("agmasync: reading event stream: %w", err))
				return
			}

			ev := Event{Type: raw.Type}
			if s.positioned {
				ev.ID = raw.LastEventID
			}

			switch {
			case raw.Data == "":

			case raw.Type == EventMasterdataChanged || raw.Type == EventMasterdataDeactivated:
				obj, err := ObjectOf([]byte(raw.Data))
				if err != nil {
					yield(Event{}, fmt.Errorf("agmasync: decoding %s frame: %w", raw.Type, err))
					return
				}
				ev.Object = obj

			case raw.Type == EventRouteChanged:
				var selection oapi.RouteChangedEventData
				if err := json.Unmarshal([]byte(raw.Data), &selection); err != nil {
					yield(Event{}, fmt.Errorf("agmasync: decoding %s frame: %w", raw.Type, err))
					return
				}
				ev.RouteChange = &selection

			case raw.Type == EventMasterdataReset:
				var reset oapi.MasterdataResetEventData
				if err := json.Unmarshal([]byte(raw.Data), &reset); err != nil {
					yield(Event{}, fmt.Errorf("agmasync: decoding %s frame: %w", raw.Type, err))
					return
				}
				ev.Reset = &reset
			}

			if !yield(ev, nil) {
				return
			}
		}
	}
}

// DefaultMaxEventSize is the frame bound a [Stream] applies unless told
// otherwise.
const DefaultMaxEventSize = 16 << 20

func errIsStreamEnd(err error) bool {
	return errors.Is(err, io.EOF)
}

// Events opens the application's live master-data stream.
//
// The stream is application-scoped: it carries every tenant the application is
// routed to and every entity type it is opted into, so a receiver partitions
// on each object's tenantId. agrirouter does not echo a change to the endpoint
// that wrote it, but does deliver it to the application's other endpoints.
//
// lastEventID must be the ID of a frame the participant has durably applied,
// not the last one it read, because delivery is at-least-once. A participant
// applying frames in parallel resends the ID of the earliest frame, in
// delivery order, that it has not yet applied. An empty lastEventID delivers
// everything the application is entitled to.
//
// Positions do not expire. agrirouter serves catch-up from each entity's
// current state rather than a retained log, so a long absence means a larger
// catch-up, and a participant sees each changed entity once with its current
// value, not every intermediate change. Catch-up ends with an [EventCaughtUp]
// frame.
func Events(ctx context.Context, api *oapiclient.ClientWithResponses, lastEventID string) (*Stream, error) {
	var params oapi.StreamMasterdataEventsParams
	if lastEventID != "" {
		params.LastEventID = &lastEventID
	}
	resp, err := api.StreamMasterdataEvents(ctx, &params, acceptEventStream)
	if err != nil {
		return nil, fmt.Errorf("agmasync: opening stream: %w", err)
	}
	return openStream(resp, true)
}

// InitialLoadEvents opens this endpoint's initial-load stream and collects the
// canonical set it is owed.
//
// The set is every object of every entity type the endpoint is entitled to.
//
// A referenced object precedes the objects that reference it, and opt-in is
// dependency-closed, so each object can be applied on arrival. A reference to
// an unselected type is ignored. Do not rely on any other property of the
// order, such as the position of one type relative to another or a type being
// complete.
//
// The stream carries no position. After a dropped connection, reconnect and
// take the set from the beginning.
func InitialLoadEvents(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	externalEndpointID string, tenantID uuid.UUID,
) (*Stream, error) {
	resp, err := api.StreamInitialLoadEvents(ctx, externalEndpointID,
		&oapi.StreamInitialLoadEventsParams{XAgrirouterTenantId: tenantID},
		acceptEventStream)
	if err != nil {
		return nil, fmt.Errorf("agmasync: opening stream: %w", err)
	}
	return openStream(resp, false)
}

// acceptEventStream asks the two stream operations for the media type they
// answer in, and says the response must not be cached.
var acceptEventStream oapiclient.RequestEditorFn = func(
	_ context.Context, req *http.Request,
) error {
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-store")
	return nil
}

// openStream turns the response of one of the generated stream operations into
// a [Stream].
//
// Callers use the generated low-level method, not the WithResponse wrapper,
// because the wrapper reads the whole body before returning.
//
// positioned says whether the stream carries a delivery position.
func openStream(resp *http.Response, positioned bool) (*Stream, error) {
	const maxErrorBody = 4096
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		_ = resp.Body.Close()
		return nil, writeResult{statusCode: resp.StatusCode, body: body}.err()
	}
	// Parsed as a stream, a 200 in another media type would yield no frames
	// and no error.
	if mt, _, err := mime.ParseMediaType(resp.Header.Get("Content-Type")); err != nil || mt != "text/event-stream" {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		_ = resp.Body.Close()
		return nil, &APIError{resp.StatusCode, fmt.Sprintf("stream answered as %q, not text/event-stream: %s",
			resp.Header.Get("Content-Type"), unexpected(body)), ErrNotEventStream}
	}
	return &Stream{resp: resp, positioned: positioned}, nil
}
