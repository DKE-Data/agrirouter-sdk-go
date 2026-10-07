package cmd

import (
	"fmt"
	"log/slog"

	"github.com/spf13/cobra"
)

var reportUserAttentionCmd = &cobra.Command{
	Use:   "report-user-attention",
	Short: "Report that an endpoint's initial load is waiting on a user",
	Long: `Raises awaiting_user on the endpoint's initial-load status. It names no state,
so it may be sent from any state before COMPLETED; agrirouter clears the flag on
the endpoint's next transition.`,
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

		slog.Info("Reporting user attention", "externalEndpointID", externalEndpointID, "tenantID", tenantID)

		status, err := client.ReportUserAttention(ctx, externalEndpointID, tenantID)
		if err != nil {
			return fmt.Errorf("failed to report user attention: %w", err)
		}
		return printJSON(status)
	},
}

func init() {
	rootCmd.AddCommand(reportUserAttentionCmd)
	addExternalEndpointIDFlag(reportUserAttentionCmd)
	addTenantIDFlag(reportUserAttentionCmd)
}
