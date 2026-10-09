package cmd

import (
	"context"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
)

func init() {
	rootCmd.AddCommand(mdPutCmd("field-boundary",
		func(fb *agrirouter.FieldBoundary, id string) { fb.LocalId = &id },
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int, body *agrirouter.FieldBoundary) (*agrirouter.FieldBoundary, error) {
			return c.PutFieldBoundary(ctx, localID, ep, tn, br, body)
		}))

	rootCmd.AddCommand(mdMappingCmd("bind-field-boundary-mapping",
		"Bind a local identifier to an existing field boundary",
		"Declares that the canonical field boundary in --agrirouter-id is the one this endpoint already knows as --local-id.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.BindMapping(ctx, agrirouter.EntityTypeFieldBoundary, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdMappingCmd("unbind-field-boundary-mapping",
		"Declare that this endpoint no longer holds a field boundary",
		"Declares that the endpoint no longer holds the canonical field boundary under --local-id. The canonical object itself is untouched.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.UnbindMapping(ctx, agrirouter.EntityTypeFieldBoundary, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdDeactivateCmd("field-boundary",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.FieldBoundary, error) {
			obj, err := c.DeactivateEntity(ctx, agrirouter.EntityTypeFieldBoundary, localID, ep, tn, br)
			if err != nil {
				return nil, err
			}
			return obj.FieldBoundary, nil
		}))

	rootCmd.AddCommand(mdRequestCmd("field-boundary",
		func(ctx context.Context, c *agrirouter.Client, ep, tn, ar uuid.UUID) error {
			return c.RequestEntity(ctx, agrirouter.EntityTypeFieldBoundary, ep, tn, ar)
		}))
}
