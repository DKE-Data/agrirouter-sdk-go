package agmasync

import (
	"context"
	"fmt"
	"net/http"

	oapiclient "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi"
	oapi "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
)

// The initial-load states, in order. agrirouter starts the load and declares
// the set sent; the endpoint declares reconciliation done and the push
// finished.
const (
	// StateLoadingFromAgrirouter means the canonical set is owed to this
	// endpoint. agrirouter enters it when a user opts the endpoint into master
	// data or into a further entity type. A participant setting it gets
	// [ErrInitialLoadConflict].
	StateLoadingFromAgrirouter = oapi.LOADINGFROMAGRIROUTER

	// StateReconciling means agrirouter has sent the whole set. Only
	// agrirouter enters it, so this state, not the stream closing, proves the
	// set arrived.
	StateReconciling = oapi.RECONCILING

	// StateLoadingToAgrirouter is set by the endpoint to declare
	// reconciliation finished, carrying the bindings it produced. It cannot be
	// reached without having been in StateReconciling.
	StateLoadingToAgrirouter = oapi.LOADINGTOAGRIROUTER

	// StateCompleted is set by the endpoint once it has sent everything.
	// Steady-state synchronization applies from here on.
	StateCompleted = oapi.COMPLETED
)

// DeclareCapabilities builds the masterdata configuration declaring the given entity
// types, closed over entity dependencies (see [DependencyClosure]).
//
// The result goes in the `masterdata` field of the application's PutEndpoint
// call. This package does not make that call: PutEndpoint replaces the whole
// endpoint, so the application sends the declaration alongside its message
// capabilities and subscriptions.
//
// Withdrawing a type removes it from any selection naming it, as if the user
// had deselected it. A declaration never adds to what is delivered.
func DeclareCapabilities(types ...EntityType) oapi.MasterdataConfig {
	closure := DependencyClosure(types)
	capabilities := make([]oapi.EntityTypeToggle, 0, len(closure))
	for _, t := range closure {
		capabilities = append(capabilities, oapi.EntityTypeToggle{EntityType: string(t)})
	}
	return oapi.MasterdataConfig{Capabilities: capabilities}
}

// DependencyClosure expands a set of entity types to the dependency-closed set
// that a declaration or a selection carries.
//
// agrirouter rejects a declaration that is not closed. It closes the user's
// selection itself, so a participant uses this to predict what opting into a
// type brings with it.
func DependencyClosure(types []EntityType) []EntityType {
	want := make(map[EntityType]bool, len(EntityTypes))
	var add func(EntityType)
	add = func(t EntityType) {
		if want[t] {
			return
		}
		want[t] = true
		// A boundary's field is the only required reference; optional ones
		// constrain nothing.
		if t == TypeFieldBoundary {
			add(TypeField)
		}
	}
	for _, t := range types {
		add(t)
	}

	// Ordered by EntityTypes so the result is stable rather than map-ordered.
	out := make([]EntityType, 0, len(want))
	for _, t := range EntityTypes {
		if want[t] {
			out = append(out, t)
		}
	}
	return out
}

// SelectedTypes reads the entity types the user selected, off an
// [EventRouteChanged] frame.
//
// The result is in [DependencyOrder]. The frame states the whole selection, so
// an empty result means the endpoint exchanges nothing: the user deselected
// the last type or removed the route.
func SelectedTypes(s oapi.RouteChangedEventData) []EntityType {
	out := make([]EntityType, 0, len(s.EntityTypes))
	for _, typ := range DependencyOrder {
		for _, entry := range s.EntityTypes {
			if entry.EntityType == string(typ) {
				out = append(out, typ)
				break
			}
		}
	}
	return out
}

// DeclaresCapability reports whether the endpoint declared it can exchange a type.
//
// It reports what the software supports, not what the user opted it into; see
// [SelectedTypes] for that.
func DeclaresCapability(cfg oapi.MasterdataConfig, t EntityType) bool {
	for _, toggle := range cfg.Capabilities {
		if toggle.EntityType == string(t) {
			return true
		}
	}
	return false
}

// GetInitialLoadStatus reads the endpoint's initial-load state.
//
// The state, not the initial-load stream ending, tells whether the canonical
// set was sent: a dropped connection ends the stream just as completion does.
// An endpoint still in [StateLoadingFromAgrirouter] reconnects and takes the
// set from the beginning.
//
// Returns [ErrNotFound] for an endpoint opted into no entity type, which has no
// initial-load state at all.
func GetInitialLoadStatus(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	externalEndpointID string, tenantID uuid.UUID,
) (oapi.InitialLoadStatus, error) {
	r, err := api.GetInitialLoadStatusWithResponse(ctx, externalEndpointID,
		&oapi.GetInitialLoadStatusParams{XAgrirouterTenantId: tenantID})
	if err != nil {
		return oapi.InitialLoadStatus{}, transportErr(err)
	}
	if resErr := (writeResult{
		statusCode: r.StatusCode(), forbidden: r.JSON403, notFound: r.JSON404, body: r.Body,
	}).err(); resErr != nil {
		return oapi.InitialLoadStatus{}, resErr
	}
	if r.JSON200 == nil {
		return oapi.InitialLoadStatus{}, fmt.Errorf("%w: no initial load status", ErrEmptyResponse)
	}
	return *r.JSON200, nil
}

// IsRepeatLoad reports whether a canonical set now arriving is one this
// endpoint has been sent before.
//
// A participant that treats every arriving set as a first load creates local
// duplicates of data it already holds. The marker is previousLoadCompletedAt,
// which survives opt-out, disconnection, and endpoint removal along with the
// identifier mapping. A masterdata reset discards both, so the next load is a
// first load.
func IsRepeatLoad(s oapi.InitialLoadStatus) bool {
	return s.PreviousLoadCompletedAt != nil
}

// SetInitialLoadState drives one of the endpoint's two transitions.
//
// Out-of-order transitions are [ErrInitialLoadConflict]. Repeating the current
// state succeeds and reapplies any bindings carried, recomputing their
// rejections. The mapping cannot be read back, so this is how an endpoint whose
// confirmation went unanswered learns which bindings were recorded.
//
// [StateLoadingFromAgrirouter] is refused from every state, including itself.
// To report that a user is needed, use [ReportUserAttention].
func SetInitialLoadState(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	externalEndpointID string, tenantID uuid.UUID,
	upd oapi.InitialLoadStateUpdate,
) (oapi.InitialLoadStatus, error) {
	r, err := api.SetInitialLoadStateWithResponse(ctx, externalEndpointID,
		&oapi.SetInitialLoadStateParams{XAgrirouterTenantId: tenantID}, upd)
	if err != nil {
		return oapi.InitialLoadStatus{}, transportErr(err)
	}
	if r.StatusCode() == http.StatusConflict {
		msg := "initial load transition out of order"
		if r.JSON409 != nil && r.JSON409.Message != "" {
			msg = r.JSON409.Message
		}
		return oapi.InitialLoadStatus{}, &APIError{r.StatusCode(), msg, ErrInitialLoadConflict}
	}
	if resErr := (writeResult{
		statusCode: r.StatusCode(), validation: r.JSON400, forbidden: r.JSON403,
		notFound: r.JSON404, body: r.Body,
	}).err(); resErr != nil {
		return oapi.InitialLoadStatus{}, resErr
	}
	if r.JSON200 == nil {
		return oapi.InitialLoadStatus{}, fmt.Errorf("%w: no initial load status", ErrEmptyResponse)
	}
	return *r.JSON200, nil
}

// ConfirmReconciled declares that reconciliation is finished and carries the
// bindings it produced.
//
// It asserts that the participant has resolved the conflicts the set
// surfaced; receipt was already recorded by agrirouter moving the endpoint to
// [StateReconciling]. Reconciliation may wait on people, so staying in
// [StateReconciling] for a long time is not a fault.
//
// Bindings are applied independently: a pair that cannot be recorded comes back
// in the status's RejectedIdMappings instead of failing the transition. Resolve
// each and rebind it with [Bind]; [NeedsUser] tells which need a person.
func ConfirmReconciled(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	externalEndpointID string, tenantID uuid.UUID,
	bindings []oapi.IdMappingBinding,
) (oapi.InitialLoadStatus, error) {
	return SetInitialLoadState(ctx, api, externalEndpointID, tenantID, oapi.InitialLoadStateUpdate{
		State:      StateLoadingToAgrirouter,
		IdMappings: &bindings,
	})
}

// CompleteInitialLoad declares that the endpoint has sent everything it holds.
func CompleteInitialLoad(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	externalEndpointID string, tenantID uuid.UUID,
) (oapi.InitialLoadStatus, error) {
	return SetInitialLoadState(ctx, api, externalEndpointID, tenantID, oapi.InitialLoadStateUpdate{State: StateCompleted})
}

// ReportUserAttention tells agrirouter that this endpoint's reconciliation is
// waiting on a person.
func ReportUserAttention(
	ctx context.Context, api *oapiclient.ClientWithResponses,
	externalEndpointID string, tenantID uuid.UUID,
) (oapi.InitialLoadStatus, error) {
	r, err := api.ReportUserAttentionWithResponse(ctx, externalEndpointID,
		&oapi.ReportUserAttentionParams{XAgrirouterTenantId: tenantID})
	if err != nil {
		return oapi.InitialLoadStatus{}, transportErr(err)
	}
	if r.StatusCode() == http.StatusConflict {
		msg := "the initial load is completed and waits on nobody"
		if r.JSON409 != nil && r.JSON409.Message != "" {
			msg = r.JSON409.Message
		}
		return oapi.InitialLoadStatus{}, &APIError{r.StatusCode(), msg, ErrInitialLoadConflict}
	}
	if resErr := (writeResult{
		statusCode: r.StatusCode(), forbidden: r.JSON403, notFound: r.JSON404, body: r.Body,
	}).err(); resErr != nil {
		return oapi.InitialLoadStatus{}, resErr
	}
	if r.JSON200 == nil {
		return oapi.InitialLoadStatus{}, fmt.Errorf("%w: no initial load status", ErrEmptyResponse)
	}
	return *r.JSON200, nil
}
