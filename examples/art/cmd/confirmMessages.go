package cmd

import (
	"fmt"
	"log/slog"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const (
	messageIDOpt = "message-id"
	endpointIDOpt = "endpoint-id"
)

var confirmMessagesCmd = &cobra.Command{
	Use:   "confirm-messages",
	Short: "confirm-messages confirms one or more received messages with the agrirouter",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		tenantIDParsed, err := uuidFlagOrEnv(cmd, tenantIDOpt, "ART_TENANT_ID")
		if err != nil {
			return err
		}

		endpointID, err := cmd.Flags().GetString(endpointIDOpt)
		if err != nil {
			return fmt.Errorf("failed to get endpoint-id flag: %w", err)
		}
		if endpointID == "" {
			return fmt.Errorf("endpoint-id flag is required")
		}
		endpointIDParsed, err := uuid.Parse(endpointID)
		if err != nil {
			return fmt.Errorf("failed to parse endpoint-id '%s' as UUID: %w", endpointID, err)
		}

		messageIDs, err := cmd.Flags().GetStringArray(messageIDOpt)
		if err != nil {
			return fmt.Errorf("failed to get message-id flag: %w", err)
		}
		if len(messageIDs) == 0 {
			return fmt.Errorf("at least one --message-id is required")
		}
		confirmations := make([]agrirouter.MessageConfirmation, 0, len(messageIDs))
		for _, messageID := range messageIDs {
			messageIDParsed, err := uuid.Parse(messageID)
			if err != nil {
				return fmt.Errorf("failed to parse message-id '%s' as UUID: %w", messageID, err)
			}
			confirmations = append(confirmations, agrirouter.MessageConfirmation{
				EndpointId: endpointIDParsed,
				MessageId:  messageIDParsed,
			})
		}

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		slog.Info("Confirming messages",
			"tenantID", tenantIDParsed,
			"endpointID", endpointIDParsed,
			"messageIDs", messageIDs,
		)

		if err := client.ConfirmMessages(ctx, &agrirouter.ConfirmMessagesParams{
			XAgrirouterTenantId: tenantIDParsed,
		}, agrirouter.ConfirmMessagesRequest{
			Confirmations: confirmations,
		}); err != nil {
			return fmt.Errorf("failed to confirm messages: %w", err)
		}

		fmt.Printf("Confirmed %d message(s) for endpoint %s\n", len(messageIDs), endpointIDParsed)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(confirmMessagesCmd)

	confirmMessagesCmd.Flags().StringP(tenantIDOpt, "t", "", "ID of the tenant the messages belong to (default: $ART_TENANT_ID)")

	confirmMessagesCmd.Flags().StringP(endpointIDOpt, "e", "", "ID of the endpoint that received the messages being confirmed")
	_ = confirmMessagesCmd.MarkFlagRequired(endpointIDOpt)

	confirmMessagesCmd.Flags().StringArrayP(messageIDOpt, "m", []string{}, "Message ID to confirm (repeat for multiple)")
	_ = confirmMessagesCmd.MarkFlagRequired(messageIDOpt)
}
