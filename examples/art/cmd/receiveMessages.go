package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

var receiveMessagesCmd = &cobra.Command{
	Use:   "receive-messages",
	Short: "reads messages from the agrirouter",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		savePayloadsTo, err := cmd.Flags().GetString("save-payloads-to")
		if err != nil {
			return fmt.Errorf("failed to get save-payloads-to flag: %w", err)
		}

		if savePayloadsTo != "" {
			// Create the directory if it doesn't exist
			if err := os.MkdirAll(savePayloadsTo, 0755); err != nil {
				return fmt.Errorf("failed to create save-payloads-to directory: %w", err)
			}
		}

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		err = client.ReceiveMessages(ctx, func(ctx context.Context, message *agrirouter.Message) {
			fmt.Printf("Received message:\n")
			fmt.Printf("  AppMessageID: %s\n", message.AppMessageID)
			fmt.Printf("  Type: %s\n", message.MessageType)
			fmt.Printf("  ReceivingEndpointID: %s\n", message.ReceivingEndpointID)
			if savePayloadsTo != "" {
				filename := getFilename(message, savePayloadsTo)
				if err := os.WriteFile(filename, message.Payload, 0644); err != nil {
					fmt.Printf("  Failed to save payload to file: %v\n", err)
				} else {
					fmt.Printf("  Payload saved to file: %s\n", filename)
				}
			}
		}, func(err error) {
			fmt.Printf("Error receiving messages: %v\n", err)
		})
		if err != nil {
			return fmt.Errorf("failed to receive messages: %w", err)
		}
		return nil
	},
}

func getFilename(message *agrirouter.Message, savePayloadsTo string) string {
	extension := messageTypeToFileExtension(message.MessageType)
	if extension == "" {
		extension = ".bin"
	}
	return fmt.Sprintf("%s/%s%s", savePayloadsTo, message.AppMessageID, extension)
}

func messageTypeToFileExtension(messageType string) string {
	switch messageType {
	case "iso:11783:-10:taskdata:zip":
		return ".isobus.taskdata.zip"
	case "iso:11783:-10:device_description:protobuf":
		return ".isobus.devicedescription.pb"
	case "iso:11783:-10:time_log:protobuf":
		return ".isobus.timelog.pb"
	case "gps:info":
		return ".gps.info.pb"
	case "img:bmp":
		return ".bmp"
	case "img:jpeg":
		return ".jpeg"
	case "img:png":
		return ".png"
	case "shp:shape:zip":
		return ".shape.zip"
	case "doc:pdf":
		return ".pdf"
	case "vid:avi":
		return ".avi"
	case "vid:mp4":
		return ".mp4"
	case "vid:wmv":
		return ".wmv"
	}
	return ""
}

func init() {
	rootCmd.AddCommand(receiveMessagesCmd)

	receiveMessagesCmd.Flags().String("save-payloads-to", "", "The directory to save payload files to (if empty, payloads are not saved to files)")
}
