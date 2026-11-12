//nolint:dupl // Archive and Unarchive are mirror images, each kept as one readable switch per entity type.
package agmasync

import (
	"context"
	"fmt"

	oapiclient "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi"
	oapi "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
)

// Archive signals that an entity was archived in its source system.
//
// Archiving covers a record archived or deleted in the source, or any other
// state in which the source no longer keeps the entity in use. The object and
// its identifier mapping are retained, so references and synchronization stay
// intact.
//
// Archiving an entity that is not archived is an ordinary write under
// concurrency control: agrirouter rejects it if base cannot be reconciled with
// the current revision. Archiving an archived entity ignores base, creates no
// revision, and forwards nothing, so retrying a call with an unknown outcome
// is safe.
//
// [Unarchive] is the way back. A Put never changes the state: it must carry
// the object's current `archived`.
//
// The result is the canonical object as it now stands.
func Archive(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	endpointID, tenantID uuid.UUID,
	t EntityType, localID string, base *int,
) (Object, error) {
	var res writeResult
	switch t {
	case TypeParty:
		r, err := api.ArchivePartyWithResponse(ctx, localID, &oapi.ArchivePartyParams{
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
		r, err := api.ArchiveFarmWithResponse(ctx, localID, &oapi.ArchiveFarmParams{
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
		r, err := api.ArchiveFieldWithResponse(ctx, localID, &oapi.ArchiveFieldParams{
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
		r, err := api.ArchiveFieldBoundaryWithResponse(ctx, localID, &oapi.ArchiveFieldBoundaryParams{
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
