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
		externalID := "test-endpoint"
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
	params := &agrirouter.SendMessagesParams{
		XAgrirouterIsPublish:     true,
		XAgrirouterEndpointId:    endpointID,
		ContentLength:            100,
		XAgrirouterSentTimestamp: time.Now(),
		XAgrirouterMessageType:   "img:png",
		XAgrirouterTenantId:      uuid.New(),
		XAgrirouterContextId:     "test-context",
	}

	payload := newTestPayload(100)
	err := client.SendMessages(context.Background(), params, bytes.NewReader(payload.bytes))
	require.NoError(t, err, "Failed to send messages")
	events := testContainer.Events

	events.Expect("sendMessages", `
    {
      "endpointId":"`+endpointID.String()+`",
      "messageType":"img:png",
      "payload":"`+payload.encodedB64+`",
	  "appMessageId":"test-context-0"
    }`)

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		log.Printf("Checking expectations for sendMessages event for endpoint %s", endpointID.String())
		assert.NoError(c, events.CheckExpectations(c))
	}, 10*time.Second, 1*time.Second, "Event not received in time")
}

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
		connectErr := client.ReceiveMessages(receivingContext, func(message *agrirouter.Message) {
			receivedMessages = append(receivedMessages, message)
		}, func(err error) {
			receiveErr = err
		})
		assert.EqualError(t, connectErr, "context deadline exceeded")
		close(finishedReceiving)
	}()

	endpointID := uuid.New()
	params := &agrirouter.SendMessagesParams{
		XAgrirouterIsPublish:     true,
		XAgrirouterEndpointId:    endpointID,
		ContentLength:            100,
		XAgrirouterSentTimestamp: time.Now(),
		XAgrirouterMessageType:   "img:png",
		XAgrirouterTenantId:      uuid.New(),
		XAgrirouterContextId:     "test-context",
	}
	payload := newTestPayload(100)
	err := client.SendMessages(context.Background(), params, bytes.NewReader(payload.bytes))
	require.NoError(t, err, "Failed to put endpoint")
	events := testContainer.Events
	events.Expect("sendMessages",
		`{  "endpointId":"`+endpointID.String()+`",
            "messageType":"img:png",
            "payload":"`+payload.encodedB64+`",
            "appMessageId":"test-context-0"
	     }`)
	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		assert.NoError(c, events.CheckExpectations(c))
	}, 10*time.Second, 1*time.Second, "Event not received in time")

	assert.EventuallyWithT(t, func(c *assert.CollectT) {
		if assert.Len(c, receivedMessages, 1, "Should have received exactly one message") {
			assert.Equal(c, "img:png", receivedMessages[0].MessageType, "Message type should match")
			assert.Equal(c, payload.bytes, receivedMessages[0].Payload, "Payload should match")
			assert.Equal(c, "test-context-0", receivedMessages[0].AppMessageID, "AppMessageId should match")
		}
	}, 10*time.Second, 1*time.Second)
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
		connectErr = client.ReceiveMessages(twoSecondsContext, func(message *agrirouter.Message) {
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
