package agmasync_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/agmasync"
	oapiclient "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi"
	"github.com/google/uuid"
)

// TestEnvelopeOfReadsTheRecipientEndpoint pins that a delivered frame names the
// endpoint its localIds belong to: the mapping is scoped to the endpoint.
func TestEnvelopeOfReadsTheRecipientEndpoint(t *testing.T) {
	recipient := uuid.MustParse("2b3c4d5e-6f70-8192-a3b4-c5d6e7f80912")
	env, err := agmasync.EnvelopeOf([]byte(`{"type":"farm","recipient_endpoint_id":"` + recipient.String() +
		`","content":{"id":{"referenceId":"abc"},"name":"Hof"}}`))
	if err != nil {
		t.Fatalf("EnvelopeOf: %v", err)
	}
	if env.RecipientEndpointId == nil || *env.RecipientEndpointId != recipient {
		t.Errorf("RecipientEndpointId = %v, want %s", env.RecipientEndpointId, recipient)
	}
}

// TestBaseRevisionRequiredCarriesTheCurrentRevision pins that a 428 hands back
// the revision its body names, so a participant whose create landed but whose
// response was lost can rebase its resend.
func TestBaseRevisionRequiredCarriesTheCurrentRevision(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusPreconditionRequired)
		_, _ = w.Write([]byte(`{"message":"base revision required","current_revision":4}`))
	}))
	t.Cleanup(srv.Close)
	api, err := oapiclient.NewClientWithResponses(srv.URL)
	if err != nil {
		t.Fatal(err)
	}

	_, err = agmasync.Deactivate(context.Background(), api, uuid.New(), uuid.New(), agmasync.TypeFarm, "farm-1", nil)
	var conflict *agmasync.RevisionConflict
	if !errors.As(err, &conflict) || !errors.Is(err, agmasync.ErrBaseRevisionRequired) {
		t.Fatalf("error = %v, want a RevisionConflict matching ErrBaseRevisionRequired", err)
	}
	if conflict.CurrentRevision != 4 {
		t.Errorf("CurrentRevision = %d, want 4", conflict.CurrentRevision)
	}
	if errors.Is(err, agmasync.ErrRevisionConflict) {
		t.Errorf("a 428 matches ErrRevisionConflict")
	}
}
