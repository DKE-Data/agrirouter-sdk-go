package agrirouter

import (
	"slices"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/agmasync"
	internal_models "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
)

// Re-exporting master-data models for convenience, so users can access them
// directly from the agrirouter package, e.g. agrirouter.Party.

// Party is a master-data party: a person, an organization, or a party whose
// party type its sender does not record (no Details).
type Party = internal_models.Party

// PartyDetails states a party's party type, with the attributes specific to it.
type PartyDetails = internal_models.PartyDetails

// PersonDetails are the details of a party that is a natural person.
type PersonDetails = internal_models.PersonDetails

// OrganizationDetails are the details of a party that is an organization.
type OrganizationDetails = internal_models.OrganizationDetails

// RouteChangedEventData is the payload of a MasterdataEventRouteChanged frame:
// the entity types the user has selected on one endpoint.
type RouteChangedEventData = internal_models.RouteChangedEventData

// MasterdataResetEventData is the payload of a MasterdataEventReset frame.
type MasterdataResetEventData = internal_models.MasterdataResetEventData

// Farm is a master-data farm, optionally held by a party and worked by partners.
type Farm = internal_models.Farm

// Field is a master-data field, optionally belonging to a farm and held by a party.
type Field = internal_models.Field

// FieldBoundary is a master-data boundary of the field it references, carrying a GeoJSON geometry.
type FieldBoundary = internal_models.FieldBoundary

// Address is a postal address used by parties.
type Address = internal_models.Address

// Contact is contact information (phone, email, ...) of a party.
type Contact = internal_models.Contact

// Membership links a person to an organization together with the role held there.
type Membership = internal_models.Membership

// Partner is a party holding a role on a farm, such as a contractor or advisor.
type Partner = internal_models.Partner

// EntityReference references another master-data entity by canonical and/or local id.
type EntityReference = internal_models.EntityReference

// Geometry is a GeoJSON geometry (RFC 7946) with positions as [longitude, latitude(, altitude)].
type Geometry = internal_models.Geometry

// HarvestPeriod is a harvest period expressed as an interval.
type HarvestPeriod = internal_models.HarvestPeriod

// SoilInfo describes the soil characteristics of a field.
type SoilInfo = internal_models.SoilInfo

// Obstacle is an obstacle located within a field boundary.
type Obstacle = internal_models.Obstacle

// EntityRequest is the body used to lazy-load a master-data entity by its canonical id.
type EntityRequest = internal_models.EntityRequest

// MasterdataConfig declares the entity types an endpoint can exchange, carried on
// PutEndpointRequest. It must be dependency-closed: field boundaries require
// fields.
type MasterdataConfig = internal_models.MasterdataConfig

// EntityTypeToggle opts an endpoint into master-data exchange for one entity type
// (party, farm, field, fieldBoundary).
type EntityTypeToggle = internal_models.EntityTypeToggle

// InitialLoadStatus is an endpoint's initial-load state across every entity type it is opted into.
type InitialLoadStatus = internal_models.InitialLoadStatus

// InitialLoadStateUpdate advances an endpoint's initial-load state and may carry reconciliation bindings.
type InitialLoadStateUpdate = internal_models.InitialLoadStateUpdate

// InitialLoadState is the per-endpoint initial-load state value.
type InitialLoadState = internal_models.InitialLoadState

// IDMappingBinding binds a canonical object to a local identifier during reconciliation.
type IDMappingBinding = internal_models.IdMappingBinding

// IDMappingRejection reports a binding that could not be recorded, with its reason.
type IDMappingRejection = internal_models.IdMappingRejection

// MasterdataError is the error body returned by master-data operations.
type MasterdataError = internal_models.MasterdataError

// RevisionConflictError is returned on a base-revision precondition failure; it carries the current revision.
type RevisionConflictError = internal_models.RevisionConflictError

// MappingConflictError is returned when a local identifier or canonical object is already mapped.
type MappingConflictError = internal_models.MappingConflictError

// Initial-load states, in the order they are normally reached.
const (
	// InitialLoadStateLoadingFromAgrirouter means agrirouter is still sending the canonical set.
	InitialLoadStateLoadingFromAgrirouter = internal_models.LOADINGFROMAGRIROUTER
	// InitialLoadStateReconciling means the set has been delivered and the endpoint is reconciling conflicts.
	InitialLoadStateReconciling = internal_models.RECONCILING
	// InitialLoadStateLoadingToAgrirouter means the endpoint is sending objects the canonical set lacked.
	InitialLoadStateLoadingToAgrirouter = internal_models.LOADINGTOAGRIROUTER
	// InitialLoadStateCompleted means steady-state synchronization applies.
	InitialLoadStateCompleted = internal_models.COMPLETED
)

// EntityType names a kind of master-data object.
type EntityType = agmasync.EntityType

// The master-data entity types.
const (
	EntityTypeParty         = agmasync.TypeParty
	EntityTypeFarm          = agmasync.TypeFarm
	EntityTypeField         = agmasync.TypeField
	EntityTypeFieldBoundary = agmasync.TypeFieldBoundary
)

// Master-data stream event types
const (
	MasterdataEventChanged      = agmasync.EventMasterdataChanged
	MasterdataEventDeactivated  = agmasync.EventMasterdataDeactivated
	MasterdataEventCaughtUp     = agmasync.EventCaughtUp
	MasterdataEventRouteChanged = agmasync.EventRouteChanged
	MasterdataEventReset        = agmasync.EventMasterdataReset
)

// DeclareCapabilities builds the MasterdataConfig an endpoint declares on
// PutEndpoint, closed over entity dependencies.
func DeclareCapabilities(types ...EntityType) MasterdataConfig {
	return agmasync.DeclareCapabilities(types...)
}

// SelectedTypes reads the entity types a MasterdataEventRouteChanged frame says
// the user has selected, in dependency order.
func SelectedTypes(s RouteChangedEventData) []EntityType { return agmasync.SelectedTypes(s) }

// DependencyOrder lists every entity type so that a referenced type precedes the
// types that reference it: the order a participant sends in. The result is a
// copy the caller may modify.
func DependencyOrder() []EntityType { return slices.Clone(agmasync.DependencyOrder) }

// IsRepeatLoad reports whether the canonical set now arriving is one this
// endpoint has been sent before. A participant must not infer a first
// connection from the arrival of a set: it would create local duplicates of data
// it already holds. The marker survives an opt-out, a disconnection, and endpoint
// removal; a masterdata reset discards it, so the load after one is a first load.
func IsRepeatLoad(s InitialLoadStatus) bool { return agmasync.IsRepeatLoad(s) }
