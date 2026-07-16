package cmd

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

var receiveFilesCmd = &cobra.Command{
	Use:   "receive-files",
	Short: "reads files from the agrirouter",
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

		err = client.ReceiveFiles(ctx, func(ctx context.Context, file *agrirouter.File) {
			fmt.Printf("Received file:\n")
			fmt.Printf("  Type: %s\n", file.MessageType)
			fmt.Printf("  ReceivingEndpointID: %s\n", file.ReceivingEndpointID)
			fmt.Printf("  Size: %d\n", file.Size)
			if file.Filename != nil {
				fmt.Printf("  Filename: %s\n", *file.Filename)
			}
			fmt.Printf("  MessageIDs: %v\n", file.MessageIDs)
			if savePayloadsTo != "" {
				filename := getFileFilename(file, savePayloadsTo)
				if err := saveFilePayload(filename, file.Payload); err != nil {
					fmt.Printf("  Failed to save payload to file: %v\n", err)
				} else {
					fmt.Printf("  Payload saved to file: %s\n", filename)
				}
			}
		}, func(err error) {
			fmt.Printf("Error receiving files: %v\n", err)
		})
		if err != nil {
			return fmt.Errorf("failed to receive files: %w", err)
		}
		return nil
	},
}

func getFileFilename(file *agrirouter.File, savePayloadsTo string) string {
	// If the sender provided a filename, use it as-is (it typically already
	// carries an extension).
	if file.Filename != nil && *file.Filename != "" {
		return fmt.Sprintf("%s/%s", savePayloadsTo, *file.Filename)
	}

	name := "file"
	if len(file.MessageIDs) > 0 {
		name = file.MessageIDs[0].String()
	}
	extension := messageTypeToFileExtension(file.MessageType)
	if extension == "" {
		extension = ".bin"
	}
	return fmt.Sprintf("%s/%s%s", savePayloadsTo, name, extension)
}

func saveFilePayload(filename string, payload io.Reader) error {
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()
	if _, err := io.Copy(out, payload); err != nil {
		return err
	}
	return nil
}

func init() {
	rootCmd.AddCommand(receiveFilesCmd)

	receiveFilesCmd.Flags().String("save-payloads-to", "", "The directory to save payload files to (if empty, payloads are not saved to files)")
}
