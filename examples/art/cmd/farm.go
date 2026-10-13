package cmd

import (
	"context"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
)

func init() {
	rootCmd.AddCommand(mdPutCmd("farm",
		func(f *agrirouter.Farm, id string) { f.LocalId = &id },
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int, body *agrirouter.Farm) (*agrirouter.Farm, error) {
			return c.PutFarm(ctx, localID, ep, tn, br, body)
		}))

	rootCmd.AddCommand(mdMappingCmd("bind-farm-mapping",
		"Bind a local identifier to an existing farm",
		"Declares that the canonical farm in --agrirouter-id is the one this endpoint already knows as --local-id.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.BindMapping(ctx, agrirouter.EntityTypeFarm, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdMappingCmd("unbind-farm-mapping",
		"Declare that this endpoint no longer holds a farm",
		"Declares that the endpoint no longer holds the canonical farm under --local-id. The canonical object itself is untouched.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.UnbindMapping(ctx, agrirouter.EntityTypeFarm, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdDeactivateCmd("farm",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Farm, error) {
			obj, err := c.DeactivateEntity(ctx, agrirouter.EntityTypeFarm, localID, ep, tn, br)
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
