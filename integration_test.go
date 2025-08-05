package agrirouter_test

import (
	"context"
	"io"
	"log"
	"os"
	"testing"
	"time"

	"github.com/DKE-Data/agrirouter-go-sdk/internal/tests/agriroutertestcontainer"
	"github.com/DKE-Data/agrirouter-go-sdk/pkg/agrirouter"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var testContainer *agriroutertestcontainer.AgrirouterContainer

func TestMain(m *testing.M) {
	os.Exit(testMain(m))
}

func testMain(m *testing.M) int {
	container, err := agriroutertestcontainer.Run(context.Background())
	if err != nil {
		panic(err)
	}
	defer container.TerminateOrLog()
	testContainer = container
	exitCode := m.Run()
	if exitCode != 0 {
		logReader, err := testContainer.Logs(context.Background())
		if err != nil {
			log.Printf("Failed to get logs from test container: %v", err)
		}
		logs, err := io.ReadAll(logReader)
		if err != nil {
			log.Printf("Failed to read logs from test container: %v", err)
		} else {
			log.Printf("Test container logs:\n%s", logs)
		}
	}
	return exitCode
}

func TestIntegration(t *testing.T) {
	client, err := agrirouter.NewClient(testContainer.BaseURL)
	require.NoError(t, err, "Failed to create agrirouter client")

	t.Run("PutEndpoint", func(t *testing.T) {
		externalID := "test-endpoint"
		req := &agrirouter.PutEndpointRequest{

			Capabilities: []agrirouter.EndpointCapability{},
		}

		resp, err := client.PutEndpoint(context.Background(), externalID, &agrirouter.PutEndpointParams{
			AgrirouterTenantId: uuid.New(),
		}, req)
		require.NoError(t, err, "Failed to put endpoint")
		require.NotNil(t, resp, "Response should not be nil")
		require.Equal(t, externalID, resp.ExternalId, "External ID should match")
		events := testContainer.Events

		events.Expect("putEndpoint", `{"externalId":"`+externalID+`"}`)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.NoError(c, events.CheckExpectations(c))
		}, 3*time.Second, 1*time.Second, "Event not received in time")
	})
}
