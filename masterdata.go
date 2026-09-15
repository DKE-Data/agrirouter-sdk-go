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

// ---------------------------------------------------------------- organizations

// PutOrganization sends (creates or updates) an organization identified by the
// application's own localID. On an update to an existing object, baseRevision
// must carry the revision it was edited from; pass nil when creating.
func (c *Client) PutOrganization(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
	org *Organization,
) (*Organization, error) {
	res, err := c.oapiClient.PutOrganizationWithResponse(ctx, localID, &internal_models.PutOrganizationParams{
		XAgrirouterEndpointId:   endpointID,
		XAgrirouterTenantId:     tenantID,
		XAgrirouterBaseRevision: baseRevision,
	}, *org)
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

// BindOrganizationMapping declares that the canonical organization agrirouterID
// is the one this endpoint already knows as localID, so a later PUT under that
// localID updates it instead of creating a duplicate.
func (c *Client) BindOrganizationMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	res, err := c.oapiClient.BindOrganizationMappingWithResponse(ctx, localID, agrirouterID, &internal_models.BindOrganizationMappingParams{
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

// UnbindOrganizationMapping declares that this endpoint no longer holds the
// canonical organization under localID. The canonical object itself is untouched.
func (c *Client) UnbindOrganizationMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	params := &internal_models.UnbindOrganizationMappingParams{
		XAgrirouterEndpointId: endpointID,
		XAgrirouterTenantId:   tenantID,
	}
	res, err := c.oapiClient.UnbindOrganizationMappingWithResponse(ctx, localID, agrirouterID, params)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrMasterdataCallFailed, err)
	}
	if res.StatusCode() == http.StatusNoContent {
		return nil
	}
	return c.masterdataErr(res.HTTPResponse, res.Body)
}

// DeactivateOrganization signals that the organization was deactivated in the
// source system. The canonical object is kept but marked inactive.
func (c *Client) DeactivateOrganization(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
) (*Organization, error) {
	res, err := c.oapiClient.DeactivateOrganizationWithResponse(ctx, localID, &internal_models.DeactivateOrganizationParams{
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

// RequestOrganization refetches an organization the calling endpoint is entitled
// to but does not currently hold. The object arrives asynchronously on the
// master-data event stream (see StreamMasterdataEvents).
func (c *Client) RequestOrganization(
	ctx context.Context,
	endpointID, tenantID, agrirouterID uuid.UUID,
) error {
	res, err := c.oapiClient.RequestOrganizationWithResponse(ctx, &internal_models.RequestOrganizationParams{
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

// ---------------------------------------------------------------------- persons

// PutPerson sends (creates or updates) a person. See PutOrganization for the
// meaning of baseRevision.
func (c *Client) PutPerson(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
	person *Person,
) (*Person, error) {
	res, err := c.oapiClient.PutPersonWithResponse(ctx, localID, &internal_models.PutPersonParams{
		XAgrirouterEndpointId:   endpointID,
		XAgrirouterTenantId:     tenantID,
		XAgrirouterBaseRevision: baseRevision,
	}, *person)
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

// BindPersonMapping binds a local identifier to an existing person. See BindOrganizationMapping.
func (c *Client) BindPersonMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	res, err := c.oapiClient.BindPersonMappingWithResponse(ctx, localID, agrirouterID, &internal_models.BindPersonMappingParams{
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

// UnbindPersonMapping declares that this endpoint no longer holds a person. See UnbindOrganizationMapping.
func (c *Client) UnbindPersonMapping(
	ctx context.Context,
	localID string,
	agrirouterID, endpointID, tenantID uuid.UUID,
) error {
	res, err := c.oapiClient.UnbindPersonMappingWithResponse(ctx, localID, agrirouterID, &internal_models.UnbindPersonMappingParams{
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

// DeactivatePerson deactivates a person. See DeactivateOrganization.
func (c *Client) DeactivatePerson(
	ctx context.Context,
	localID string,
	endpointID, tenantID uuid.UUID,
	baseRevision *int,
) (*Person, error) {
	res, err := c.oapiClient.DeactivatePersonWithResponse(ctx, localID, &internal_models.DeactivatePersonParams{
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

// RequestPerson refetches a person by its canonical id. See RequestOrganization.
func (c *Client) RequestPerson(
	ctx context.Context,
	endpointID, tenantID, agrirouterID uuid.UUID,
) error {
	res, err := c.oapiClient.RequestPersonWithResponse(ctx, &internal_models.RequestPersonParams{
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

// PutFarm sends (creates or updates) a farm. See PutOrganization for baseRevision.
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

// BindFarmMapping binds a local identifier to an existing farm. See BindOrganizationMapping.
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

// UnbindFarmMapping declares that this endpoint no longer holds a farm. See UnbindOrganizationMapping.
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

// DeactivateFarm deactivates a farm. See DeactivateOrganization.
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

// RequestFarm refetches a farm by its canonical id. See RequestOrganization.
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

// PutField sends (creates or updates) a field. See PutOrganization for baseRevision.
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

// BindFieldMapping binds a local identifier to an existing field. See BindOrganizationMapping.
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

// UnbindFieldMapping declares that this endpoint no longer holds a field. See UnbindOrganizationMapping.
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

// DeactivateField deactivates a field. See DeactivateOrganization.
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

// RequestField refetches a field by its canonical id. See RequestOrganization.
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

// PutFieldBoundary sends (creates or updates) a field boundary. See PutOrganization for baseRevision.
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

// BindFieldBoundaryMapping binds a local identifier to an existing field boundary. See BindOrganizationMapping.
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

// UnbindFieldBoundaryMapping declares that this endpoint no longer holds a field boundary. See UnbindOrganizationMapping.
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

// DeactivateFieldBoundary deactivates a field boundary. See DeactivateOrganization.
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

// RequestFieldBoundary refetches a field boundary by its canonical id. See RequestOrganization.
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
