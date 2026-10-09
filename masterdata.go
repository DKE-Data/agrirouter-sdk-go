package agrirouter

import (
	"context"
	"fmt"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/agmasync"
	"github.com/google/uuid"
)

// ------------------------------------------------------------- any entity type

// The operations that carry only identifiers take the entity type as a value:
// a participant handling every type holds it as data, as Envelope.Type on a
// delivery or in its own store and send queue.

// BindMapping declares that the canonical object agrirouterID of entityType is
// the one this endpoint already knows as localID, so a later PUT under that
// localID updates it instead of creating a duplicate.
func (c *Client) BindMapping(
	ctx context.Context,
	entityType EntityType,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	if err := agmasync.Bind(ctx, c.oapiClient, endpointID, tenantID, entityType, localID, agrirouterID); err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return nil
}

// UnbindMapping declares that this endpoint no longer holds the canonical
// object of entityType under localID. The canonical object itself is untouched.
func (c *Client) UnbindMapping(
	ctx context.Context,
	entityType EntityType,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	if err := agmasync.Unbind(ctx, c.oapiClient, endpointID, tenantID, entityType, localID, agrirouterID); err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return nil
}

// DeactivateEntity signals that the entity of entityType under localID was
// deactivated in the source system. The canonical object is kept but marked
// inactive; the result is the canonical object as it now stands.
func (c *Client) DeactivateEntity(
	ctx context.Context,
	entityType EntityType,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
) (*MasterdataObject, error) {
	res, err := agmasync.Deactivate(ctx, c.oapiClient, endpointID, tenantID, entityType, localID, baseRevision)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return &res, nil
}

// RequestEntity refetches an object of entityType the calling endpoint is
// entitled to but does not currently hold. The object arrives asynchronously on the
// master-data event stream (see StreamMasterdataEvents).
func (c *Client) RequestEntity(
	ctx context.Context,
	entityType EntityType,
	endpointID, tenantID, agrirouterID uuid.UUID,
) error {
	if err := agmasync.Request(ctx, c.oapiClient, endpointID, tenantID, entityType, agrirouterID); err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return nil
}

// ------------------------------------------------------------------------ writes

// PutParty sends (creates or updates) a party identified by the
// application's own localID. On an update to an existing object, baseRevision
// must carry the revision it was edited from; pass nil when creating.
func (c *Client) PutParty(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
	party *Party,
) (*Party, error) {
	v := *party
	v.LocalId = &localID
	res, err := agmasync.PutParty(ctx, c.oapiClient, endpointID, tenantID, v, baseRevision)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return &res, nil
}

// PutFarm sends (creates or updates) a farm. See PutParty for baseRevision.
func (c *Client) PutFarm(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
	farm *Farm,
) (*Farm, error) {
	v := *farm
	v.LocalId = &localID
	res, err := agmasync.PutFarm(ctx, c.oapiClient, endpointID, tenantID, v, baseRevision)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return &res, nil
}

// PutField sends (creates or updates) a field. See PutParty for baseRevision.
func (c *Client) PutField(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
	field *Field,
) (*Field, error) {
	v := *field
	v.LocalId = &localID
	res, err := agmasync.PutField(ctx, c.oapiClient, endpointID, tenantID, v, baseRevision)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return &res, nil
}

// PutFieldBoundary sends (creates or updates) a field boundary. See PutParty for baseRevision.
func (c *Client) PutFieldBoundary(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
	fieldBoundary *FieldBoundary,
) (*FieldBoundary, error) {
	v := *fieldBoundary
	v.LocalId = &localID
	res, err := agmasync.PutFieldBoundary(ctx, c.oapiClient, endpointID, tenantID, v, baseRevision)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return &res, nil
}

// ---------------------------------------------------------------- initial load

// GetInitialLoadStatus returns the endpoint's initial-load state across every
// entity type it is opted into. The endpoint is addressed by the application's
// own externalEndpointID.
func (c *Client) GetInitialLoadStatus(
	ctx context.Context,
	externalEndpointID string,
	tenantID uuid.UUID,
) (*InitialLoadStatus, error) {
	s, err := agmasync.GetInitialLoadStatus(ctx, c.oapiClient, externalEndpointID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return &s, nil
}

// SetInitialLoadState advances the endpoint's initial-load state and, when
// confirming reconciliation, carries the identifier bindings it produced.
func (c *Client) SetInitialLoadState(
	ctx context.Context,
	externalEndpointID string,
	tenantID uuid.UUID,
	update InitialLoadStateUpdate,
) (*InitialLoadStatus, error) {
	s, err := agmasync.SetInitialLoadState(ctx, c.oapiClient, externalEndpointID, tenantID, update)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return &s, nil
}

// ReportUserAttention raises awaiting_user on the endpoint's initial-load
// status, telling the user the load is waiting for them in the application. It
// names no state, so it may be sent from any state before COMPLETED; agrirouter
// clears the flag on the endpoint's next transition.
func (c *Client) ReportUserAttention(
	ctx context.Context,
	externalEndpointID string,
	tenantID uuid.UUID,
) (*InitialLoadStatus, error) {
	s, err := agmasync.ReportUserAttention(ctx, c.oapiClient, externalEndpointID, tenantID)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return &s, nil
}

// --------------------------------------------------------------- event streams

// MasterdataEvent is one decoded frame of a master-data stream.
type MasterdataEvent = agmasync.Event

// MasterdataObject is a canonical object of any entity type, decoded into its
// model: Envelope.Type says which of Party, Farm, Field, and FieldBoundary is
// set.
type MasterdataObject = agmasync.Object

// MasterdataEnvelope holds the fields common to every entity type.
type MasterdataEnvelope = agmasync.Envelope

// StreamMasterdataEvents opens the persistent master-data change stream for
// every tenant and entity type the application is opted into, invoking handler
// for each received frame. Pass a non-empty lastEventID to resume after the last
// durably applied frame; pass "" for a first connection, which redelivers
// everything the application is entitled to and ends with a CAUGHT_UP frame.
//
// This call blocks until the context is canceled or an error occurs, so it is
// typically run in its own goroutine; the terminal error is returned.
func (c *Client) StreamMasterdataEvents(
	ctx context.Context,
	lastEventID string,
	handler func(ctx context.Context, event *MasterdataEvent),
) error {
	s, err := agmasync.Events(ctx, c.oapiClient, lastEventID)
	return consume(ctx, s, err, handler)
}

// StreamInitialLoadEvents streams every object of every opted-in entity type the
// endpoint is entitled to, while it is loading from agrirouter. agrirouter closes
// the response once the whole set has been sent; a dropped connection is recovered
// by requesting the set again from the beginning (this stream carries no position).
//
// This call blocks until the context is canceled or an error occurs, so it is
// typically run in its own goroutine; the terminal error is returned.
func (c *Client) StreamInitialLoadEvents(
	ctx context.Context,
	externalEndpointID string,
	tenantID uuid.UUID,
	handler func(ctx context.Context, event *MasterdataEvent),
) error {
	s, err := agmasync.InitialLoadEvents(ctx, c.oapiClient, externalEndpointID, tenantID)
	return consume(ctx, s, err, handler)
}

func consume(ctx context.Context, s *agmasync.Stream, err error, handler func(ctx context.Context, event *MasterdataEvent)) error {
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	defer func() { _ = s.Close() }()
	for ev, err := range s.Events() {
		if err != nil {
			return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
		}
		handler(ctx, &ev)
	}
	return nil
}
