package cmd

import (
	"context"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
)

func init() {
	rootCmd.AddCommand(mdPutCmd("party",
		func(p *agrirouter.Party, id string) { p.LocalId = &id },
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int, body *agrirouter.Party) (*agrirouter.Party, error) {
			return c.PutParty(ctx, localID, ep, tn, br, body)
		}))

	rootCmd.AddCommand(mdMappingCmd("bind-party-mapping",
		"Bind a local identifier to an existing party",
		"Declares that the canonical party in --agrirouter-id is the one this endpoint already knows as --local-id.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.BindMapping(ctx, agrirouter.EntityTypeParty, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdMappingCmd("unbind-party-mapping",
		"Declare that this endpoint no longer holds a party",
		"Declares that the endpoint no longer holds the canonical party under --local-id. The canonical object itself is untouched.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.UnbindMapping(ctx, agrirouter.EntityTypeParty, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdDeactivateCmd("party",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Party, error) {
			obj, err := c.DeactivateEntity(ctx, agrirouter.EntityTypeParty, localID, ep, tn, br)
			if err != nil {
				return nil, err
			}
			return obj.Party, nil
		}))

	rootCmd.AddCommand(mdRequestCmd("party",
		func(ctx context.Context, c *agrirouter.Client, ep, tn, ar uuid.UUID) error {
			return c.RequestEntity(ctx, agrirouter.EntityTypeParty, ep, tn, ar)
		}))
}
