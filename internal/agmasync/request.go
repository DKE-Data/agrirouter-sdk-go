package agmasync

import (
	"context"
	"fmt"

	oapiclient "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi"
	oapi "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
)

// Request asks for a single entity by its canonical identifier.
//
// agrirouter answers 202 with no body and delivers the object on the
// application's event stream, even when this endpoint was its last writer.
//
// A request never widens what a participant can see. It recovers an object
// lost locally, or one referenced by a live-stream object that the
// initial-load stream has not reached yet. A participant that has lost too
// many objects to name asks for the whole canonical set instead.
//
// It also unblocks an object whose reference target the participant does not
// hold, such as a farm delivered with an owner reference that has no localId.
// That object cannot be sent until the target is requested, created locally,
// and bound with [Bind].
func Request(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	endpointID, tenantID uuid.UUID,
	t EntityType, agrirouterID uuid.UUID,
) error {
	body := oapi.EntityRequest{AgrirouterId: agrirouterID}

	switch t {
	case TypeParty:
		r, err := api.RequestPartyWithResponse(ctx,
			&oapi.RequestPartyParams{XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID}, body)
		if err != nil {
			return transportErr(err)
		}
		return writeResult{
			statusCode: r.StatusCode(), forbidden: r.JSON403,
			notFound: r.JSON404, body: r.Body,
		}.err()

	case TypeFarm:
		r, err := api.RequestFarmWithResponse(ctx,
			&oapi.RequestFarmParams{XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID}, body)
		if err != nil {
			return transportErr(err)
		}
		return writeResult{
			statusCode: r.StatusCode(), forbidden: r.JSON403,
			notFound: r.JSON404, body: r.Body,
		}.err()

	case TypeField:
		r, err := api.RequestFieldWithResponse(ctx,
			&oapi.RequestFieldParams{XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID}, body)
		if err != nil {
			return transportErr(err)
		}
		return writeResult{
			statusCode: r.StatusCode(), forbidden: r.JSON403,
			notFound: r.JSON404, body: r.Body,
		}.err()

	case TypeFieldBoundary:
		r, err := api.RequestFieldBoundaryWithResponse(ctx,
			&oapi.RequestFieldBoundaryParams{XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID}, body)
		if err != nil {
			return transportErr(err)
		}
		return writeResult{
			statusCode: r.StatusCode(), forbidden: r.JSON403,
			notFound: r.JSON404, body: r.Body,
		}.err()

	default:
		return fmt.Errorf("agmasync: %w: %q", ErrUnknownEntityType, t)
	}
}
