// Package agrirouter provides a client for interacting with the new agrirouter API.
package agrirouter

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/DKE-Data/agrirouter-sdk-go/internal/oapi"
)

var (
	// ErrPutEndpointFailed is returned when an error occurs while trying to put an endpoint.
	ErrPutEndpointFailed = errors.New("failed to put endpoint")

	// ErrFailedStatusCode is returned when the agrirouter API returns a status code that is not expected.
	ErrFailedStatusCode = errors.New("unexpected status code received from agrirouter API")

	// ErrAPICallFailed is returned when an API call fails due to network or server issues.
	ErrAPICallFailed = errors.New("API call failed")

	// ErrURLIsInvalid is returned when the provided URL is invalid.
	ErrURLIsInvalid = errors.New("provided URL is invalid")
)

// Client is the structure that allows interaction with the agrirouter API.
type Client struct {
	oapiClient *oapi.ClientWithResponses
	serverURL  *url.URL
}

// NewClient creates a new agrirouter client with the given server URL.
// The server URL should be the base URL of the agrirouter API, e.g. "https://api.qa.agrirouter.farm".
func NewClient(serverURL string, opts ...ClientOption) (*Client, error) {
	parsedURL, err := url.Parse(serverURL)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to parse server URL: %w", ErrURLIsInvalid, err)
	}
	client, err := oapi.NewClientWithResponses(serverURL, opts...)
	if err != nil {
		return nil, err
	}

	return &Client{
		oapiClient: client,
		serverURL:  parsedURL,
	}, nil
}

// PutEndpoint sends a request to the agrirouter API to create or update an endpoint.
//
// Identifier externalId must uniquely identify created or updated endpoint,
// if endpoint with the same externalId already exists, it will be updated, but only
// if client is authorized to change that endpoint, f.e application owns it.
//
// The request body must contain all endpoint capabilities and subscriptions.
// It is not possible to update only part of endpoint configuration.
// For example if subscriptions are not provided, they will be removed (set to empty list).
func (c *Client) PutEndpoint(
	ctx context.Context,
	externalID string,
	params *PutEndpointParams,
	req *PutEndpointRequest,
) (*Endpoint, error) {
	res, err := c.oapiClient.PutEndpointWithResponse(ctx, externalID, params, *req)
	if err != nil {
		return nil, putEndpointError(ErrPutEndpointFailed, err)
	}

	if res.JSON200 != nil {
		return res.JSON200, nil
	}

	if res.JSON201 != nil {
		return res.JSON201, nil
	}

	return nil, putEndpointError(ErrFailedStatusCode, httpResponseToErr(res.HTTPResponse, res.Body))
}

// SendMessages sends a message to the agrirouter API.
//
// The body of the request must be a valid payload of agrirouter message.
func (c *Client) SendMessages(
	ctx context.Context,
	params *SendMessagesParams,
	body io.Reader,
) error {
	res, err := c.oapiClient.SendMessagesWithBodyWithResponse(ctx, params, "application/octet-stream", body)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrAPICallFailed, err)
	}

	if res.StatusCode() == http.StatusOK || res.StatusCode() == http.StatusAccepted {
		return nil
	}

	return fmt.Errorf("%w: unexpected status code %d, body: %s", ErrFailedStatusCode, res.StatusCode(), string(res.Body))
}

func httpResponseToErr(res *http.Response, body []byte) error {
	if res == nil {
		return fmt.Errorf("%w: response is nil", ErrAPICallFailed)
	}
	return fmt.Errorf(
		"%w: error body: %s, status code: %d, contentType: %s",
		ErrAPICallFailed,
		body,
		res.StatusCode,
		res.Header.Get("Content-Type"),
	)
}

func putEndpointError(err error, err2 error) error {
	if err2 == nil {
		return fmt.Errorf("%w: %w", ErrPutEndpointFailed, err)
	}
	return fmt.Errorf("%w: %w: %w", ErrPutEndpointFailed, err, err2)
}

// ClientOption is a type for options that can be passed to the agrirouter client.
type ClientOption = oapi.ClientOption

// WithHTTPClient allows to set a custom HTTP client for the agrirouter client.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return oapi.WithHTTPClient(httpClient)
}

// RequestEditorFn is a function that can modify the request before it is sent.
type RequestEditorFn = oapi.RequestEditorFn

// WithRequestEditorFn allows to set a custom request editor function to modify request before sending it.
func WithRequestEditorFn(fn RequestEditorFn) ClientOption {
	return oapi.WithRequestEditorFn(fn)
}
