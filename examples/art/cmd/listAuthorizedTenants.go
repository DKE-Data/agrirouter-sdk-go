package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var listAuthorizedTenantsCmd = &cobra.Command{
	Use:     "list-authorized-tenants",
	Aliases: []string{"lat"},
	Short:   "lists all tenants the application is authorized for, with their visible endpoints",
	Long: `Calls GET /tenants and prints all authorized tenants, with the state of
their most recent authorization grant when present, together with the
endpoints visible to the application in each tenant.

This is intended as the primary global synchronization endpoint after
application startup or recovery.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		tenants, err := client.ListAuthorizedTenants(ctx)
		if err != nil {
			return fmt.Errorf("failed to list authorized tenants: %w", err)
		}

		fmt.Printf("Authorized tenants (%d):\n", len(tenants))
		for _, tenant := range tenants {
			fmt.Printf("- Tenant %s (%d endpoints):\n", tenant.TenantId, len(tenant.Endpoints))
			if tenant.State != nil {
				fmt.Printf("    State: %s\n", *tenant.State)
			}
			for _, ep := range tenant.Endpoints {
				printTenantEndpoint(ep, "    ")
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(listAuthorizedTenantsCmd)
}
