package agrirouter_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testTenantID = "8f2c6b9e-3a51-4d7e-9c1a-0b6f5e4d3c2a"

const testTenantEndpointJSON = `{
	"id": "1d3f7a2b-6c4e-4b8a-9f0d-2e5c7a1b3d4f",
	"name": "Tractor 1",
	"endpoint_type": "virtual_communication_unit",
	"application_id": "5a7b9c1d-3e5f-4a6b-8c9d-0e1f2a3b4c5d",
	"owned_by_your_application": true,
	"capabilities": {"can_send": [], "can_receive": []}
}`

// newTenantsTestClient starts a fake agrirouter API that answers path with body
// and returns a client pointed at it.
func newTenantsTestClient(t *testing.T, path, body string) *agrirouter.Client {
	t.Helper()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(server.Close)

	client, err := agrirouter.NewClient(server.URL)
	require.NoError(t, err)

	return client
}

func TestListAuthorizedTenantsExposesCompanyName(t *testing.T) {
	client := newTenantsTestClient(t, "/tenants", `{"tenants": [
		{"tenant_id": "`+testTenantID+`", "company_name": "Example Farm GmbH", "endpoints": []},
		{"tenant_id": "`+testTenantID+`", "endpoints": []}
	]}`)

	tenants, err := client.ListAuthorizedTenants(context.Background())
	require.NoError(t, err)
	require.Len(t, tenants, 2)

	require.NotNil(t, tenants[0].CompanyName)
	assert.Equal(t, "Example Farm GmbH", *tenants[0].CompanyName)
	assert.Nil(t, tenants[1].CompanyName, "company name must be absent when not returned")
}

func TestGetTenantEndpointsExposesCompanyName(t *testing.T) {
	tenantID := uuid.MustParse(testTenantID)
	client := newTenantsTestClient(t, "/tenants/"+testTenantID+"/endpoints", `{
		"company_name": "Example Farm GmbH",
		"endpoints": [`+testTenantEndpointJSON+`]
	}`)

	res, err := client.GetTenantEndpoints(context.Background(), tenantID)
	require.NoError(t, err)

	require.NotNil(t, res.CompanyName)
	assert.Equal(t, "Example Farm GmbH", *res.CompanyName)
	require.Len(t, res.Endpoints, 1)
	assert.Equal(t, "Tractor 1", res.Endpoints[0].Name)
}

func TestGetTenantEndpointsWithoutCompanyName(t *testing.T) {
	tenantID := uuid.MustParse(testTenantID)
	client := newTenantsTestClient(t, "/tenants/"+testTenantID+"/endpoints", `{"endpoints": []}`)

	res, err := client.GetTenantEndpoints(context.Background(), tenantID)
	require.NoError(t, err)

	assert.Nil(t, res.CompanyName, "company name must be absent when not returned")
	assert.Empty(t, res.Endpoints)
}

func TestListTenantEndpointsStillReturnsEndpoints(t *testing.T) {
	tenantID := uuid.MustParse(testTenantID)
	client := newTenantsTestClient(t, "/tenants/"+testTenantID+"/endpoints", `{
		"company_name": "Example Farm GmbH",
		"endpoints": [`+testTenantEndpointJSON+`]
	}`)

	endpoints, err := client.ListTenantEndpoints(context.Background(), tenantID)
	require.NoError(t, err)

	require.Len(t, endpoints, 1)
	assert.Equal(t, "Tractor 1", endpoints[0].Name)
}

func TestGetTenantEndpointsReturnsErrorOnFailedStatus(t *testing.T) {
	tenantID := uuid.MustParse(testTenantID)
	client := newTenantsTestClient(t, "/does-not-match", "")

	res, err := client.GetTenantEndpoints(context.Background(), tenantID)
	require.ErrorIs(t, err, agrirouter.ErrAPICallFailed)
	assert.Nil(t, res)
}
