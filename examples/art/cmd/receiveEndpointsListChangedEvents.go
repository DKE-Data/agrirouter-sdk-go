package cmd

import (
	"context"
	"fmt"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

var receiveEndpointsListChangedEventsCmd = &cobra.Command{
	Use:     "receive-endpoints-list-changed-events",
	Aliases: []string{"relc"},
	Short:   "listens for ENDPOINTS_LIST_CHANGED events from the agrirouter",
	Long: `Subscribes to the events stream and prints every ENDPOINTS_LIST_CHANGED
event received. Each event carries the complete current list of endpoints
visible to the application in the affected tenant.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		err = client.ReceiveEndpointsListChangedEvents(ctx, func(ctx context.Context, event *agrirouter.EndpointsListChangedEventData) {
			fmt.Printf("Endpoints list changed in tenant %s (%d endpoints):\n", event.TenantId, len(event.Endpoints))
			for _, ep := range event.Endpoints {
				printTenantEndpoint(ep, "  ")
			}
		}, func(err error) {
			fmt.Printf("Error receiving endpoints-list-changed events: %v\n", err)
		})
		if err != nil {
			return fmt.Errorf("failed to receive endpoints-list-changed events: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(receiveEndpointsListChangedEventsCmd)
}
