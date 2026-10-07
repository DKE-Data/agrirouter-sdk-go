package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var getInitialLoadStatusCmd = &cobra.Command{
	Use:   "get-initial-load-status",
	Short: "Get an endpoint's initial-load status",
	Long: `Returns the endpoint's initial-load state, covering every entity type it is
opted into. The endpoint is addressed by its application-defined external ID.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		externalEndpointID, err := resolveExternalEndpointID(cmd)
		if err != nil {
			return err
		}
		tenantID, err := resolveTenantID(cmd)
		if err != nil {
			return err
		}

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		slog.Info("Getting initial load status", "externalEndpointID", externalEndpointID, "tenantID", tenantID)

		status, err := client.GetInitialLoadStatus(ctx, externalEndpointID, tenantID)
		if err != nil {
			return fmt.Errorf("failed to get initial load status: %w", err)
		}
		return printJSON(status)
	},
}

func init() {
	rootCmd.AddCommand(getInitialLoadStatusCmd)
	addExternalEndpointIDFlag(getInitialLoadStatusCmd)
	addTenantIDFlag(getInitialLoadStatusCmd)
}
