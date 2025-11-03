package cmd

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

var sendMessagesCmd = &cobra.Command{
	Use:   "send-messages",
	Short: "send-messages sends one or several messages to the agrirouter",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()
		fromFile, err := cmd.Flags().GetString("from-file")
		if err != nil {
			return fmt.Errorf("failed to get from-file flag: %w", err)
		}

		if fromFile == "" {
			return fmt.Errorf("from-file flag is required")
		}

		fileContent, err := os.ReadFile(fromFile)
		if err != nil {
			return fmt.Errorf("failed to read from specified file '%s': %w", fromFile, err)
		}

		messageType, err := cmd.Flags().GetString("message-type")
		if err != nil {
			return fmt.Errorf("failed to get message-type flag: %w", err)
		}

		if messageType == "" {
			return fmt.Errorf("message-type flag is required")
		}

		endpointID, err := cmd.Flags().GetString("endpoint-id")
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

		tenantID, err := cmd.Flags().GetString("tenant-id")
		if err != nil {
			return fmt.Errorf("failed to get tenant-id flag: %w", err)
		}

		if tenantID == "" {
			return fmt.Errorf("tenant-id flag is required")
		}

		tenantIDParsed, err := uuid.Parse(tenantID)
		if err != nil {
			return fmt.Errorf("failed to parse tenant-id '%s' as UUID: %w", tenantID, err)
		}

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		// get file name from path
		filename := filepath.Base(fromFile)

		err = client.SendMessages(ctx, &agrirouter.SendMessagesParams{
			XAgrirouterIsPublish:     true,
			ContentLength:            int64(len(fileContent)),
			XAgrirouterSentTimestamp: time.Now(),
			XAgrirouterMessageType:   messageType,
			XAgrirouterContextId:     uuid.New().String(),
			XAgrirouterEndpointId:    endpointIDParsed,
			XAgrirouterTenantId:      tenantIDParsed,
			XAgrirouterFilename:      &filename,
		}, bytes.NewReader(fileContent))
		if err != nil {
			return fmt.Errorf("failed to send messages: %w", err)
		}

		fmt.Println("Successfully sent message(s) to agrirouter")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(sendMessagesCmd)

	sendMessagesCmd.Flags().StringP("from-file", "f", "", "Path to file containing payload to send to agrirouter")
	sendMessagesCmd.MarkFlagRequired("from-file")
	sendMessagesCmd.Flags().StringP("endpoint-id", "e", "", "ID of the endpoint to send the message from")
	sendMessagesCmd.MarkFlagRequired("endpoint-id")
	sendMessagesCmd.Flags().StringP("message-type", "m", "", "Type of the message to send")
	sendMessagesCmd.MarkFlagRequired("message-type")

	sendMessagesCmd.Flags().StringP("tenant-id", "t", "", "ID of the tenant to send the message in")
	sendMessagesCmd.MarkFlagRequired("tenant-id")
}
