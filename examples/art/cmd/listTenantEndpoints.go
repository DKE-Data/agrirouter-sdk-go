package cmd

import (
	"fmt"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

var listTenantEndpointsCmd = &cobra.Command{
	Use:     "list-tenant-endpoints",
	Aliases: []string{"lte"},
	Short:   "lists endpoints in one tenant with their capabilities and route information",
	Long: `Calls GET /tenants/{tenantId}/endpoints and prints the current list of
endpoints in the tenant together with their capabilities. Endpoints owned
by the authorized application also include route-derived send/receive maps.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		tenantID, err := uuidFlagOrEnv(cmd, tenantIDOpt, "ART_TENANT_ID")
		if err != nil {
			return err
		}

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		endpoints, err := client.ListTenantEndpoints(ctx, tenantID)
		if err != nil {
			return fmt.Errorf("failed to list tenant endpoints: %w", err)
		}

		fmt.Printf("Endpoints in tenant %s (%d):\n", tenantID, len(endpoints))
		for _, ep := range endpoints {
			printTenantEndpoint(ep, "  ")
		}
		return nil
	},
}

func printTenantEndpoint(ep agrirouter.TenantEndpointInfo, indent string) {
	fmt.Printf("%s- %s (%s)\n", indent, ep.Name, ep.Id)
	fmt.Printf("%s  Type: %s\n", indent, ep.EndpointType)
	if ep.ExternalId != nil {
		fmt.Printf("%s  ExternalID: %s\n", indent, *ep.ExternalId)
	}
	fmt.Printf("%s  ApplicationID: %s\n", indent, ep.ApplicationId)
	fmt.Printf("%s  OwnedByYourApplication: %t\n", indent, ep.OwnedByYourApplication)
	if len(ep.Capabilities.CanSend) > 0 {
		fmt.Printf("%s  CanSend: %v\n", indent, ep.Capabilities.CanSend)
	}
	if len(ep.Capabilities.CanReceive) > 0 {
		fmt.Printf("%s  CanReceive: %v\n", indent, ep.Capabilities.CanReceive)
	}
	if ep.RoutedEndpoints != nil {
		if ep.RoutedEndpoints.CanSendTo != nil && len(*ep.RoutedEndpoints.CanSendTo) > 0 {
			fmt.Printf("%s  CanSendTo:\n", indent)
			for peerID, types := range *ep.RoutedEndpoints.CanSendTo {
				fmt.Printf("%s    %s: %v\n", indent, peerID, types)
			}
		}
		if ep.RoutedEndpoints.CanReceiveFrom != nil && len(*ep.RoutedEndpoints.CanReceiveFrom) > 0 {
			fmt.Printf("%s  CanReceiveFrom:\n", indent)
			for peerID, types := range *ep.RoutedEndpoints.CanReceiveFrom {
				fmt.Printf("%s    %s: %v\n", indent, peerID, types)
			}
		}
	}
}

func init() {
	rootCmd.AddCommand(listTenantEndpointsCmd)

	listTenantEndpointsCmd.Flags().StringP(tenantIDOpt, "t", "", "ID of the tenant to list endpoints for (default: $ART_TENANT_ID)")
}
