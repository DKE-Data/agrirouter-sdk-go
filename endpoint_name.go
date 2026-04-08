package agrirouter

import (
	"errors"
	"regexp"
	"strings"
)

const (
	// EndpointNameMinLength is the minimum length, in runes, of a valid endpoint name
	// accepted by the agrirouter API.
	EndpointNameMinLength = 1

	// EndpointNameMaxLength is the maximum length, in runes, of a valid endpoint name
	// accepted by the agrirouter API. Names longer than this will be rejected by the
	// server, so [NormalizeEndpointName] truncates to this length.
	EndpointNameMaxLength = 200
)

// ErrEndpointNameBlank is returned by [NormalizeEndpointName] when the input,
// after normalization, would be empty or consist only of whitespace, which is
// not accepted as an endpoint name by the agrirouter API.
var ErrEndpointNameBlank = errors.New("endpoint name is blank or contains no allowed characters")

var (
	endpointNameDisallowedRe      = regexp.MustCompile(`[^\p{L}\p{N} _.,:\-]`)
	endpointNameMultiUnderscoreRe = regexp.MustCompile(`_+`)
)

// NormalizeEndpointName turns an arbitrary string into a value accepted by the
// agrirouter API as an endpoint name (see the `name` property of the put
// endpoint request in openapi.yaml).
//
// The agrirouter API limits names to 1-200 characters and requires every
// character to be either a letter (any script), a digit, a space or one of
// `-`, `_`, `.`, `,`, `:`. Names that consist only of whitespace are also
// rejected by the server.
//
// This helper applies the following normalization to the input:
//  1. Every disallowed character is replaced with a single underscore (`_`).
//  2. Runs of consecutive underscores are collapsed into a single underscore.
//  3. The result is truncated (rune-aware) to [EndpointNameMaxLength].
//
// If the resulting value is empty or contains only whitespace, the function
// returns [ErrEndpointNameBlank]. Otherwise it returns the normalized name.
func NormalizeEndpointName(name string) (string, error) {
	normalized := endpointNameDisallowedRe.ReplaceAllString(name, "_")
	normalized = endpointNameMultiUnderscoreRe.ReplaceAllString(normalized, "_")

	runes := []rune(normalized)
	if len(runes) > EndpointNameMaxLength {
		runes = runes[:EndpointNameMaxLength]
		normalized = string(runes)
	}

	if strings.TrimSpace(normalized) == "" {
		return "", ErrEndpointNameBlank
	}

	return normalized, nil
}
