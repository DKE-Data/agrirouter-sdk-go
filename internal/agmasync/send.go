package agmasync

import (
	"context"
	"fmt"

	oapiclient "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi"
	oapi "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
)

// PutParty sends a party. PutFarm, PutField, and PutFieldBoundary likewise
// send an entity (a creation or an update).
//
// base is the revision the participant edited from, sent in the
// x-agrirouter-base-revision header. It is nil only for a create; an update
// without it fails with [ErrBaseRevisionRequired].
//
// The entity is a merge patch: an omitted attribute is left unchanged, and one
// sent as null is removed (the models' nullable fields allow this). It must
// carry its local_id as `content.id.referenceId` ([LocalID] builds one); its
// type is set from the model. Required attributes are
// required on every write.
//
// The returned model is the resulting canonical object; apply it as you would
// a delivered one. agrirouter does not echo the write on the sender's own
// stream, so this response is the only place the sender learns the
// agrirouterId of a new object, or the merged result where agrirouter
// reconciled the write with a concurrent change.
//
//nolint:dupl // One readable function per entity type is preferred over a generic helper.
func PutParty(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	endpointID, tenantID uuid.UUID,
	v oapi.Party, base *int,
) (oapi.Party, error) {
	localID := v.Content.Id.ReferenceId
	if localID == nil || *localID == "" {
		return oapi.Party{}, localIDRequired(TypeParty)
	}
	v.Type = string(TypeParty)
	r, err := api.PutPartyWithResponse(ctx, *localID, &oapi.PutPartyParams{
		XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID, XAgrirouterBaseRevision: base,
	}, v)
	if err != nil {
		return oapi.Party{}, transportErr(err)
	}
	if err := (writeResult{
		statusCode: r.StatusCode(), validation: r.JSON400, forbidden: r.JSON403,
		conflict: r.JSON409, precond: r.JSON412, required: r.JSON428, body: r.Body,
	}).err(); err != nil {
		return oapi.Party{}, err
	}
	if r.JSON200 != nil {
		return *r.JSON200, nil
	}
	if r.JSON201 != nil {
		return *r.JSON201, nil
	}
	return oapi.Party{}, emptyWriteResponse(TypeParty, r.StatusCode())
}

// PutFarm sends a farm. See [PutParty].
//
//nolint:dupl // One readable function per entity type is preferred over a generic helper.
func PutFarm(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	endpointID, tenantID uuid.UUID,
	v oapi.Farm, base *int,
) (oapi.Farm, error) {
	localID := v.Content.Id.ReferenceId
	if localID == nil || *localID == "" {
		return oapi.Farm{}, localIDRequired(TypeFarm)
	}
	v.Type = string(TypeFarm)
	r, err := api.PutFarmWithResponse(ctx, *localID, &oapi.PutFarmParams{
		XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID, XAgrirouterBaseRevision: base,
	}, v)
	if err != nil {
		return oapi.Farm{}, transportErr(err)
	}
	if err := (writeResult{
		statusCode: r.StatusCode(), validation: r.JSON400, forbidden: r.JSON403,
		conflict: r.JSON409, precond: r.JSON412, required: r.JSON428, body: r.Body,
	}).err(); err != nil {
		return oapi.Farm{}, err
	}
	if r.JSON200 != nil {
		return *r.JSON200, nil
	}
	if r.JSON201 != nil {
		return *r.JSON201, nil
	}
	return oapi.Farm{}, emptyWriteResponse(TypeFarm, r.StatusCode())
}

// PutField sends a field. See [PutParty].
//
//nolint:dupl // One readable function per entity type is preferred over a generic helper.
func PutField(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	endpointID, tenantID uuid.UUID,
	v oapi.Field, base *int,
) (oapi.Field, error) {
	localID := v.Content.Id.ReferenceId
	if localID == nil || *localID == "" {
		return oapi.Field{}, localIDRequired(TypeField)
	}
	v.Type = string(TypeField)
	r, err := api.PutFieldWithResponse(ctx, *localID, &oapi.PutFieldParams{
		XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID, XAgrirouterBaseRevision: base,
	}, v)
	if err != nil {
		return oapi.Field{}, transportErr(err)
	}
	if err := (writeResult{
		statusCode: r.StatusCode(), validation: r.JSON400, forbidden: r.JSON403,
		conflict: r.JSON409, precond: r.JSON412, required: r.JSON428, body: r.Body,
	}).err(); err != nil {
		return oapi.Field{}, err
	}
	if r.JSON200 != nil {
		return *r.JSON200, nil
	}
	if r.JSON201 != nil {
		return *r.JSON201, nil
	}
	return oapi.Field{}, emptyWriteResponse(TypeField, r.StatusCode())
}

// PutFieldBoundary sends a field boundary. See [PutParty].
//
//nolint:dupl // One readable function per entity type is preferred over a generic helper.
func PutFieldBoundary(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	endpointID, tenantID uuid.UUID,
	v oapi.FieldBoundary, base *int,
) (oapi.FieldBoundary, error) {
	localID := v.Content.Id.ReferenceId
	if localID == nil || *localID == "" {
		return oapi.FieldBoundary{}, localIDRequired(TypeFieldBoundary)
	}
	v.Type = string(TypeFieldBoundary)
	r, err := api.PutFieldBoundaryWithResponse(ctx, *localID, &oapi.PutFieldBoundaryParams{
		XAgrirouterEndpointId: endpointID, XAgrirouterTenantId: tenantID, XAgrirouterBaseRevision: base,
	}, v)
	if err != nil {
		return oapi.FieldBoundary{}, transportErr(err)
	}
	if err := (writeResult{
		statusCode: r.StatusCode(), validation: r.JSON400, forbidden: r.JSON403,
		conflict: r.JSON409, precond: r.JSON412, required: r.JSON428, body: r.Body,
	}).err(); err != nil {
		return oapi.FieldBoundary{}, err
	}
	if r.JSON200 != nil {
		return *r.JSON200, nil
	}
	if r.JSON201 != nil {
		return *r.JSON201, nil
	}
	return oapi.FieldBoundary{}, emptyWriteResponse(TypeFieldBoundary, r.StatusCode())
}

func localIDRequired(t EntityType) error {
	return fmt.Errorf("%w: sending a %s", ErrLocalIDRequired, t)
}

// emptyWriteResponse is a successful write that carried no object to apply.
func emptyWriteResponse(t EntityType, statusCode int) error {
	return fmt.Errorf("%w: HTTP %d carried no %s", ErrEmptyResponse, statusCode, t)
}
