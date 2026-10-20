package agmasync

import (
	"encoding/json"
	"fmt"

	oapi "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
)

// EntityType is one of the master-data entity types.
//
// The values match both the `type` discriminator on an entity and the
// `entityType` toggle, so a capability declaration or a routing selection
// compares against an entity's type directly.
//
// URL paths spell collections differently (`field-boundaries` for
// `fieldBoundary`). Nothing here converts between the two because the
// generated client has one operation per collection.
type EntityType string

// The entity types.
const (
	TypeParty         EntityType = "party"
	TypeFarm          EntityType = "farm"
	TypeField         EntityType = "field"
	TypeFieldBoundary EntityType = "fieldBoundary"
)

// EntityTypes lists every supported type.
var EntityTypes = []EntityType{
	TypeParty, TypeFarm, TypeField, TypeFieldBoundary,
}

// DependencyOrder lists the types so that a referenced type precedes the types
// that reference it. A participant sends in this order, and agrirouter
// delivers in it.
//
// agrirouter resolves a reference against the sender's identifier mapping, so
// a field sent before the farm it names has nothing to resolve to. Sending in
// this order binds every reference target before it is used.
//
// Within parties the order is per object, because a membership references one
// party from another; see [PartyTier].
var DependencyOrder = []EntityType{
	TypeParty, TypeFarm, TypeField, TypeFieldBoundary,
}

// Valid reports whether the type is one this version of the protocol defines.
func (t EntityType) Valid() bool {
	switch t {
	case TypeParty, TypeFarm, TypeField, TypeFieldBoundary:
		return true
	default:
		return false
	}
}

// Envelope is the set of fields every entity shares, plus the type
// discriminator and the object's identity.
//
// Every field but Type is a pointer because agrirouter assigns it, so it is
// absent on an object not yet sent. AgrirouterID and LocalID are not envelope
// fields on the wire but the two entries of `content.id` (see [IdentityOf]). A
// delivered object without LocalID has no mapping for the receiving
// application: the receiver must create it locally and bind it.
type Envelope struct {
	oapi.Envelope
	Type EntityType `json:"type"`

	AgrirouterID *uuid.UUID `json:"-"`
	LocalID      *string    `json:"-"`
}

// EnvelopeOf reads the common fields of an entity of any type from its JSON.
//
// The event stream carries all types over one connection, so a receiver reads
// the envelope before it knows which schema to decode into.
func EnvelopeOf(raw []byte) (Envelope, error) {
	env, err := readEnvelope(raw)
	if err != nil {
		return env, err
	}
	if !env.Type.Valid() {
		return env, fmt.Errorf("agmasync: %w: %q", ErrUnknownEntityType, env.Type)
	}
	return env, nil
}

// readEnvelope reads the envelope and the identity in `content.id`, without
// checking the type.
func readEnvelope(raw []byte) (Envelope, error) {
	var probe struct {
		Envelope
		Content struct {
			ID oapi.Id `json:"id"`
		} `json:"content"`
	}
	if err := json.Unmarshal(raw, &probe); err != nil {
		return Envelope{}, fmt.Errorf("agmasync: reading entity envelope: %w", err)
	}
	env := probe.Envelope
	id, err := IdentityOf(probe.Content.ID)
	if err != nil {
		return Envelope{}, err
	}
	env.AgrirouterID, env.LocalID = id.AgrirouterID, id.LocalID
	return env, nil
}

// Object is a canonical object of any entity type, decoded into its model.
type Object struct {
	// Envelope holds the object's common fields. Its Type says which of Party,
	// Farm, Field, and FieldBoundary is set.
	Envelope Envelope

	// The object, decoded into its model. Exactly one is set, the one
	// Envelope.Type names.
	Party         *oapi.Party
	Farm          *oapi.Farm
	Field         *oapi.Field
	FieldBoundary *oapi.FieldBoundary
}

// ObjectOf decodes a canonical object of any entity type from its JSON.
func ObjectOf(raw []byte) (Object, error) {
	env, err := EnvelopeOf(raw)
	if err != nil {
		return Object{}, err
	}
	return decodeObject(env, raw)
}

// objectAs decodes the answer to an operation that names its entity type, as
// an object of type t. The type travels on the request, so the body need not
// repeat it; when it does, it must agree.
func objectAs(t EntityType, raw []byte) (Object, error) {
	env, err := readEnvelope(raw)
	if err != nil {
		return Object{}, err
	}
	switch env.Type {
	case "":
		env.Type = t
	case t:
	default:
		return Object{}, fmt.Errorf("agmasync: %w: asked for a %s, answered with a %s",
			ErrEntityTypeMismatch, t, env.Type)
	}
	return decodeObject(env, raw)
}

// decodeObject decodes raw into the model env.Type names.
func decodeObject(env Envelope, raw []byte) (Object, error) {
	var err error
	o := Object{Envelope: env}
	switch env.Type {
	case TypeParty:
		o.Party = new(oapi.Party)
		err = json.Unmarshal(raw, o.Party)
	case TypeFarm:
		o.Farm = new(oapi.Farm)
		err = json.Unmarshal(raw, o.Farm)
	case TypeField:
		o.Field = new(oapi.Field)
		err = json.Unmarshal(raw, o.Field)
	case TypeFieldBoundary:
		o.FieldBoundary = new(oapi.FieldBoundary)
		err = json.Unmarshal(raw, o.FieldBoundary)
	}
	if err != nil {
		return Object{}, fmt.Errorf("agmasync: decoding %s: %w", env.Type, err)
	}
	return o, nil
}

// The party types `_agmasync_details` states. A party without them is of
// unknown party type.
const (
	PartyTypePerson       = "PERSON"
	PartyTypeOrganization = "ORGANIZATION"
)

// PartyTypeOf reads the party type a party states: [PartyTypePerson],
// [PartyTypeOrganization], or "" where it carries no `_agmasync_details`.
func PartyTypeOf(p oapi.Party) (string, error) {
	details := p.Content.UnderscoreAgmasyncDetails
	if !details.IsSpecified() || details.IsNull() {
		return "", nil
	}
	d, err := details.Get()
	if err != nil {
		return "", fmt.Errorf("agmasync: reading party details: %w", err)
	}
	t, err := d.Discriminator()
	if err != nil {
		return "", fmt.Errorf("agmasync: reading party details: %w", err)
	}
	return t, nil
}

// PartyTier is the position of a party in send and delivery order: 1 for a
// party with person `_agmasync_details`, which may name organizations through its
// memberships, and 0 for any other, which references nothing.
//
// Sending tier 0 before tier 1 binds every membership target before a person
// names it.
func PartyTier(p oapi.Party) (int, error) {
	t, err := PartyTypeOf(p)
	if err != nil {
		return 0, err
	}
	if t == PartyTypePerson {
		return 1, nil
	}
	return 0, nil
}
