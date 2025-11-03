package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

const (
	endpointTypeOpt      = "endpoint-type"
	applicationIDOpt     = "application-id"
	softwareVersionIDOpt = "software-version-id"
	tenantIDOpt          = "tenant-id"
	externalIDOpt        = "external-id"
	withCapabilityOpt    = "with-capability"
	withSubscriptionOpt  = "with-subscription"
)

var putEndpointCmd = &cobra.Command{
	Use:   "put-endpoint",
	Short: "put-endpoint registers a new endpoint with the agrirouter, or updates an existing one",
	Long:  ``,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := cmd.Context()

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		externalID, err := cmd.Flags().GetString(externalIDOpt)
		if err != nil {
			return fmt.Errorf("failed to get external-id flag: %w", err)
		}
		if externalID == "" {
			return fmt.Errorf("external-id flag is required")
		}

		tenantID, err := cmd.Flags().GetString(tenantIDOpt)
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

		applicationID, err := cmd.Flags().GetString(applicationIDOpt)
		if err != nil {
			return fmt.Errorf("failed to get application-id flag: %w", err)
		}
		if applicationID == "" {
			return fmt.Errorf("application-id flag is required")
		}
		applicationIDParsed, err := uuid.Parse(applicationID)
		if err != nil {
			return fmt.Errorf("failed to parse application-id '%s' as UUID: %w", applicationID, err)
		}

		endpointType, err := cmd.Flags().GetString(endpointTypeOpt)
		if err != nil {
			return fmt.Errorf("failed to get endpoint-type flag: %w", err)
		}
		if endpointType == "" {
			return fmt.Errorf("endpoint-type flag is required")
		}

		softwareVersionID, err := cmd.Flags().GetString(softwareVersionIDOpt)
		if err != nil {
			return fmt.Errorf("failed to get software-version-id flag: %w", err)
		}
		if softwareVersionID == "" {
			return fmt.Errorf("software-version-id flag is required")
		}
		softwareVersionIDParsed, err := uuid.Parse(softwareVersionID)
		if err != nil {
			return fmt.Errorf("failed to parse software-version-id '%s' as UUID: %w", softwareVersionID, err)
		}

		switch endpointType {
		case string(agrirouter.CommunicationUnit), string(agrirouter.VirtualCommunicationUnit), string(agrirouter.FarmingSoftware):
			// valid
		default:
			return fmt.Errorf("invalid endpoint-type '%s', must be one of: %s, %s, %s", endpointType, agrirouter.CommunicationUnit, agrirouter.VirtualCommunicationUnit, agrirouter.FarmingSoftware)
		}

		var capabilities []agrirouter.EndpointCapability
		capStrs, err := cmd.Flags().GetStringSlice(withCapabilityOpt)
		if err != nil {
			return fmt.Errorf("failed to get with-capability flag: %w", err)
		}
		for _, capStr := range capStrs {
			var direction agrirouter.EndpointCapabilityDirection
			var messageType string
			parts := strings.Split(capStr, "=")
			if len(parts) != 2 {
				return fmt.Errorf("invalid capability format '%s', must be in format '<messageType>=<direction>'", capStr)
			}
			messageType = parts[0]
			direction = agrirouter.EndpointCapabilityDirection(parts[1])
			switch direction {
			case agrirouter.CapabilityDirectionSend, agrirouter.CapabilityDirectionReceive, agrirouter.CapabilityDirectionSendReceive:
				// valid
			default:
				return fmt.Errorf("invalid capability direction '%s' in capability '%s', must be one of: %s, %s, %s", direction, capStr, agrirouter.CapabilityDirectionSend, agrirouter.CapabilityDirectionReceive, agrirouter.CapabilityDirectionSendReceive)
			}
			capabilities = append(capabilities, agrirouter.EndpointCapability{
				MessageType: messageType,
				Direction:   direction,
			})
		}

		var subscriptions []agrirouter.EndpointSubscription = make([]agrirouter.EndpointSubscription, 0)
		subStrs, err := cmd.Flags().GetStringSlice(withSubscriptionOpt)
		if err != nil {
			return fmt.Errorf("failed to get with-subscription flag: %w", err)
		}
		for _, subStr := range subStrs {
			if subStr == "" {
				return fmt.Errorf("invalid subscription format '%s', must be a non-empty message type", subStr)
			}
			subscriptions = append(subscriptions, agrirouter.EndpointSubscription{
				MessageType: subStr,
			})
		}

		slog.Info("Putting endpoint",
			"externalID", externalID,
			"tenantID", tenantIDParsed,
			"applicationID", applicationIDParsed,
			"softwareVersionID", softwareVersionIDParsed,
			"endpointType", endpointType,
			"capabilities", capabilities,
			"subscriptions", subscriptions,
		)

		epResult, err := client.PutEndpoint(ctx, externalID, &agrirouter.PutEndpointParams{
			XAgrirouterTenantId: tenantIDParsed,
		}, &agrirouter.PutEndpointRequest{
			ApplicationId:     applicationIDParsed,
			SoftwareVersionId: softwareVersionIDParsed,
			EndpointType:      agrirouter.EndpointType(endpointType),
			Capabilities:      capabilities,
			Subscriptions:     subscriptions,
		})
		if err != nil {
			return fmt.Errorf("failed to put endpoint: %w", err)
		}

		fmt.Printf("Put endpoint result: %+v\n", epResult)

		return nil
	},
}

func init() {
	rootCmd.AddCommand(putEndpointCmd)

	putEndpointCmd.Flags().String(externalIDOpt, "", "The external ID of the endpoint")
	putEndpointCmd.MarkFlagRequired(externalIDOpt)

	putEndpointCmd.Flags().StringP(tenantIDOpt, "t", "", "ID of the tenant to create endpoint in")
	putEndpointCmd.MarkFlagRequired(tenantIDOpt)

	putEndpointCmd.Flags().String(applicationIDOpt, "", "The application ID of the endpoint")
	putEndpointCmd.MarkFlagRequired(applicationIDOpt)

	putEndpointCmd.Flags().String(softwareVersionIDOpt, "", "The software version ID of the endpoint")
	putEndpointCmd.MarkFlagRequired(softwareVersionIDOpt)

	putEndpointCmd.Flags().String(
		endpointTypeOpt,
		string(agrirouter.FarmingSoftware),
		fmt.Sprintf("The type of the endpoint, available types: %s, %s, %s",
			agrirouter.CommunicationUnit,
			agrirouter.VirtualCommunicationUnit,
			agrirouter.FarmingSoftware,
		),
	)
	putEndpointCmd.MarkFlagRequired(endpointTypeOpt)

	putEndpointCmd.Flags().StringSlice(withCapabilityOpt, []string{}, `Capabilities to assign to the endpoint, 
	every capability should be formatted as '<messageType>=<direction>', 
	where direction is either 'SEND', 'RECEIVE' or 'SEND_RECEIVE', for example: 'iso:11783:-10:taskdata:zip=SEND'`)

	putEndpointCmd.Flags().StringSlice(withSubscriptionOpt, []string{}, `Subscriptions to assign to the endpoint,
	every subscription should be simply a message type, for example: 'iso:11783:-10:taskdata:zip'`)
}
