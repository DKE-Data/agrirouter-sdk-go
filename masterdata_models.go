package agrirouter

import (
	internal_models "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
)

// Re-exporting master-data models for convenience, so users can access them
// directly from the agrirouter package, e.g. agrirouter.Organization.

// Organization is a master-data organization (a company or similar legal party).
type Organization = internal_models.Organization

// Person is a master-data person, optionally a member of one or more organizations.
type Person = internal_models.Person

// Farm is a master-data farm, owned by a party and optionally worked by partners.
type Farm = internal_models.Farm

// Field is a master-data field belonging to a farm.
type Field = internal_models.Field

// FieldBoundary is a master-data boundary of a field, carrying a GeoJSON geometry.
type FieldBoundary = internal_models.FieldBoundary

// Address is a postal address used by organizations and persons.
type Address = internal_models.Address

// Contact is contact information (phone, email, ...) of an organization or person.
type Contact = internal_models.Contact

// Membership links a person to an organization together with the role held there.
type Membership = internal_models.Membership

// Partner is a party holding a role on a farm, such as a contractor or advisor.
type Partner = internal_models.Partner

// PartyReference references a party (organization or person) by canonical and/or local id.
type PartyReference = internal_models.PartyReference

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

// MasterdataCapabilities is an endpoint's per-entity-type opt-in for master-data
// exchange, carried on PutEndpointRequest. Absence of a toggle for an entity type
// means the endpoint is not opted in for it.
type MasterdataCapabilities = internal_models.MasterdataCapabilities

// EntityTypeToggle opts an endpoint into master-data exchange for one entity type
// (e.g. organizations, persons, farms, fields, field-boundaries).
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
