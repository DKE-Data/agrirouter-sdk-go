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
	MessageType         string    // MessageType is the URN type of the message
	Payload             []byte    // Payload is the raw message payload
	AppMessageID        string    // AppMessageID is the ID assigned by the sending endpoint
	ReceivingEndpointID uuid.UUID // ReceivingEndpointID is the UUID of the endpoint that received the message
	Filename            *string   // Filename is optional as sent by sender endpoint
}

// MessageHandler is a function that handles a received message.
type MessageHandler func(ctx context.Context, message *Message)

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
			Filename:            messageReceivedEvent.Filename,
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
			payload, err := c.fetchMessagePayload(ctx, payloadURI, errorHandler)
			if err != nil {
				errorHandler(err)
				return
			}
			message.Payload = payload
		}

		messageHandler(ctx, message)
	}, errorHandler)
}

func (c *Client) fetchMessagePayload(
	ctx context.Context,
	payloadURIStr string,
	errorHandler func(err error),
) ([]byte, error) {
	payloadURI, err := url.Parse(payloadURIStr)
	if err != nil {
		return nil, err
	}
	req := &http.Request{Method: http.MethodGet, URL: payloadURI}
	resp, err := c.messagePayloadsClient.Do(req.WithContext(ctx))
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
	client := sse.DefaultClient
	client.ResponseValidator = func(r *http.Response) error {
		err := sse.DefaultValidator(r)
		if err != nil {
			// include body in the error for easier debugging
			body, _ := io.ReadAll(r.Body)
			return fmt.Errorf("%w: %v", err, string(body))
		}
		return nil
	}
	httpClient := c.oapiClient.ClientInterface.(*oapi.Client).Client
	client.HTTPClient = httpClient.(*http.Client)
	conn := client.NewConnection(req)
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

// File represents a file received from agrirouter.
//
// Typically files would have larger payloads than messages,
// so the payload is provided as an io.Reader to allow streaming.
type File struct {
	ReceivingEndpointID uuid.UUID // ReceivingEndpointID is the UUID of the endpoint that received the file
	Payload             io.Reader // Payload is the file payload as a stream
	Filename            *string   // Filename is optional as sent by sender endpoint
	MessageType         string    // MessageType is the URN type of the message
	Size                int64     // Size of file payload in bytes
}

// ReceiveFiles listens for incoming files from the agrirouter API and
// calls the provided handler for each received file.
//
// This function blocks until the context is canceled or an error occurs.
// It is recommended to run this function in a separate goroutine.
func (c *Client) ReceiveFiles(
	ctx context.Context,
	fileHandler func(
		ctx context.Context,
		file *File,
	),
	errorHandler func(err error),
) error {
	return c.receiveAndHandleEvents(ctx, &[]internal_models.ReceiveEventsParamsTypes{
		internal_models.FILERECEIVED,
	}, func(event internal_models.GenericEventData) {
		fileReceivedEvent, err := event.AsFileReceivedEventData()
		if err != nil {
			errorHandler(err)
			return
		}
		if fileReceivedEvent.PayloadUri == nil {
			errorHandler(ErrMissingPayload)
			return
		}
		payloadURI := *fileReceivedEvent.PayloadUri
		payload, err := c.fetchFilePayload(ctx, payloadURI, errorHandler, c.filePayloadsClient)
		if err != nil {
			errorHandler(err)
			return
		}
		fileHandler(ctx, &File{
			Payload:             payload,
			ReceivingEndpointID: fileReceivedEvent.ReceivingEndpointId,
			Filename:            fileReceivedEvent.Filename,
			MessageType:         fileReceivedEvent.MessageType,
			Size:                fileReceivedEvent.Size,
		})
	}, errorHandler)
}

func (c *Client) fetchFilePayload(
	ctx context.Context,
	payloadURIStr string,
	errorHandler func(err error),
	httpClient oapi.HttpRequestDoer,
) (io.Reader, error) {
	payloadURI, err := url.Parse(payloadURIStr)
	if err != nil {
		return nil, err
	}
	req := &http.Request{Method: http.MethodGet, URL: payloadURI}
	resp, err := httpClient.Do(req.WithContext(ctx))
	if err != nil {
		err = fmt.Errorf("%w: %v", ErrFailedToFetchPayload, err)
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		closeErr := resp.Body.Close()
		if closeErr != nil {
			errorHandler(fmt.Errorf("%w: %v", ErrToCloseResponseBody, closeErr))
		}
		return nil, fmt.Errorf("%w: received status code was: %d", ErrUnexpectedStatusCodeWhenFetchingPayload, resp.StatusCode)
	}
	return resp.Body, nil
}
