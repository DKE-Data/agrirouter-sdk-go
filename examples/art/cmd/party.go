package cmd

import (
	"context"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
)

func init() {
	rootCmd.AddCommand(mdPutCmd("party",
		func(p *agrirouter.Party, id string) { p.Content.Id.ReferenceId = &id },
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, archived bool, br *int, body *agrirouter.Party) (*agrirouter.Party, error) {
			return c.PutParty(ctx, localID, ep, tn, archived, br, body)
		}))

	rootCmd.AddCommand(mdMappingCmd("bind-party-mapping",
		"Bind a local identifier to an existing party",
		"Declares that the canonical party in --agrirouter-id is the one this endpoint already knows as --local-id.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.BindMapping(ctx, agrirouter.EntityTypeParty, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdCorrectMappingsCmd("party", agrirouter.EntityTypeParty))

	rootCmd.AddCommand(mdArchiveCmd("party",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Party, error) {
			obj, err := c.ArchiveEntity(ctx, agrirouter.EntityTypeParty, localID, ep, tn, br)
			if err != nil {
				return nil, err
			}
			return obj.Party, nil
		}))

	rootCmd.AddCommand(mdUnarchiveCmd("party",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Party, error) {
			obj, err := c.UnarchiveEntity(ctx, agrirouter.EntityTypeParty, localID, ep, tn, br)
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
