package agrirouter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/oapi"
	internal_models "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
	"github.com/tmaxmax/go-sse"
)

// ErrMasterdataCallFailed is returned when a master-data (AgmaSync) API call fails.
var ErrMasterdataCallFailed = errors.New("master-data API call failed")

func (c *Client) masterdataErr(res *http.Response, body []byte) error {
	return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, httpResponseToErr(res, body))
}

// ---------------------------------------------------------------------- parties

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
	res, err := c.oapiClient.PutPartyWithResponse(ctx, localID, &internal_models.PutPartyParams{
		XAgrirouterEndpointId:   endpointID,
		XAgrirouterTenantId:     tenantID,
		XAgrirouterBaseRevision: baseRevision,
	}, *party)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	if res.JSON201 != nil {
		return res.JSON201, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
}

// BindPartyMapping declares that the canonical party agrirouterID
// is the one this endpoint already knows as localID, so a later PUT under that
// localID updates it instead of creating a duplicate.
func (c *Client) BindPartyMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	res, err := c.oapiClient.BindPartyMappingWithResponse(ctx, localID, agrirouterID, &internal_models.BindPartyMappingParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusNoContent {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// UnbindPartyMapping declares that this endpoint no longer holds the
// canonical party under localID. The canonical object itself is untouched.
func (c *Client) UnbindPartyMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	params := &internal_models.UnbindPartyMappingParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	}
	res, err := c.oapiClient.UnbindPartyMappingWithResponse(ctx, localID, agrirouterID, params)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusNoContent {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// DeactivateParty signals that the party was deactivated in the
// source system. The canonical object is kept but marked inactive.
func (c *Client) DeactivateParty(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
) (*Party, error) {
	res, err := c.oapiClient.DeactivatePartyWithResponse(ctx, localID, &internal_models.DeactivatePartyParams{
		XAgrirouterEndpointId:   endpointID,
		XAgrirouterTenantId:     tenantID,
		XAgrirouterBaseRevision: baseRevision,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
}

// RequestParty refetches a party the calling endpoint is entitled
// to but does not currently hold. The object arrives asynchronously on the
// master-data event stream (see StreamMasterdataEvents).
func (c *Client) RequestParty(
	ctx context.Context,
	endpointID, tenantID, agrirouterID uuid.UUID,
) error {
	res, err := c.oapiClient.RequestPartyWithResponse(ctx, &internal_models.RequestPartyParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	}, internal_models.EntityRequest{AgrirouterId: agrirouterID})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusAccepted {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// ------------------------------------------------------------------------ farms

// PutFarm sends (creates or updates) a farm. See PutParty for baseRevision.
func (c *Client) PutFarm(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
	farm *Farm,
) (*Farm, error) {
	res, err := c.oapiClient.PutFarmWithResponse(ctx, localID, &internal_models.PutFarmParams{
		XAgrirouterEndpointId:   endpointID,
		XAgrirouterTenantId:     tenantID,
		XAgrirouterBaseRevision: baseRevision,
	}, *farm)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	if res.JSON201 != nil {
		return res.JSON201, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
}

// BindFarmMapping binds a local identifier to an existing farm. See BindPartyMapping.
func (c *Client) BindFarmMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	res, err := c.oapiClient.BindFarmMappingWithResponse(ctx, localID, agrirouterID, &internal_models.BindFarmMappingParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusNoContent {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// UnbindFarmMapping declares that this endpoint no longer holds a farm. See UnbindPartyMapping.
func (c *Client) UnbindFarmMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	res, err := c.oapiClient.UnbindFarmMappingWithResponse(ctx, localID, agrirouterID, &internal_models.UnbindFarmMappingParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusNoContent {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// DeactivateFarm deactivates a farm. See DeactivateParty.
func (c *Client) DeactivateFarm(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
) (*Farm, error) {
	res, err := c.oapiClient.DeactivateFarmWithResponse(ctx, localID, &internal_models.DeactivateFarmParams{
		XAgrirouterEndpointId:   endpointID,
		XAgrirouterTenantId:     tenantID,
		XAgrirouterBaseRevision: baseRevision,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
}

// RequestFarm refetches a farm by its canonical id. See RequestParty.
func (c *Client) RequestFarm(
	ctx context.Context,
	endpointID, tenantID, agrirouterID uuid.UUID,
) error {
	res, err := c.oapiClient.RequestFarmWithResponse(ctx, &internal_models.RequestFarmParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	}, internal_models.EntityRequest{AgrirouterId: agrirouterID})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusAccepted {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// ----------------------------------------------------------------------- fields

// PutField sends (creates or updates) a field. See PutParty for baseRevision.
func (c *Client) PutField(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
	field *Field,
) (*Field, error) {
	res, err := c.oapiClient.PutFieldWithResponse(ctx, localID, &internal_models.PutFieldParams{
		XAgrirouterEndpointId:   endpointID,
		XAgrirouterTenantId:     tenantID,
		XAgrirouterBaseRevision: baseRevision,
	}, *field)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	if res.JSON201 != nil {
		return res.JSON201, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
}

// BindFieldMapping binds a local identifier to an existing field. See BindPartyMapping.
func (c *Client) BindFieldMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	res, err := c.oapiClient.BindFieldMappingWithResponse(ctx, localID, agrirouterID, &internal_models.BindFieldMappingParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusNoContent {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// UnbindFieldMapping declares that this endpoint no longer holds a field. See UnbindPartyMapping.
func (c *Client) UnbindFieldMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	res, err := c.oapiClient.UnbindFieldMappingWithResponse(ctx, localID, agrirouterID, &internal_models.UnbindFieldMappingParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusNoContent {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// DeactivateField deactivates a field. See DeactivateParty.
func (c *Client) DeactivateField(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
) (*Field, error) {
	res, err := c.oapiClient.DeactivateFieldWithResponse(ctx, localID, &internal_models.DeactivateFieldParams{
		XAgrirouterEndpointId:   endpointID,
		XAgrirouterTenantId:     tenantID,
		XAgrirouterBaseRevision: baseRevision,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
}

// RequestField refetches a field by its canonical id. See RequestParty.
func (c *Client) RequestField(
	ctx context.Context,
	endpointID, tenantID, agrirouterID uuid.UUID,
) error {
	res, err := c.oapiClient.RequestFieldWithResponse(ctx, &internal_models.RequestFieldParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	}, internal_models.EntityRequest{AgrirouterId: agrirouterID})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusAccepted {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// -------------------------------------------------------------- field boundaries

// PutFieldBoundary sends (creates or updates) a field boundary. See PutParty for baseRevision.
func (c *Client) PutFieldBoundary(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
	fieldBoundary *FieldBoundary,
) (*FieldBoundary, error) {
	res, err := c.oapiClient.PutFieldBoundaryWithResponse(ctx, localID, &internal_models.PutFieldBoundaryParams{
		XAgrirouterEndpointId:   endpointID,
		XAgrirouterTenantId:     tenantID,
		XAgrirouterBaseRevision: baseRevision,
	}, *fieldBoundary)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	if res.JSON201 != nil {
		return res.JSON201, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
}

// BindFieldBoundaryMapping binds a local identifier to an existing field boundary. See BindPartyMapping.
func (c *Client) BindFieldBoundaryMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	res, err := c.oapiClient.BindFieldBoundaryMappingWithResponse(ctx, localID, agrirouterID, &internal_models.BindFieldBoundaryMappingParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusNoContent {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// UnbindFieldBoundaryMapping declares that this endpoint no longer holds a field boundary. See UnbindPartyMapping.
func (c *Client) UnbindFieldBoundaryMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	params := &internal_models.UnbindFieldBoundaryMappingParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	}
	res, err := c.oapiClient.UnbindFieldBoundaryMappingWithResponse(ctx, localID, agrirouterID, params)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusNoContent {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// DeactivateFieldBoundary deactivates a field boundary. See DeactivateParty.
func (c *Client) DeactivateFieldBoundary(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
) (*FieldBoundary, error) {
	res, err := c.oapiClient.DeactivateFieldBoundaryWithResponse(ctx, localID, &internal_models.DeactivateFieldBoundaryParams{
		XAgrirouterEndpointId:   endpointID,
		XAgrirouterTenantId:     tenantID,
		XAgrirouterBaseRevision: baseRevision,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
}

// RequestFieldBoundary refetches a field boundary by its canonical id. See RequestParty.
func (c *Client) RequestFieldBoundary(
	ctx context.Context,
	endpointID, tenantID, agrirouterID uuid.UUID,
) error {
	res, err := c.oapiClient.RequestFieldBoundaryWithResponse(ctx, &internal_models.RequestFieldBoundaryParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	}, internal_models.EntityRequest{AgrirouterId: agrirouterID})
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusAccepted {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
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
	res, err := c.oapiClient.GetInitialLoadStatusWithResponse(ctx, externalEndpointID, &internal_models.GetInitialLoadStatusParams{
		XAgrirouterTenantId: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
}

// SetInitialLoadState advances the endpoint's initial-load state and, when
// confirming reconciliation, carries the identifier bindings it produced.
func (c *Client) SetInitialLoadState(
	ctx context.Context,
	externalEndpointID string,
	tenantID uuid.UUID,
	update InitialLoadStateUpdate,
) (*InitialLoadStatus, error) {
	res, err := c.oapiClient.SetInitialLoadStateWithResponse(ctx, externalEndpointID, &internal_models.SetInitialLoadStateParams{
		XAgrirouterTenantId: tenantID,
	}, update)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
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
	res, err := c.oapiClient.ReportUserAttentionWithResponse(ctx, externalEndpointID, &internal_models.ReportUserAttentionParams{
		XAgrirouterTenantId: tenantID,
	})
	if err != nil {
		return nil, fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.JSON200 != nil {
		return res.JSON200, nil
	}
	return nil, c.masterdataErr(res.HTTPResponse, res.Body)
}

// --------------------------------------------------------------- event streams

// MasterdataEventFrame is a single Server-Sent Event frame from a master-data
// stream. The frame payload is delivered verbatim: the master-data event schema
// is not modeled by this API, so Data carries the raw JSON as sent by agrirouter.
type MasterdataEventFrame struct {
	// ID is the SSE id: field. For StreamMasterdataEvents, persist it once the
	// frame is durably applied and send it back as lastEventID to resume; it is
	// opaque and must not be interpreted, compared, or modified.
	ID string
	// Event is the SSE event: field (the event type), empty for an unnamed event.
	Event string
	// Data is the SSE data: field, the raw JSON payload of the event.
	Data string
}

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
	handler func(ctx context.Context, frame *MasterdataEventFrame),
) error {
	params := &internal_models.StreamMasterdataEventsParams{}
	if lastEventID != "" {
		params.LastEventID = &lastEventID
	}
	req, err := oapi.NewStreamMasterdataEventsRequest(c.serverURL.String(), params)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return c.streamMasterdataFrames(ctx, req, handler)
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
	handler func(ctx context.Context, frame *MasterdataEventFrame),
) error {
	params := &internal_models.StreamInitialLoadEventsParams{XAgrirouterTenantId: tenantID}
	req, err := oapi.NewStreamInitialLoadEventsRequest(c.serverURL.String(), externalEndpointID, params)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	return c.streamMasterdataFrames(ctx, req, handler)
}

func (c *Client) streamMasterdataFrames(
	ctx context.Context,
	req *http.Request,
	handler func(ctx context.Context, frame *MasterdataEventFrame),
) error {
	req = req.WithContext(ctx)
	client := sse.DefaultClient
	client.ResponseValidator = func(r *http.Response) error {
		if err := sse.DefaultValidator(r); err != nil {
			body, _ := io.ReadAll(r.Body)
			return fmt.Errorf("%w: %v", err, string(body))
		}
		return nil
	}
	httpClient := c.oapiClient.ClientInterface.(*oapi.Client).Client
	client.HTTPClient = httpClient.(*http.Client)
	conn := client.NewConnection(req)
	unsubscribe := conn.SubscribeToAll(func(event sse.Event) {
		handler(ctx, &MasterdataEventFrame{
			ID:    event.LastEventID,
			Event: event.Type,
			Data:  event.Data,
		})
	})
	defer unsubscribe()
	return conn.Connect()
}
