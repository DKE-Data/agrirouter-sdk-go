package cmd

import (
	"context"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
)

func init() {
	rootCmd.AddCommand(mdPutCmd("field",
		func(f *agrirouter.Field, id string) { f.LocalId = &id },
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int, body *agrirouter.Field) (*agrirouter.Field, error) {
			return c.PutField(ctx, localID, ep, tn, br, body)
		}))

	rootCmd.AddCommand(mdMappingCmd("bind-field-mapping",
		"Bind a local identifier to an existing field",
		"Declares that the canonical field in --agrirouter-id is the one this endpoint already knows as --local-id.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.BindFieldMapping(ctx, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdMappingCmd("unbind-field-mapping",
		"Declare that this endpoint no longer holds a field",
		"Declares that the endpoint no longer holds the canonical field under --local-id. The canonical object itself is untouched.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.UnbindFieldMapping(ctx, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdDeactivateCmd("field",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Field, error) {
			return c.DeactivateField(ctx, localID, ep, tn, br)
		}))

	rootCmd.AddCommand(mdRequestCmd("field",
		func(ctx context.Context, c *agrirouter.Client, ep, tn, ar uuid.UUID) error {
			return c.RequestField(ctx, ep, tn, ar)
		}))
}
