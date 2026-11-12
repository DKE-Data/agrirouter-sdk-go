package cmd

import (
	"context"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
)

func init() {
	rootCmd.AddCommand(mdPutCmd("field",
		func(f *agrirouter.Field, id string) { f.Content.Id.ReferenceId = &id },
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, active bool, br *int, body *agrirouter.Field) (*agrirouter.Field, error) {
			return c.PutField(ctx, localID, ep, tn, active, br, body)
		}))

	rootCmd.AddCommand(mdMappingCmd("bind-field-mapping",
		"Bind a local identifier to an existing field",
		"Declares that the canonical field in --agrirouter-id is the one this endpoint already knows as --local-id.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.BindMapping(ctx, agrirouter.EntityTypeField, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdCorrectMappingsCmd("field", agrirouter.EntityTypeField))

	rootCmd.AddCommand(mdDeactivateCmd("field",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Field, error) {
			obj, err := c.DeactivateEntity(ctx, agrirouter.EntityTypeField, localID, ep, tn, br)
			if err != nil {
				return nil, err
			}
			return obj.Field, nil
		}))

	rootCmd.AddCommand(mdReactivateCmd("field",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Field, error) {
			obj, err := c.ReactivateEntity(ctx, agrirouter.EntityTypeField, localID, ep, tn, br)
			if err != nil {
				return nil, err
			}
			return obj.Field, nil
		}))

	rootCmd.AddCommand(mdRequestCmd("field",
		func(ctx context.Context, c *agrirouter.Client, ep, tn, ar uuid.UUID) error {
			return c.RequestEntity(ctx, agrirouter.EntityTypeField, ep, tn, ar)
		}))
}
