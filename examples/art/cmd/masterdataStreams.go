package cmd

import (
	"context"
	"fmt"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

func printMasterdataFrame(_ context.Context, frame *agrirouter.MasterdataEventFrame) {
	fmt.Printf("[%s] id=%s\n%s\n", frame.Event, frame.ID, frame.Data)
}

var streamMasterdataEventsCmd = &cobra.Command{
	Use:     "stream-masterdata-events",
	Aliases: []string{"sme"},
	Short:   "Stream master-data changes (Server-Sent Events)",
	Long: fmt.Sprintf(`Opens the persistent master-data change stream for every tenant and entity type
the application is opted into, and prints every received frame.

Pass --%s to resume after the last durably applied frame (the id: of an earlier
frame). Omit it for a first connection, which redelivers everything the
application is entitled to and ends with a CAUGHT_UP frame.`, lastEventIDOpt),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		lastEventID, err := cmd.Flags().GetString(lastEventIDOpt)
		if err != nil {
			return fmt.Errorf("failed to get %s flag: %w", lastEventIDOpt, err)
		}

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		if err := client.StreamMasterdataEvents(ctx, lastEventID, printMasterdataFrame); err != nil {
			return fmt.Errorf("failed to stream master-data events: %w", err)
		}
		return nil
	},
}

var streamInitialLoadEventsCmd = &cobra.Command{
	Use:     "stream-initial-load-events",
	Aliases: []string{"sile"},
	Short:   "Stream an endpoint's canonical set (Server-Sent Events)",
	Long: `Streams every object of every opted-in entity type the endpoint is entitled to,
while it is loading from agrirouter, and prints every received frame. agrirouter
closes the stream once the whole set has been sent.`,
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

		if err := client.StreamInitialLoadEvents(ctx, externalEndpointID, tenantID, printMasterdataFrame); err != nil {
			return fmt.Errorf("failed to stream initial-load events: %w", err)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(streamMasterdataEventsCmd)
	streamMasterdataEventsCmd.Flags().String(lastEventIDOpt, "", "Opaque last event id to resume from; omit for a first connection")

	rootCmd.AddCommand(streamInitialLoadEventsCmd)
	addExternalEndpointIDFlag(streamInitialLoadEventsCmd)
	addTenantIDFlag(streamInitialLoadEventsCmd)
}
