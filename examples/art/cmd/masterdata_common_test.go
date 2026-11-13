package cmd

import (
	"reflect"
	"testing"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
)

func TestParseMappingCorrections(t *testing.T) {
	id := uuid.MustParse("7f3c2b1a-0000-4000-8000-0000000000a1")

	got, err := parseMappingCorrections([]string{"p1=" + id.String(), "p2="})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := []agrirouter.IDMappingCorrectionEntry{
		agrirouter.MappingCorrection("p1", &id),
		agrirouter.MappingCorrection("p2", nil),
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %+v, want %+v", got, want)
	}
}

func TestParseMappingCorrectionsRejectsInvalidEntries(t *testing.T) {
	for _, raw := range [][]string{
		nil,
		{"p1"},
		{"=" + uuid.NewString()},
		{"p1=not-a-uuid"},
	} {
		if _, err := parseMappingCorrections(raw); err == nil {
			t.Errorf("parseMappingCorrections(%q): expected an error", raw)
		}
	}
}
