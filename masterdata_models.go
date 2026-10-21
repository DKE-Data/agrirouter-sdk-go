package agrirouter

import (
	"slices"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/agmasync"
	internal_models "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
)

// Re-exporting master-data models for convenience, so users can access them
// directly from the agrirouter package, e.g. agrirouter.Party.

// Party is a master-data party: a person, an organization, or a party whose
// party type its sender does not record (no _agmasync_details). Its master
// data is in Content.
type Party = internal_models.Party

// PartyContent is a party's master data, based on the ADAPT Party.
type PartyContent = internal_models.PartyContent

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

// FarmContent is a farm's master data, based on the ADAPT Farm.
type FarmContent = internal_models.FarmContent

// Field is a master-data field, optionally belonging to a farm and held by a party.
type Field = internal_models.Field

// FieldContent is a field's master data, based on the ADAPT Field.
type FieldContent = internal_models.FieldContent

// FieldBoundary is a master-data boundary of the field it references.
type FieldBoundary = internal_models.FieldBoundary

// FieldBoundaryContent is a field boundary's master data, based on the ADAPT FieldBoundary.
type FieldBoundaryContent = internal_models.FieldBoundaryContent

// Id identifies an object or a reference's target: the canonical identifier
// and/or the near end's own local one in referenceId. See IdentityOf.
type Id = internal_models.Id //nolint:revive // Name follows the specification.

// UniqueId is one identifier in an Id.
type UniqueId = internal_models.UniqueId //nolint:revive // Name follows the specification.

// Ref references another master-data entity by its Id.
type Ref = internal_models.Ref

// ContactInfo is a party's contact data: addresses and telecommunication contacts.
type ContactInfo = internal_models.ContactInfo

// AddressContactMethod is an address of a party or a farm.
type AddressContactMethod = internal_models.AddressContactMethod

// TelecommunicationContactMethod is a phone number, fax number, or email address.
type TelecommunicationContactMethod = internal_models.TelecommunicationContactMethod

// Area is an area in hectares.
type Area = internal_models.Area

// Boundary is the geometry of a field boundary, as WKT, with how it was recorded.
type Boundary = internal_models.Boundary

// ObstacleBoundary is the geometry of an obstacle, as WKT.
type ObstacleBoundary = internal_models.ObstacleBoundary

// Membership links a person to an organization together with the role held there.
type Membership = internal_models.Membership

// Partner is a party holding a role on a farm, such as a contractor or advisor.
type Partner = internal_models.Partner

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

// Identity is what an Id says: the canonical identifier and the local one.
type Identity = agmasync.Identity

// AgrirouterIDSource is the idSource of the canonical identifier in an Id.
const AgrirouterIDSource = agmasync.AgrirouterIDSource

// IdentityOf reads an Id, an object's own content.id or a reference's.
func IdentityOf(id Id) (Identity, error) { return agmasync.IdentityOf(id) }

// LocalRef builds a reference from the application's own identifier for the
// target, which agrirouter resolves against its mapping. Send the target first.
func LocalRef(localID string) Ref { return agmasync.LocalRef(localID) }

// CanonicalRef builds a reference from the target's canonical identifier.
func CanonicalRef(agrirouterID uuid.UUID) Ref { return agmasync.AgrirouterRef(agrirouterID) }

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
