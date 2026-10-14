package agmasync

import (
	"context"
	"fmt"

	oapiclient "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi"
	oapi "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
)

// Deactivate signals that an entity was deactivated in its source system.
//
// Deactivation covers archival, deletion, or any state in which the source no
// longer considers the entity active. The object and its identifier mapping
// are retained, so references and synchronization stay intact.
//
// Deactivating an active entity is an ordinary write under concurrency
// control: agrirouter rejects it if base cannot be reconciled with the
// current revision. Deactivating an inactive entity ignores base, creates no
// revision, and forwards nothing, so retrying a call with an unknown outcome
// is safe.
//
// To reactivate, Put the entity with active set to true.
//
// The result is the canonical object as it now stands.
func Deactivate(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	endpointID, tenantID uuid.UUID,
	t EntityType, localID string, base *int,
) (Object, error) {
	var res writeResult
	switch t {
	case TypeParty:
		r, err := api.DeactivatePartyWithResponse(ctx, localID, &oapi.DeactivatePartyParams{
			XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID, XAgrirouterBaseRevision: base,
		})
		if err != nil {
			return Object{}, transportErr(err)
		}
		res = writeResult{
			statusCode: r.StatusCode(), forbidden: r.JSON403, notFound: r.JSON404,
			precond: r.JSON412, required: r.JSON428, body: r.Body,
		}
	case TypeFarm:
		r, err := api.DeactivateFarmWithResponse(ctx, localID, &oapi.DeactivateFarmParams{
			XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID, XAgrirouterBaseRevision: base,
		})
		if err != nil {
			return Object{}, transportErr(err)
		}
		res = writeResult{
			statusCode: r.StatusCode(), forbidden: r.JSON403, notFound: r.JSON404,
			precond: r.JSON412, required: r.JSON428, body: r.Body,
		}
	case TypeField:
		r, err := api.DeactivateFieldWithResponse(ctx, localID, &oapi.DeactivateFieldParams{
			XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID, XAgrirouterBaseRevision: base,
		})
		if err != nil {
			return Object{}, transportErr(err)
		}
		res = writeResult{
			statusCode: r.StatusCode(), forbidden: r.JSON403, notFound: r.JSON404,
			precond: r.JSON412, required: r.JSON428, body: r.Body,
		}
	case TypeFieldBoundary:
		r, err := api.DeactivateFieldBoundaryWithResponse(ctx, localID, &oapi.DeactivateFieldBoundaryParams{
			XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID, XAgrirouterBaseRevision: base,
		})
		if err != nil {
			return Object{}, transportErr(err)
		}
		res = writeResult{
			statusCode: r.StatusCode(), forbidden: r.JSON403, notFound: r.JSON404,
			precond: r.JSON412, required: r.JSON428, body: r.Body,
		}
	default:
		return Object{}, fmt.Errorf("agmasync: %w: %q", ErrUnknownEntityType, t)
	}
	if err := res.err(); err != nil {
		return Object{}, err
	}
	return objectAs(t, res.body)
}
