package agrirouter_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeEndpointName(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "already valid",
			input: "My Endpoint 1",
			want:  "My Endpoint 1",
		},
		{
			name:  "all allowed special characters preserved",
			input: "name-1_2.3,4:5",
			want:  "name-1_2.3,4:5",
		},
		{
			name:  "non-latin letters preserved",
			input: "Поле 42 — трактор",
			want:  "Поле 42 _ трактор",
		},
		{
			name:  "disallowed characters replaced with underscore",
			input: "hello@world!",
			want:  "hello_world_",
		},
		{
			name:  "consecutive disallowed characters collapse to single underscore",
			input: "a???b",
			want:  "a_b",
		},
		{
			name:  "consecutive underscores in input also collapse",
			input: "a___b",
			want:  "a_b",
		},
		{
			name:  "mix of underscores and disallowed collapse together",
			input: "a_!_?b",
			want:  "a_b",
		},
		{
			name:  "leading and trailing disallowed produce edge underscores",
			input: "!hello!",
			want:  "_hello_",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := agrirouter.NormalizeEndpointName(tc.input)
			require.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestNormalizeEndpointName_TruncatesToMaxLength(t *testing.T) {
	input := strings.Repeat("a", agrirouter.EndpointNameMaxLength+50)

	got, err := agrirouter.NormalizeEndpointName(input)
	require.NoError(t, err)
	assert.Equal(t, agrirouter.EndpointNameMaxLength, len([]rune(got)))
	assert.Equal(t, strings.Repeat("a", agrirouter.EndpointNameMaxLength), got)
}

func TestNormalizeEndpointName_TruncatesByRunesNotBytes(t *testing.T) {
	// Each `é` is 2 bytes in UTF-8 but counts as one character/rune for the
	// agrirouter length limit.
	input := strings.Repeat("é", agrirouter.EndpointNameMaxLength+10)

	got, err := agrirouter.NormalizeEndpointName(input)
	require.NoError(t, err)
	assert.Equal(t, agrirouter.EndpointNameMaxLength, len([]rune(got)))
}

func TestNormalizeEndpointName_BlankInputs(t *testing.T) {
	blanks := []struct {
		name  string
		input string
	}{
		{name: "empty string", input: ""},
		{name: "only ASCII spaces", input: "    "},
	}

	for _, tc := range blanks {
		t.Run(tc.name, func(t *testing.T) {
			got, err := agrirouter.NormalizeEndpointName(tc.input)
			assert.Empty(t, got)
			require.Error(t, err)
			assert.True(
				t,
				errors.Is(err, agrirouter.ErrEndpointNameBlank),
				"expected ErrEndpointNameBlank, got %v",
				err,
			)
		})
	}
}

func TestNormalizeEndpointName_NonPrintableReplacedWithUnderscore(t *testing.T) {
	// Tabs, newlines and other disallowed characters are replaced with `_`,
	// not treated as whitespace. Repeated runs collapse to a single underscore.
	got, err := agrirouter.NormalizeEndpointName("\t\n\r")
	require.NoError(t, err)
	assert.Equal(t, "_", got)
}
