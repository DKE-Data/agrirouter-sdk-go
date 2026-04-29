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

// EventType identifies a kind of event sent by the agrirouter events stream.
type EventType = internal_models.ReceiveEventsParamsTypes

// Event types accepted by [Client.ReceiveEvents] and emitted by the events stream.
const (
	EventTypeMessageReceived      = internal_models.MESSAGERECEIVED
	EventTypeFileReceived         = internal_models.FILERECEIVED
	EventTypeEndpointDeleted      = internal_models.ENDPOINTDELETED
	EventTypeEndpointsListChanged = internal_models.ENDPOINTSLISTCHANGED
	EventTypeAuthorizationAdded   = internal_models.AUTHORIZATIONADDED
	EventTypeAuthorizationRevoked = internal_models.AUTHORIZATIONREVOKED
)

// EventHandlers groups optional per-event-type callbacks for [Client.ReceiveEvents].
//
// Only handlers that are set will be invoked; events whose handler is nil are
// silently dropped after parsing. Note that [Client.ReceiveEvents] does not
// itself filter on handler presence — events are filtered server-side via the
// types argument. If a handler is nil for an event type that was requested,
// matching events still arrive but are discarded.
type EventHandlers struct {
	OnMessage              MessageHandler
	OnFile                 func(ctx context.Context, file *File)
	OnEndpointDeleted      EndpointDeletionHandler
	OnEndpointsListChanged func(ctx context.Context, event *EndpointsListChangedEventData)
	OnAuthorizationAdded   func(ctx context.Context, event *AuthorizationAddedEventData)
	OnAuthorizationRevoked func(ctx context.Context, event *AuthorizationRevokedEventData)
}

// ReceiveEvents listens for events from the agrirouter API and dispatches each
// received event to the matching handler in handlers.
//
// types restricts which event types the server streams. If types is empty or
// nil, the server streams all supported event types.
//
// This function blocks until the context is canceled or an error occurs.
// It is recommended to run this function in a separate goroutine.
func (c *Client) ReceiveEvents(
	ctx context.Context,
	types []EventType,
	handlers EventHandlers,
	errorHandler func(err error),
) error {
	var typesParam *[]internal_models.ReceiveEventsParamsTypes
	if len(types) > 0 {
		t := append([]internal_models.ReceiveEventsParamsTypes(nil), types...)
		typesParam = &t
	}
	return c.receiveAndHandleEvents(ctx, typesParam, func(event internal_models.GenericEventData) {
		c.dispatchEvent(ctx, event, handlers, errorHandler)
	}, errorHandler)
}

func (c *Client) dispatchEvent(
	ctx context.Context,
	event internal_models.GenericEventData,
	handlers EventHandlers,
	errorHandler func(err error),
) {
	discriminator, err := event.Discriminator()
	if err != nil {
		errorHandler(err)
		return
	}
	switch EventType(discriminator) {
	case EventTypeMessageReceived:
		c.dispatchMessageReceived(ctx, event, handlers.OnMessage, errorHandler)
	case EventTypeFileReceived:
		c.dispatchFileReceived(ctx, event, handlers.OnFile, errorHandler)
	case EventTypeEndpointDeleted:
		dispatchEndpointDeleted(ctx, event, handlers.OnEndpointDeleted, errorHandler)
	case EventTypeEndpointsListChanged:
		dispatchEndpointsListChanged(ctx, event, handlers.OnEndpointsListChanged, errorHandler)
	case EventTypeAuthorizationAdded:
		dispatchAuthorizationAdded(ctx, event, handlers.OnAuthorizationAdded, errorHandler)
	case EventTypeAuthorizationRevoked:
		dispatchAuthorizationRevoked(ctx, event, handlers.OnAuthorizationRevoked, errorHandler)
	}
}

func (c *Client) dispatchMessageReceived(
	ctx context.Context,
	event internal_models.GenericEventData,
	handler MessageHandler,
	errorHandler func(err error),
) {
	if handler == nil {
		return
	}
	data, err := event.AsMessageReceivedEventData()
	if err != nil {
		errorHandler(err)
		return
	}
	message, err := c.messageFromEventData(ctx, &data, errorHandler)
	if err != nil {
		errorHandler(err)
		return
	}
	handler(ctx, message)
}

func (c *Client) dispatchFileReceived(
	ctx context.Context,
	event internal_models.GenericEventData,
	handler func(ctx context.Context, file *File),
	errorHandler func(err error),
) {
	if handler == nil {
		return
	}
	data, err := event.AsFileReceivedEventData()
	if err != nil {
		errorHandler(err)
		return
	}
	file, err := c.fileFromEventData(ctx, &data, errorHandler)
	if err != nil {
		errorHandler(err)
		return
	}
	handler(ctx, file)
}

func dispatchEndpointDeleted(
	ctx context.Context,
	event internal_models.GenericEventData,
	handler EndpointDeletionHandler,
	errorHandler func(err error),
) {
	if handler == nil {
		return
	}
	data, err := event.AsEndpointDeletedEventData()
	if err != nil {
		errorHandler(err)
		return
	}
	handler(ctx, &DeletedEndpoint{ID: data.Id, ExternalID: data.ExternalId})
}

func dispatchEndpointsListChanged(
	ctx context.Context,
	event internal_models.GenericEventData,
	handler func(ctx context.Context, event *EndpointsListChangedEventData),
	errorHandler func(err error),
) {
	if handler == nil {
		return
	}
	data, err := event.AsEndpointsListChangedEventData()
	if err != nil {
		errorHandler(err)
		return
	}
	handler(ctx, &data)
}

func dispatchAuthorizationAdded(
	ctx context.Context,
	event internal_models.GenericEventData,
	handler func(ctx context.Context, event *AuthorizationAddedEventData),
	errorHandler func(err error),
) {
	if handler == nil {
		return
	}
	data, err := event.AsAuthorizationAddedEventData()
	if err != nil {
		errorHandler(err)
		return
	}
	handler(ctx, &data)
}

func dispatchAuthorizationRevoked(
	ctx context.Context,
	event internal_models.GenericEventData,
	handler func(ctx context.Context, event *AuthorizationRevokedEventData),
	errorHandler func(err error),
) {
	if handler == nil {
		return
	}
	data, err := event.AsAuthorizationRevokedEventData()
	if err != nil {
		errorHandler(err)
		return
	}
	handler(ctx, &data)
}

func (c *Client) messageFromEventData(
	ctx context.Context,
	data *internal_models.MessageReceivedEventData,
	errorHandler func(err error),
) (*Message, error) {
	message := &Message{
		ID:                  data.Id,
		MessageType:         data.MessageType,
		AppMessageID:        data.AppMessageId,
		ReceivingEndpointID: data.ReceivingEndpointId,
		Filename:            data.Filename,
		TenantID:            data.TenantId,
		TeamsetContextID:    data.TeamsetContextId,
	}
	if data.PayloadUri == nil {
		if data.Payload == nil {
			return nil, ErrMissingPayload
		}
		message.Payload = *data.Payload
		return message, nil
	}
	payload, err := c.fetchMessagePayload(ctx, *data.PayloadUri, errorHandler)
	if err != nil {
		return nil, err
	}
	message.Payload = payload
	return message, nil
}

func (c *Client) fileFromEventData(
	ctx context.Context,
	data *internal_models.FileReceivedEventData,
	errorHandler func(err error),
) (*File, error) {
	if data.PayloadUri == nil {
		return nil, ErrMissingPayload
	}
	payload, err := c.fetchFilePayload(ctx, *data.PayloadUri, errorHandler)
	if err != nil {
		return nil, err
	}
	return &File{
		Payload:             payload,
		ReceivingEndpointID: data.ReceivingEndpointId,
		Filename:            data.Filename,
		MessageType:         data.MessageType,
		Size:                data.Size,
		MessageIDs:          data.MessageIds,
		TenantID:            data.TenantId,
		TeamsetContextID:    data.TeamsetContextId,
	}, nil
}

// Message represents a message received from agrirouter.
type Message struct {
	ID                  uuid.UUID // ID is the agrirouter message ID, generated by agrirouter
	MessageType         string    // MessageType is the URN type of the message
	Payload             []byte    // Payload is the raw message payload
	AppMessageID        string    // AppMessageID is the ID assigned by the sending endpoint
	ReceivingEndpointID uuid.UUID // ReceivingEndpointID is the UUID of the endpoint that received the message
	Filename            *string   // Filename is optional as sent by sender endpoint
	TenantID            *string   // TenantID is the tenant to which the receiving endpoint belongs
	TeamsetContextID    *string   // TeamsetContextID is the teamset context ID provided by the sending application, if any
}

// MessageHandler is a function that handles a received message.
type MessageHandler func(ctx context.Context, message *Message)

// DeletedEndpoint represents an endpoint that has been deleted in agrirouter,
// which is received when server sends us ENDPOINT_DELETED event.
type DeletedEndpoint struct {
	ID         uuid.UUID // ID is the agrirouter endpoint ID of the deleted endpoint
	ExternalID string    // ExternalID is the external ID the endpoint was registered with
}

// EndpointDeletionHandler is a function that handles an endpoint-deletion event.
type EndpointDeletionHandler func(ctx context.Context, deletion *DeletedEndpoint)

// ReceiveEndpointDeletedEvents listens for endpoint-deletion events from the agrirouter API
// and calls the provided handler for each deleted endpoint.
//
// This function blocks until the context is canceled or an error occurs.
// It is recommended to run this function in a separate goroutine.
func (c *Client) ReceiveEndpointDeletedEvents(
	ctx context.Context,
	deletionHandler EndpointDeletionHandler,
	errorHandler func(err error),
) error {
	return c.receiveAndHandleEvents(ctx, &[]internal_models.ReceiveEventsParamsTypes{
		internal_models.ENDPOINTDELETED,
	}, func(event internal_models.GenericEventData) {
		deletedEvent, err := event.AsEndpointDeletedEventData()
		if err != nil {
			errorHandler(err)
			return
		}
		deletionHandler(ctx, &DeletedEndpoint{
			ID:         deletedEvent.Id,
			ExternalID: deletedEvent.ExternalId,
		})
	}, errorHandler)
}

// ReceiveEndpointsListChangedEvents listens for ENDPOINTS_LIST_CHANGED events from the
// agrirouter API and calls the provided handler for each received event.
//
// The event carries the complete current list of endpoints visible to the
// application in the affected tenant.
//
// This function blocks until the context is canceled or an error occurs.
// It is recommended to run this function in a separate goroutine.
func (c *Client) ReceiveEndpointsListChangedEvents(
	ctx context.Context,
	handler func(ctx context.Context, event *EndpointsListChangedEventData),
	errorHandler func(err error),
) error {
	return c.receiveAndHandleEvents(ctx, &[]internal_models.ReceiveEventsParamsTypes{
		internal_models.ENDPOINTSLISTCHANGED,
	}, func(event internal_models.GenericEventData) {
		data, err := event.AsEndpointsListChangedEventData()
		if err != nil {
			errorHandler(err)
			return
		}
		handler(ctx, &data)
	}, errorHandler)
}

// ReceiveAuthorizationAddedEvents listens for AUTHORIZATION_ADDED events from the
// agrirouter API and calls the provided handler for each received event.
//
// This function blocks until the context is canceled or an error occurs.
// It is recommended to run this function in a separate goroutine.
func (c *Client) ReceiveAuthorizationAddedEvents(
	ctx context.Context,
	handler func(ctx context.Context, event *AuthorizationAddedEventData),
	errorHandler func(err error),
) error {
	return c.receiveAndHandleEvents(ctx, &[]internal_models.ReceiveEventsParamsTypes{
		internal_models.AUTHORIZATIONADDED,
	}, func(event internal_models.GenericEventData) {
		data, err := event.AsAuthorizationAddedEventData()
		if err != nil {
			errorHandler(err)
			return
		}
		handler(ctx, &data)
	}, errorHandler)
}

// ReceiveAuthorizationRevokedEvents listens for AUTHORIZATION_REVOKED events from the
// agrirouter API and calls the provided handler for each received event.
//
// This function blocks until the context is canceled or an error occurs.
// It is recommended to run this function in a separate goroutine.
func (c *Client) ReceiveAuthorizationRevokedEvents(
	ctx context.Context,
	handler func(ctx context.Context, event *AuthorizationRevokedEventData),
	errorHandler func(err error),
) error {
	return c.receiveAndHandleEvents(ctx, &[]internal_models.ReceiveEventsParamsTypes{
		internal_models.AUTHORIZATIONREVOKED,
	}, func(event internal_models.GenericEventData) {
		data, err := event.AsAuthorizationRevokedEventData()
		if err != nil {
			errorHandler(err)
			return
		}
		handler(ctx, &data)
	}, errorHandler)
}

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
			ID:                  messageReceivedEvent.Id,
			MessageType:         messageReceivedEvent.MessageType,
			AppMessageID:        messageReceivedEvent.AppMessageId,
			ReceivingEndpointID: messageReceivedEvent.ReceivingEndpointId,
			Filename:            messageReceivedEvent.Filename,
			TenantID:            messageReceivedEvent.TenantId,
			TeamsetContextID:    messageReceivedEvent.TeamsetContextId,
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
	resp, err := c.payloadsClient.Do(req.WithContext(ctx))
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
	ReceivingEndpointID uuid.UUID   // ReceivingEndpointID is the UUID of the endpoint that received the file
	Payload             io.Reader   // Payload is the file payload as a stream
	Filename            *string     // Filename is optional as sent by sender endpoint
	MessageType         string      // MessageType is the URN type of the message
	Size                int64       // Size of file payload in bytes
	MessageIDs          []uuid.UUID // MessageIDs are the agrirouter message IDs of the messages that carried the file payload chunks
	TenantID            *string     // TenantID is the tenant to which the receiving endpoint belongs
	TeamsetContextID    *string     // TeamsetContextID is the teamset context ID provided by the sending application, if any
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
		payload, err := c.fetchFilePayload(ctx, payloadURI, errorHandler)
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
			MessageIDs:          fileReceivedEvent.MessageIds,
			TenantID:            fileReceivedEvent.TenantId,
			TeamsetContextID:    fileReceivedEvent.TeamsetContextId,
		})
	}, errorHandler)
}

func (c *Client) fetchFilePayload(
	ctx context.Context,
	payloadURIStr string,
	errorHandler func(err error),
) (io.Reader, error) {
	payloadURI, err := url.Parse(payloadURIStr)
	if err != nil {
		return nil, err
	}
	req := &http.Request{Method: http.MethodGet, URL: payloadURI}
	resp, err := c.payloadsClient.Do(req.WithContext(ctx))
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
