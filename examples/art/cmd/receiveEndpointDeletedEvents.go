package cmd

import (
	"context"
	"fmt"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

var receiveEndpointDeletedEventsCmd = &cobra.Command{
	Use:     "receive-endpoint-deleted-events",
	Aliases: []string{"rede"},
	Short:   "listens for endpoint-deletion events from the agrirouter",
	Long:    ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		err = client.ReceiveEndpointDeletedEvents(ctx, func(ctx context.Context, deletion *agrirouter.DeletedEndpoint) {
			fmt.Printf("Endpoint deleted:\n")
			fmt.Printf("  ID: %s\n", deletion.ID)
			fmt.Printf("  ExternalID: %s\n", deletion.ExternalID)
		}, func(err error) {
			fmt.Printf("Error receiving endpoint-deleted events: %v\n", err)
		})
		if err != nil {
			return fmt.Errorf("failed to receive endpoint-deleted events: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(receiveEndpointDeletedEventsCmd)
}
