package cmd

import (
	"fmt"
	"log/slog"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var deleteEndpointCmd = &cobra.Command{
	Use:   "delete-endpoint",
	Short: "delete-endpoint removes an existing endpoint from the agrirouter",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		externalID, err := cmd.Flags().GetString(externalIDOpt)
		if err != nil {
			return fmt.Errorf("failed to get external-id flag: %w", err)
		}
		if externalID == "" {
			return fmt.Errorf("external-id flag is required")
		}

		tenantID, err := cmd.Flags().GetString(tenantIDOpt)
		if err != nil {
			return fmt.Errorf("failed to get tenant-id flag: %w", err)
		}
		if tenantID == "" {
			return fmt.Errorf("tenant-id flag is required")
		}
		tenantIDParsed, err := uuid.Parse(tenantID)
		if err != nil {
			return fmt.Errorf("failed to parse tenant-id '%s' as UUID: %w", tenantID, err)
		}

		slog.Info("Deleting endpoint",
			"externalID", externalID,
			"tenantID", tenantIDParsed,
		)

		if err := client.DeleteEndpoint(ctx, externalID, &agrirouter.DeleteEndpointParams{
			XAgrirouterTenantId: tenantIDParsed,
		}); err != nil {
			return fmt.Errorf("failed to delete endpoint: %w", err)
		}

		fmt.Printf("Deleted endpoint with external ID '%s'\n", externalID)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(deleteEndpointCmd)

	deleteEndpointCmd.Flags().String(externalIDOpt, "", "The external ID of the endpoint to delete")
	deleteEndpointCmd.MarkFlagRequired(externalIDOpt)

	deleteEndpointCmd.Flags().StringP(tenantIDOpt, "t", "", "ID of the tenant the endpoint belongs to")
	deleteEndpointCmd.MarkFlagRequired(tenantIDOpt)
}
