package agrirouter_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/DKE-Data/agrirouter-sdk-go/internal/tests/agriroutertestcontainer"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testEnvironment struct {
	testContainer *agriroutertestcontainer.AgrirouterContainer
	client        *agrirouter.Client
}

func setupTestEnvironment(t *testing.T) *testEnvironment {
	container, err := agriroutertestcontainer.Run(context.Background())
	require.NoError(t, err, "Failed to start agrirouter test container")

	t.Cleanup(func() {
		if t.Failed() {
			streamContainerLogs(container)
		}
		container.TerminateOrLog()
	})

	client, err := agrirouter.NewClient(
		container.BaseURL,
		agrirouter.WithHTTPClient(http.DefaultClient),
	)
	require.NoError(t, err, "Failed to create agrirouter client")

	// Wait for the test events stream to be ready
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.True(c, container.Events.IsReady())
	}, 5*time.Second, 100*time.Millisecond, "Test events stream not ready")

	return &testEnvironment{
		client:        client,
		testContainer: container,
	}
}

func streamContainerLogs(container *agriroutertestcontainer.AgrirouterContainer) {
	logReader, err := container.Logs(context.Background())
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

func TestPutEndpoint(t *testing.T) {
	env := setupTestEnvironment(t)
	client := env.client
	testContainer := env.testContainer

	t.Run("PutEndpoint", func(t *testing.T) {
		externalID := "urn:test-app:endpoint:1"
		req := &agrirouter.PutEndpointRequest{

			Capabilities: []agrirouter.EndpointCapability{},
		}

		resp, err := client.PutEndpoint(context.Background(), externalID, &agrirouter.PutEndpointParams{
			XAgrirouterTenantId: uuid.New(),
		}, req)
		require.NoError(t, err, "Failed to put endpoint")
		require.NotNil(t, resp, "Response should not be nil")
		require.Equal(t, externalID, resp.ExternalId, "External ID should match")
		events := testContainer.Events

		events.Expect("putEndpoint", `{"externalId":"`+externalID+`"}`)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.NoError(c, events.CheckExpectations(c))
		}, 10*time.Second, 1*time.Second, "Event not received in time")
	})

	t.Run("PutEndpointWithName", func(t *testing.T) {
		externalID := "urn:test-app:endpoint:2"
		name := "My Test Endpoint"
		req := &agrirouter.PutEndpointRequest{
			Name:         &name,
			Capabilities: []agrirouter.EndpointCapability{},
		}

		resp, err := client.PutEndpoint(context.Background(), externalID, &agrirouter.PutEndpointParams{
			XAgrirouterTenantId: uuid.New(),
		}, req)
		require.NoError(t, err, "Failed to put endpoint")
		require.NotNil(t, resp, "Response should not be nil")
		require.Equal(t, externalID, resp.ExternalId, "External ID should match")
		events := testContainer.Events

		events.Expect("putEndpoint", `{"externalId":"`+externalID+`","name":"`+name+`"}`)

		assert.EventuallyWithT(t, func(c *assert.CollectT) {
			assert.NoError(c, events.CheckExpectations(c))
		}, 10*time.Second, 1*time.Second, "Event not received in time")
	})
}

func TestDeleteEndpoint(t *testing.T) {
	env := setupTestEnvironment(t)
	client := env.client
	testContainer := env.testContainer

	externalID := "urn:test-app:endpoint:to-delete"

	err := client.DeleteEndpoint(context.Background(), externalID, &agrirouter.DeleteEndpointParams{
		XAgrirouterTenantId: uuid.New(),
	})
	require.NoError(t, err, "Failed to delete endpoint")

	events := testContainer.Events
	events.Expect("deleteEndpoint", `{"externalId":"`+externalID+`"}`)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.NoError(c, events.CheckExpectations(c))
	}, 10*time.Second, 1*time.Second, "Event not received in time")
}

type testPayload struct {
	bytes      []byte
	encodedB64 string
}

func newTestPayload(size int) *testPayload {
	bytes := make([]byte, size)
	for i := range bytes {
		bytes[i] = byte(i % 256)
	}
	encoded := base64.StdEncoding.EncodeToString(bytes)
	return &testPayload{
		bytes:      bytes,
		encodedB64: encoded,
	}
}

func TestSendMessages(t *testing.T) {
	env := setupTestEnvironment(t)
	client := env.client
	testContainer := env.testContainer

	endpointID := uuid.New()
	tenantID := uuid.New()
	teamsetContextID := "my-teamset-123"
	params := &agrirouter.SendMessagesParams{
		XAgrirouterIsPublish:        true,
		XAgrirouterEndpointId:       endpointID,
		ContentLength:               100,
		XAgrirouterSentTimestamp:    time.Now(),
		XAgrirouterMessageType:      "gps:info",
		XAgrirouterTenantId:         tenantID,
		XAgrirouterContextId:        "test-context",
		XAgrirouterTeamsetContextId: &teamsetContextID,
	}

	payload := newTestPayload(100)
	err := client.SendMessages(context.Background(), params, bytes.NewReader(payload.bytes))
	require.NoError(t, err, "Failed to send messages")
	events := testContainer.Events

	events.Expect("sendMessages", `
    {
      "endpointId":"`+endpointID.String()+`",
      "messageType":"gps:info",
      "payload":"`+payload.encodedB64+`",
	  "appMessageId":"test-context-0",
	  "tenantId":"`+tenantID.String()+`",
	  "teamsetContextId":"my-teamset-123"
    }`)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		log.Printf("Checking expectations for sendMessages event for endpoint %s", endpointID.String())
		assert.NoError(c, events.CheckExpectations(c))
	}, 10*time.Second, 1*time.Second, "Event not received in time")
}

//nolint:funlen // Test function length is acceptable here, test needs to be detailed.
func TestSendAndReceiveMessages(t *testing.T) {
	env := setupTestEnvironment(t)
	client := env.client
	testContainer := env.testContainer

	receivingContext, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	finishedReceiving := make(chan struct{})
	var receiveErr error
	var receivedMessages []*agrirouter.Message
	defer func() {
		// make sure that no errors happened during receiving
		<-finishedReceiving
		require.NoError(t, receiveErr)
	}()

	go func() {
		connectErr := client.ReceiveMessages(receivingContext, func(_ context.Context, message *agrirouter.Message) {
			receivedMessages = append(receivedMessages, message)
		}, func(err error) {
			receiveErr = err
		})
		assert.EqualError(t, connectErr, "context deadline exceeded")
		close(finishedReceiving)
	}()

	endpointID := uuid.New()
	tenantID := uuid.New()
	teamsetContextID := "my-teamset-456"
	params := &agrirouter.SendMessagesParams{
		XAgrirouterIsPublish:        true,
		XAgrirouterEndpointId:       endpointID,
		ContentLength:               100,
		XAgrirouterSentTimestamp:    time.Now(),
		XAgrirouterMessageType:      "gps:info",
		XAgrirouterTenantId:         tenantID,
		XAgrirouterContextId:        "test-context",
		XAgrirouterTeamsetContextId: &teamsetContextID,
	}
	payload := newTestPayload(100)
	err := client.SendMessages(context.Background(), params, bytes.NewReader(payload.bytes))
	require.NoError(t, err, "Failed to put endpoint")
	events := testContainer.Events
	events.Expect("sendMessages",
		`{  "endpointId":"`+endpointID.String()+`",
            "messageType":"gps:info",
            "payload":"`+payload.encodedB64+`",
            "appMessageId":"test-context-0",
            "tenantId":"`+tenantID.String()+`",
            "teamsetContextId":"my-teamset-456"
	     }`)
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.NoError(c, events.CheckExpectations(c))
	}, 10*time.Second, 1*time.Second, "Event not received in time")

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		if assert.Len(c, receivedMessages, 1, "Should have received exactly one message") {
			assert.Equal(c, "gps:info", receivedMessages[0].MessageType, "Message type should match")
			assert.Equal(c, payload.bytes, receivedMessages[0].Payload, "Payload should match")
			assert.Equal(c, "test-context-0", receivedMessages[0].AppMessageID, "AppMessageId should match")
			assert.Equal(c, endpointID, receivedMessages[0].ReceivingEndpointID, "ReceivingEndpointID should match")
			assert.NotNil(c, receivedMessages[0].TenantID, "TenantID should not be nil")
			assert.Equal(c, tenantID.String(), *receivedMessages[0].TenantID, "TenantID should match")
			assert.NotNil(c, receivedMessages[0].TeamsetContextID, "TeamsetContextID should not be nil")
			assert.Equal(c, "my-teamset-456", *receivedMessages[0].TeamsetContextID, "TeamsetContextID should match")
		}
	}, 10*time.Second, 1*time.Second)
}

//nolint:funlen // Test function length is acceptable here, test needs to be detailed.
func TestSendAndReceiveFiles(t *testing.T) {
	env := setupTestEnvironment(t)
	client := env.client
	testContainer := env.testContainer

	receivingContext, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()
	finishedReceiving := make(chan struct{})
	var receiveErr error
	var receivedFiles []*agrirouter.File
	defer func() {
		// make sure that no errors happened during receiving
		<-finishedReceiving
		require.NoError(t, receiveErr)
	}()

	go func() {
		connectErr := client.ReceiveFiles(receivingContext, func(_ context.Context, file *agrirouter.File) {
			receivedFiles = append(receivedFiles, file)
		}, func(err error) {
			receiveErr = err
		})
		assert.EqualError(t, connectErr, "context deadline exceeded")
		close(finishedReceiving)
	}()

	endpointID := uuid.New()
	tenantID := uuid.New()
	filename := "test.png"
	teamsetContextID := "my-teamset-789"
	params := &agrirouter.SendMessagesParams{
		XAgrirouterIsPublish:        true,
		XAgrirouterEndpointId:       endpointID,
		ContentLength:               100,
		XAgrirouterSentTimestamp:    time.Now(),
		XAgrirouterMessageType:      "img:png",
		XAgrirouterTenantId:         tenantID,
		XAgrirouterContextId:        "test-context",
		XAgrirouterFilename:         &filename,
		XAgrirouterTeamsetContextId: &teamsetContextID,
	}
	payload := newTestPayload(100)
	err := client.SendMessages(context.Background(), params, bytes.NewReader(payload.bytes))
	require.NoError(t, err, "Failed to send file")
	events := testContainer.Events
	events.Expect("sendMessages",
		`{  "endpointId":"`+endpointID.String()+`",
            "messageType":"img:png",
            "payload":"`+payload.encodedB64+`",
            "appMessageId":"test-context-0",
            "filename":"test.png",
            "tenantId":"`+tenantID.String()+`",
            "teamsetContextId":"my-teamset-789"
	     }`)
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.NoError(c, events.CheckExpectations(c))
	}, 10*time.Second, 1*time.Second, "Event not received in time")

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		if assert.Len(c, receivedFiles, 1, "Should have received exactly one file") {
			receivedPayload, readErr := io.ReadAll(receivedFiles[0].Payload)
			assert.NoError(c, readErr, "Should be able to read file payload")
			assert.Equal(c, payload.bytes, receivedPayload, "Payload should match")
			assert.Equal(c, endpointID, receivedFiles[0].ReceivingEndpointID, "ReceivingEndpointID should match")
			assert.NotNil(c, receivedFiles[0].Filename, "Filename should not be nil")
			assert.Equal(c, "test.png", *receivedFiles[0].Filename, "Filename should match")
			assert.Equal(c, "img:png", receivedFiles[0].MessageType, "MessageType should match")
			assert.Equal(c, int64(len(payload.bytes)), receivedFiles[0].Size, "Size should match payload length")
			assert.NotEmpty(c, receivedFiles[0].MessageIDs, "MessageIDs should not be empty")
			assert.NotNil(c, receivedFiles[0].TenantID, "TenantID should not be nil")
			assert.Equal(c, tenantID.String(), *receivedFiles[0].TenantID, "TenantID should match")
			assert.NotNil(c, receivedFiles[0].TeamsetContextID, "TeamsetContextID should not be nil")
			assert.Equal(c, "my-teamset-789", *receivedFiles[0].TeamsetContextID, "TeamsetContextID should match")
		}
	}, 10*time.Second, 1*time.Second)
}

func TestConfirmMessages(t *testing.T) {
	env := setupTestEnvironment(t)
	client := env.client
	testContainer := env.testContainer

	messageID := uuid.New()
	endpointID := uuid.New()
	tenantID := uuid.New()

	err := client.ConfirmMessages(context.Background(), &agrirouter.ConfirmMessagesParams{
		XAgrirouterTenantId: tenantID,
	}, agrirouter.ConfirmMessagesRequest{
		Confirmations: []agrirouter.MessageConfirmation{
			{
				MessageId:  messageID,
				EndpointId: endpointID,
			},
		},
	})
	require.NoError(t, err, "Failed to confirm messages")

	events := testContainer.Events
	events.Expect("confirmMessages", `{
		"confirmations": [
			{
				"message_id": "`+messageID.String()+`",
				"endpoint_id": "`+endpointID.String()+`"
			}
		]
	}`)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.NoError(c, events.CheckExpectations(c))
	}, 10*time.Second, 1*time.Second, "Event not received in time")
}

func TestReceiveMessagesFor2SecondsAndStop(t *testing.T) {
	env := setupTestEnvironment(t)
	client := env.client

	twoSecondsContext, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	var receiveErr error
	var connectErr error
	done := make(chan struct{})
	var receivedMessages []*agrirouter.Message
	go func() {
		connectErr = client.ReceiveMessages(twoSecondsContext, func(_ context.Context, message *agrirouter.Message) {
			receivedMessages = append(receivedMessages, message)
		}, func(err error) {
			receiveErr = err
		})
		close(done)
	}()
	<-done

	assert.EqualError(t, connectErr, "context deadline exceeded")
	assert.NoError(t, receiveErr)
	assert.Len(t, receivedMessages, 0, "Should not have received any messages")
}
