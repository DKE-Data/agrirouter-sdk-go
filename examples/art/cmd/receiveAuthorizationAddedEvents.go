package cmd

import (
	"context"
	"fmt"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

var receiveAuthorizationAddedEventsCmd = &cobra.Command{
	Use:     "receive-authorization-added-events",
	Aliases: []string{"raa"},
	Short:   "listens for AUTHORIZATION_ADDED events from the agrirouter",
	Long: `Subscribes to the events stream and prints every AUTHORIZATION_ADDED
event received. Each event carries the newly authorized tenant together
with its currently visible endpoints.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		err = client.ReceiveAuthorizationAddedEvents(ctx, func(ctx context.Context, event *agrirouter.AuthorizationAddedEventData) {
			fmt.Printf("Authorization added:\n")
			fmt.Printf("  TenantID: %s\n", event.Tenant.TenantId)
			fmt.Printf("  Scope: %s\n", event.Scope)
			fmt.Printf("  Endpoints (%d):\n", len(event.Tenant.Endpoints))
			for _, ep := range event.Tenant.Endpoints {
				printTenantEndpoint(ep, "    ")
			}
		}, func(err error) {
			fmt.Printf("Error receiving authorization-added events: %v\n", err)
		})
		if err != nil {
			return fmt.Errorf("failed to receive authorization-added events: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(receiveAuthorizationAddedEventsCmd)
}
