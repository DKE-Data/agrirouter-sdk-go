package cmd

import (
	"context"
	"fmt"
	"strings"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

const eventTypesOpt = "types"

var allEventTypes = []agrirouter.EventType{
	agrirouter.EventTypeMessageReceived,
	agrirouter.EventTypeFileReceived,
	agrirouter.EventTypeEndpointDeleted,
	agrirouter.EventTypeEndpointsListChanged,
	agrirouter.EventTypeAuthorizationAdded,
	agrirouter.EventTypeAuthorizationRevoked,
}

var receiveEventsCmd = &cobra.Command{
	Use:     "receive-events",
	Aliases: []string{"re"},
	Short:   "listens for events of any kind from the agrirouter",
	Long: `Subscribes to the agrirouter events stream and prints every received event.

If --types is omitted, all event types are streamed. Otherwise the stream is
restricted to the listed types. Available types: MESSAGE_RECEIVED,
FILE_RECEIVED, ENDPOINT_DELETED, ENDPOINTS_LIST_CHANGED, AUTHORIZATION_ADDED,
AUTHORIZATION_REVOKED.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		rawTypes, err := cmd.Flags().GetStringSlice(eventTypesOpt)
		if err != nil {
			return fmt.Errorf("failed to get %s flag: %w", eventTypesOpt, err)
		}

		types, err := parseEventTypes(rawTypes)
		if err != nil {
			return err
		}

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		handlers := agrirouter.EventHandlers{
			OnMessage: func(ctx context.Context, message *agrirouter.Message) {
				fmt.Printf("[MESSAGE_RECEIVED]\n")
				fmt.Printf("  MessageID: %s\n", message.ID)
				fmt.Printf("  AppMessageID: %s\n", message.AppMessageID)
				fmt.Printf("  Type: %s\n", message.MessageType)
				fmt.Printf("  ReceivingEndpointID: %s\n", message.ReceivingEndpointID)
				fmt.Printf("  PayloadSize: %d bytes\n", len(message.Payload))
			},
			OnFile: func(ctx context.Context, file *agrirouter.File) {
				fmt.Printf("[FILE_RECEIVED]\n")
				fmt.Printf("  Type: %s\n", file.MessageType)
				fmt.Printf("  ReceivingEndpointID: %s\n", file.ReceivingEndpointID)
				fmt.Printf("  Size: %d bytes\n", file.Size)
				fmt.Printf("  MessageIDs: %v\n", file.MessageIDs)
			},
			OnEndpointDeleted: func(ctx context.Context, deletion *agrirouter.DeletedEndpoint) {
				fmt.Printf("[ENDPOINT_DELETED]\n")
				fmt.Printf("  ID: %s\n", deletion.ID)
				fmt.Printf("  ExternalID: %s\n", deletion.ExternalID)
			},
			OnEndpointsListChanged: func(ctx context.Context, event *agrirouter.EndpointsListChangedEventData) {
				fmt.Printf("[ENDPOINTS_LIST_CHANGED] tenant %s, %d endpoints:\n", event.TenantId, len(event.Endpoints))
				for _, ep := range event.Endpoints {
					printTenantEndpoint(ep, "  ")
				}
			},
			OnAuthorizationAdded: func(ctx context.Context, event *agrirouter.AuthorizationAddedEventData) {
				fmt.Printf("[AUTHORIZATION_ADDED] tenant %s, scope %s, %d endpoints:\n",
					event.Tenant.TenantId, event.Scope, len(event.Tenant.Endpoints))
				for _, ep := range event.Tenant.Endpoints {
					printTenantEndpoint(ep, "  ")
				}
			},
			OnAuthorizationRevoked: func(ctx context.Context, event *agrirouter.AuthorizationRevokedEventData) {
				fmt.Printf("[AUTHORIZATION_REVOKED] tenant %s, scope %s\n", event.TenantId, event.Scope)
			},
		}

		err = client.ReceiveEvents(ctx, types, handlers, func(err error) {
			fmt.Printf("Error receiving events: %v\n", err)
		})
		if err != nil {
			return fmt.Errorf("failed to receive events: %w", err)
		}
		return nil
	},
}

func parseEventTypes(raw []string) ([]agrirouter.EventType, error) {
	if len(raw) == 0 {
		return nil, nil
	}
	valid := make(map[string]agrirouter.EventType, len(allEventTypes))
	for _, t := range allEventTypes {
		valid[string(t)] = t
	}
	out := make([]agrirouter.EventType, 0, len(raw))
	for _, r := range raw {
		t, ok := valid[strings.ToUpper(strings.TrimSpace(r))]
		if !ok {
			return nil, fmt.Errorf("invalid event type %q (valid: %s)", r, joinEventTypes(allEventTypes))
		}
		out = append(out, t)
	}
	return out, nil
}

func joinEventTypes(types []agrirouter.EventType) string {
	s := make([]string, len(types))
	for i, t := range types {
		s[i] = string(t)
	}
	return strings.Join(s, ", ")
}

func init() {
	rootCmd.AddCommand(receiveEventsCmd)

	receiveEventsCmd.Flags().StringSlice(eventTypesOpt, nil,
		"Event types to receive. Comma-separated or repeated --types. If omitted, all event types are streamed.")
	_ = receiveEventsCmd.RegisterFlagCompletionFunc(eventTypesOpt, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		out := make([]string, len(allEventTypes))
		for i, t := range allEventTypes {
			out[i] = string(t)
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	})
}
