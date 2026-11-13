package agrirouter_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tenantStateTestTenantID = "3c9e5a71-0b2d-4f6e-8a1c-7d4b2e9f0a63"

// newTenantStateTestClient starts a fake agrirouter API that answers path with
// body using contentType and returns a client pointed at it.
func newTenantStateTestClient(t *testing.T, path, contentType, body string) *agrirouter.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", contentType)
		_, _ = w.Write([]byte(body))
		if flusher, ok := w.(http.Flusher); ok {
			flusher.Flush()
		}
		if contentType == "text/event-stream" {
			// keep the stream open until the client disconnects
			<-r.Context().Done()
		}
	}))
	t.Cleanup(server.Close)

	client, err := agrirouter.NewClient(server.URL)
	require.NoError(t, err)

	return client
}

func TestListAuthorizedTenantsExposesState(t *testing.T) {
	client := newTenantStateTestClient(t, "/tenants", "application/json", `{"tenants": [
		{"tenant_id": "`+tenantStateTestTenantID+`", "state": "c2b7e1f0-customer-4711", "endpoints": []},
		{"tenant_id": "`+tenantStateTestTenantID+`", "endpoints": []}
	]}`)

	tenants, err := client.ListAuthorizedTenants(context.Background())
	require.NoError(t, err)
	require.Len(t, tenants, 2)

	require.NotNil(t, tenants[0].State)
	assert.Equal(t, "c2b7e1f0-customer-4711", *tenants[0].State)
	assert.Nil(t, tenants[1].State, "state must be absent when not returned")
}

func TestAuthorizationAddedEventExposesTenantState(t *testing.T) {
	eventJSON := `{"event_type": "AUTHORIZATION_ADDED", "scope": "endpoints:manage", ` +
		`"state": "c2b7e1f0-customer-4711", "tenant": {"tenant_id": "` + tenantStateTestTenantID +
		`", "state": "c2b7e1f0-customer-4711", "endpoints": []}}`
	client := newTenantStateTestClient(t, "/events", "text/event-stream", fmt.Sprintf("data: %s\n\n", eventJSON))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var received *agrirouter.AuthorizationAddedEventData
	_ = client.ReceiveAuthorizationAddedEvents(ctx, func(_ context.Context, event *agrirouter.AuthorizationAddedEventData) {
		received = event
		cancel()
	}, func(err error) {
		t.Errorf("unexpected event error: %v", err)
	})

	require.NotNil(t, received, "no AUTHORIZATION_ADDED event received")
	require.NotNil(t, received.Tenant.State)
	assert.Equal(t, "c2b7e1f0-customer-4711", *received.Tenant.State)
}
