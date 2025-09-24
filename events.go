package agrirouter

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/oapi"
	internal_models "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
	"github.com/google/uuid"
	"github.com/tmaxmax/go-sse"
)

// ErrFailedToFetchPayload is returned when calling agrirouter to fetch message payload fails.
var ErrFailedToFetchPayload = errors.New("failed to fetch payload")

// ErrUnexpectedStatusCodeWhenFetchingPayload is returned when the agrirouter API
// returns an unexpected status code when fetching message payload.
var ErrUnexpectedStatusCodeWhenFetchingPayload = errors.New("unexpected status code when fetching payload")

// ErrFailedToReadPayload is returned when reading the payload from the response fails.
var ErrFailedToReadPayload = errors.New("failed to read payload")

// ErrToCloseResponseBody is returned when closing the response body fails.
var ErrToCloseResponseBody = errors.New("failed to close response body")

// ErrMissingPayload is returned when agrirouter returned no embedded payload with message nor a payload URI.
var ErrMissingPayload = errors.New("missing payload: no embedded payload and no payload URI")

// Message represents a message received from agrirouter.
type Message struct {
	MessageType         string
	Payload             []byte
	AppMessageID        string
	ReceivingEndpointID uuid.UUID
}

// MessageHandler is a function that handles a received message.
type MessageHandler func(message *Message)

// ReceiveMessages listens for incoming messages from the agrirouter API and
// calls the provided handler for each received message.
//
// This function blocks until the context is canceled or an error occurs.
// It is recommended to run this function in a separate goroutine.
func (c *Client) ReceiveMessages(
	ctx context.Context,
	messageHandler MessageHandler,
	errorHandler func(err error),
) error {
	return c.receiveAndHandleEvents(ctx, &[]internal_models.ReceiveEventsParamsTypes{
		internal_models.MESSAGERECEIVED,
	}, func(event internal_models.GenericEventData) {
		messageReceivedEvent, err := event.AsMessageReceivedEventData()
		if err != nil {
			errorHandler(err)
			return
		}
		message := &Message{
			MessageType:         messageReceivedEvent.MessageType,
			AppMessageID:        messageReceivedEvent.AppMessageId,
			ReceivingEndpointID: messageReceivedEvent.ReceivingEndpointId,
		}
		if messageReceivedEvent.PayloadUri == nil {
			if messageReceivedEvent.Payload == nil {
				errorHandler(ErrMissingPayload)
				return
			}
			// payload is embedded in the event
			message.Payload = *messageReceivedEvent.Payload
		} else {
			// payload needs to be fetched remotely
			payloadURI := *messageReceivedEvent.PayloadUri
			payload, err := c.fetchRemotePayload(ctx, payloadURI, errorHandler)
			if err != nil {
				errorHandler(err)
				return
			}
			message.Payload = payload
		}

		messageHandler(message)
	}, errorHandler)
}

func (c *Client) fetchRemotePayload(
	ctx context.Context,
	payloadURIStr string,
	errorHandler func(err error),
) ([]byte, error) {
	payloadURI, err := url.Parse(payloadURIStr)
	if err != nil {
		return nil, err
	}
	httpClient := c.oapiClient.ClientInterface.(*oapi.Client).Client
	req := &http.Request{Method: http.MethodGet, URL: payloadURI}
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrFailedToFetchPayload, err)
		return nil, err
	}
	defer func() {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			errorHandler(fmt.Errorf("%w: %v", ErrToCloseResponseBody, closeErr))
		}
	}()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: received status code was: %d", ErrUnexpectedStatusCodeWhenFetchingPayload, resp.StatusCode)
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrFailedToReadPayload, err)
	}
	return payload, nil
}

func (c *Client) receiveAndHandleEvents(
	ctx context.Context,
	receiveEventsTypes *[]internal_models.ReceiveEventsParamsTypes,
	eventHandler func(event internal_models.GenericEventData),
	errHandler func(err error),
) error {
	req, err := oapi.NewReceiveEventsRequest(c.serverURL.String(), &internal_models.ReceiveEventsParams{
		Types: receiveEventsTypes,
	})
	if err != nil {
		return fmt.Errorf("%w: %v", ErrAPICallFailed, err)
	}
	req = req.WithContext(ctx)
	conn := sse.DefaultClient.NewConnection(req)
	unsubscribe := conn.SubscribeToAll(func(event sse.Event) {
		var genericEvent internal_models.GenericEventData
		jsonErr := json.Unmarshal([]byte(event.Data), &genericEvent)
		if jsonErr != nil {
			errHandler(jsonErr)
			return
		}
		eventHandler(genericEvent)
	})
	defer unsubscribe()
	return conn.Connect()
}
