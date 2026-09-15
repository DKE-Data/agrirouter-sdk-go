package cmd

import (
	"context"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
)

func init() {
	rootCmd.AddCommand(mdPutCmd("organization",
		func(o *agrirouter.Organization, id string) { o.LocalId = &id },
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int, body *agrirouter.Organization) (*agrirouter.Organization, error) {
			return c.PutOrganization(ctx, localID, ep, tn, br, body)
		}))

	rootCmd.AddCommand(mdMappingCmd("bind-organization-mapping",
		"Bind a local identifier to an existing organization",
		"Declares that the canonical organization in --agrirouter-id is the one this endpoint already knows as --local-id.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.BindOrganizationMapping(ctx, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdMappingCmd("unbind-organization-mapping",
		"Declare that this endpoint no longer holds an organization",
		"Declares that the endpoint no longer holds the canonical organization under --local-id. The canonical object itself is untouched.",
		func(ctx context.Context, c *agrirouter.Client, localID string, ar, ep, tn uuid.UUID) error {
			return c.UnbindOrganizationMapping(ctx, localID, ar, ep, tn)
		}))

	rootCmd.AddCommand(mdDeactivateCmd("organization",
		func(ctx context.Context, c *agrirouter.Client, localID string, ep, tn uuid.UUID, br *int) (*agrirouter.Organization, error) {
			return c.DeactivateOrganization(ctx, localID, ep, tn, br)
		}))

	rootCmd.AddCommand(mdRequestCmd("organization",
		func(ctx context.Context, c *agrirouter.Client, ep, tn, ar uuid.UUID) error {
			return c.RequestOrganization(ctx, ep, tn, ar)
		}))
}
