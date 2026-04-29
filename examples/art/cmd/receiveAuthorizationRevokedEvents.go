package cmd

import (
	"context"
	"fmt"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

var receiveAuthorizationRevokedEventsCmd = &cobra.Command{
	Use:     "receive-authorization-revoked-events",
	Aliases: []string{"rar"},
	Short:   "listens for AUTHORIZATION_REVOKED events from the agrirouter",
	Long: `Subscribes to the events stream and prints every AUTHORIZATION_REVOKED
event received. When this event arrives, access to the named tenant for
the given scope has already been lost and any related local state should
be cleaned up.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		err = client.ReceiveAuthorizationRevokedEvents(ctx, func(ctx context.Context, event *agrirouter.AuthorizationRevokedEventData) {
			fmt.Printf("Authorization revoked:\n")
			fmt.Printf("  TenantID: %s\n", event.TenantId)
			fmt.Printf("  Scope: %s\n", event.Scope)
		}, func(err error) {
			fmt.Printf("Error receiving authorization-revoked events: %v\n", err)
		})
		if err != nil {
			return fmt.Errorf("failed to receive authorization-revoked events: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(receiveAuthorizationRevokedEventsCmd)
}
