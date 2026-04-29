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

// EndpointCapabilityDirection represents the direction of an endpoint's capability.
//
// It indicates whether the endpoint can send, receive, or both send and receive messages of a specific type.
type EndpointCapabilityDirection = internal_models.EndpointCapabilityDirection

// EndpointSubscription is the structure representing an endpoint's subscription.
//
// It defines which message types the endpoint is subscribed to,
// meaning it can receive messages of those types in case if sending endpoint
// has published them.
type EndpointSubscription = internal_models.EndpointSubscription

// PutEndpointParams contains parameters to create or update an endpoint.
type PutEndpointParams = internal_models.PutEndpointParams

// DeleteEndpointParams contains parameters to delete an endpoint.
type DeleteEndpointParams = internal_models.DeleteEndpointParams

// SendMessagesParams contains parameters to send a message.
type SendMessagesParams = internal_models.SendMessagesParams

// CapabilityDirectionSend indicates capability of endpoint to send messages.
const CapabilityDirectionSend = internal_models.SEND

// CapabilityDirectionReceive indicates capability of endpoint to receive messages.
const CapabilityDirectionReceive = internal_models.RECEIVE

// CapabilityDirectionSendReceive indicates capability of endpoint to both send and receive messages.
const CapabilityDirectionSendReceive = internal_models.SENDRECEIVE

// VirtualCommunicationUnit is an endpoint type representing a virtual communication unit.
//
// Virtual communication units like usual communication units represent devices
// that can send and receive agrirouter messages, but they are doing so indirectly
// via their own cloud service, which is not installed inside of a vehicle or a machine.
const VirtualCommunicationUnit = internal_models.VirtualCommunicationUnit

// FarmingSoftware is an endpoint type representing farming software applications.
//
// Farming software applications can send and receive agrirouter messages,
// and they are typically cloud deployed applications that manage agricultural data
// and provide farmers with their own typically web-based user interface.
// Deprecated: use CloudSoftware instead as industry agnostic term for same thing.
const FarmingSoftware = internal_models.FarmingSoftware

// CloudSoftware is an endpoint type representing cloud software applications.
//
// Cloud software applications can send and receive agrirouter messages,
// and they are typically backend applications that could manage data from
// specific industry (i.e agriculture) and provide users with their own UI.
const CloudSoftware = internal_models.CloudSoftware

// ConfirmMessagesParams contains parameters to confirm messages.
type ConfirmMessagesParams = internal_models.ConfirmMessagesParams

// ConfirmMessagesRequest is the request structure for confirming received messages.
type ConfirmMessagesRequest = internal_models.ConfirmMessagesRequest

// MessageConfirmation is a single message confirmation carrying a message ID and endpoint ID.
type MessageConfirmation = internal_models.MessageConfirmation

// MessageReceivedEventData represents the data received in a MessageReceived event.
//
// It contains information about the received message, including its type and payload URI.
// This event would typically be followed by fetching the actual message payload from the provided URI
// and processing it accordingly (e.g. decoding, storing, or acting upon the message content).
type MessageReceivedEventData = internal_models.MessageReceivedEventData

// EndpointDeletedEventData represents the data received in an EndpointDeleted event.
//
// It contains the agrirouter ID and the external ID of the endpoint that was deleted,
// either via this API or by other means (e.g. user-initiated deletion from the agrirouter UI).
type EndpointDeletedEventData = internal_models.EndpointDeletedEventData

// EndpointsListChangedEventData represents the data received in an EndpointsListChanged event.
//
// It is emitted whenever the set of endpoints visible to the application,
// or their respective capabilities and/or routes, change in a tenant.
type EndpointsListChangedEventData = internal_models.EndpointsListChangedEventData

// AuthorizationAddedEventData represents the data received in an AuthorizationAdded event.
//
// It is emitted whenever a user adds an authorization for a tenant to the
// current application.
type AuthorizationAddedEventData = internal_models.AuthorizationAddedEventData

// AuthorizationRevokedEventData represents the data received in an AuthorizationRevoked event.
//
// It is emitted whenever a user revokes an authorization for a tenant from
// the current application.
type AuthorizationRevokedEventData = internal_models.AuthorizationRevokedEventData

// TenantInfo describes an authorized tenant together with its visible endpoints.
type TenantInfo = internal_models.TenantInfo

// TenantEndpointInfo describes a single endpoint visible to the application within a tenant.
type TenantEndpointInfo = internal_models.TenantEndpointInfo

// TenantEndpointCapabilities describes the message types an endpoint can send or receive.
type TenantEndpointCapabilities = internal_models.TenantEndpointCapabilities

// RoutedEndpoints describes route-derived information for an application-owned endpoint.
type RoutedEndpoints = internal_models.RoutedEndpoints

// EndpointRouteMap is a map keyed by agrirouter endpoint ID, listing message types routable between two endpoints.
type EndpointRouteMap = internal_models.EndpointRouteMap

// EndpointType represents the type of an agrirouter endpoint.
//
// There are three main types of endpoints:
// 1. communication_unit: Represents devices installed inside vehicles or machines.
// 2. virtual_communication_unit: Represents virtual devices communicating via their own cloud services.
// 3. farming_software: Represents farming software applications, typically cloud-based.
type EndpointType = internal_models.EndpointType
