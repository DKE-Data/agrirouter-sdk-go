package agmasync

import (
	"fmt"

	oapi "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
)

// The agrirouter entry of an `Id`: the canonical identifier, a UUID from this
// reserved source.
const (
	AgrirouterIDSource = "https://agrirouter.com"
	IDTypeUUID         = "UUID"
)

// IDTypeString is the `idTypeCode` agrirouter delivers on a local entry when
// the application sent none.
const IDTypeString = "STRING"

// Identity is what an `Id` says: the canonical identifier and the near end's
// local one. Either may be absent.
type Identity struct {
	AgrirouterID *uuid.UUID
	LocalID      *string
}

// IdentityOf reads an `Id`, an object's own or a reference's. The local
// identifier is `referenceId`; the canonical one is the `uniqueIds` entry
// whose `idSource` is [AgrirouterIDSource]. Other entries are ignored.
func IdentityOf(id oapi.Id) (Identity, error) {
	var out Identity
	if id.ReferenceId != nil {
		local := *id.ReferenceId
		out.LocalID = &local
	}
	if id.UniqueIds == nil {
		return out, nil
	}
	for _, entry := range *id.UniqueIds {
		if entry.IdSource == nil || *entry.IdSource != AgrirouterIDSource {
			continue
		}
		parsed, err := uuid.Parse(entry.IdText)
		if err != nil {
			return Identity{}, fmt.Errorf("agmasync: reading agrirouter identifier %q: %w", entry.IdText, err)
		}
		out.AgrirouterID = &parsed
		break
	}
	return out, nil
}

// LocalID builds an `Id` from the near end's own identifier alone, which is
// all a send needs. It is what goes in `content.id` of an object written.
func LocalID(localID string) oapi.Id {
	return oapi.Id{ReferenceId: &localID}
}

// LocalRef builds a reference to another entity from the referencing
// endpoint's own identifier for the target.
//
// A reference may carry either identifier on send. Using the local one lets a
// participant build references from its own keys without tracking canonical
// ones. agrirouter resolves it against the sending application's mapping and
// rejects the write if the target has not been sent yet, so send the target
// first.
func LocalRef(localID string) oapi.Ref {
	return oapi.Ref{Id: LocalID(localID)}
}

// AgrirouterRef builds a reference to another entity from its canonical
// identifier, for a target the participant holds no local identifier for.
func AgrirouterRef(agrirouterID uuid.UUID) oapi.Ref {
	source := AgrirouterIDSource
	return oapi.Ref{Id: oapi.Id{UniqueIds: &[]oapi.UniqueId{{
		IdText: agrirouterID.String(), IdTypeCode: IDTypeUUID, IdSource: &source,
	}}}}
}
