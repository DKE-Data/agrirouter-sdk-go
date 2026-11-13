package cmd

import (
	"context"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
)

func init() {
	rootCmd.AddCommand(mdPutCmd("farm",
		func(f *agrirouter.Farm, id string) { f.Content.Id.ReferenceId = &id },
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, archived bool, br *int, body *agrirouter.Farm) (*agrirouter.Farm, error) {
			return c.PutFarm(ctx, localID, ep, tn, archived, br, body)
		}))

	rootCmd.AddCommand(mdMappingCmd("bind-farm-mapping",
		"Bind a local identifier to an existing farm",
		"Declares that the canonical farm in --agrirouter-id is the one this endpoint already knows as --local-id.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.BindMapping(ctx, agrirouter.EntityTypeFarm, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdCorrectMappingsCmd("farm", agrirouter.EntityTypeFarm))

	rootCmd.AddCommand(mdArchiveCmd("farm",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Farm, error) {
			obj, err := c.ArchiveEntity(ctx, agrirouter.EntityTypeFarm, localID, ep, tn, br)
			if err != nil {
				return nil, err
			}
			return obj.Farm, nil
		}))

	rootCmd.AddCommand(mdUnarchiveCmd("farm",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Farm, error) {
			obj, err := c.UnarchiveEntity(ctx, agrirouter.EntityTypeFarm, localID, ep, tn, br)
			if err != nil {
				return nil, err
			}
			return obj.Farm, nil
		}))

	rootCmd.AddCommand(mdRequestCmd("farm",
		func(ctx context.Context, c *agrirouter.Client, ep, tn, ar uuid.UUID) error {
			return c.RequestEntity(ctx, agrirouter.EntityTypeFarm, ep, tn, ar)
		}))
}
