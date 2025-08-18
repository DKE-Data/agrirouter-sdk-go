package agrirouter

import (
	internal_models "github.com/DKE-Data/agrirouter-sdk-go/internal/oapi/models"
)

// Re-exporting models for convenience
// This allows users to access models directly from the agrirouter package.
// Users can import "github.com/DKE-Data/agrirouter-sdk-go/pkg/agrirouter"
// and use models like agrirouter.Endpoint, agrirouter.PutEndpointRequest.

// PutEndpointRequest is the request structure for creating or updating an endpoint.
type PutEndpointRequest = internal_models.PutEndpointRequest

// Endpoint is the structure representing an agrirouter endpoint.
type Endpoint = internal_models.Endpoint

// EndpointCapability is the structure representing an endpoint's capability.
//
// It defines what endpoint can do, such as sending or receiving messages of
// specific types.
type EndpointCapability = internal_models.EndpointCapability

// EndpointSubscription is the structure representing an endpoint's subscription.
//
// It defines which message types the endpoint is subscribed to,
// meaning it can receive messages of those types in case if sending endpoint
// has published them.
type EndpointSubscription = internal_models.EndpointSubscription

// PutEndpointParams contains parameters to create or update an endpoint.
type PutEndpointParams = internal_models.PutEndpointParams

// SendMessageParams contains parameters to send a message.
type SendMessageParams = internal_models.SendMessagesParams

// CapabilityDirectionSend indicates capability of endpoint to send messages.
const CapabilityDirectionSend = internal_models.SEND

// CapabilityDirectionReceive indicates capability of endpoint to receive messages.
const CapabilityDirectionReceive = internal_models.RECEIVE

// CapabilityDirectionSendReceive indicates capability of endpoint to both send and receive messages.
const CapabilityDirectionSendReceive = internal_models.SENDRECEIVE
