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
type EndpointCapability = internal_models.EndpointCapability

// PutEndpointParams contains parameters to create or update an endpoint.
type PutEndpointParams = internal_models.PutEndpointParams

// SendMessageParams contains parameters to send a message.
type SendMessageParams = internal_models.SendMessagesParams
