package cmd

import (
	"context"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
)

func init() {
	rootCmd.AddCommand(mdPutCmd("person",
		func(p *agrirouter.Person, id string) { p.LocalId = &id },
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int, body *agrirouter.Person) (*agrirouter.Person, error) {
			return c.PutPerson(ctx, localID, ep, tn, br, body)
		}))

	rootCmd.AddCommand(mdMappingCmd("bind-person-mapping",
		"Bind a local identifier to an existing person",
		"Declares that the canonical person in --agrirouter-id is the one this endpoint already knows as --local-id.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.BindPersonMapping(ctx, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdMappingCmd("unbind-person-mapping",
		"Declare that this endpoint no longer holds a person",
		"Declares that the endpoint no longer holds the canonical person under --local-id. The canonical object itself is untouched.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.UnbindPersonMapping(ctx, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdDeactivateCmd("person",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Person, error) {
			return c.DeactivatePerson(ctx, localID, ep, tn, br)
		}))

	rootCmd.AddCommand(mdRequestCmd("person",
		func(ctx context.Context, c *agrirouter.Client, ep, tn, ar uuid.UUID) error {
			return c.RequestPerson(ctx, ep, tn, ar)
		}))
}
