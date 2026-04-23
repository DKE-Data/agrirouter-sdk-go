package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

const (
	endpointTypeOpt      = "endpoint-type"
	nameOpt              = "name"
	applicationIDOpt     = "application-id"
	softwareVersionIDOpt = "software-version-id"
	tenantIDOpt          = "tenant-id"
	externalIDOpt        = "external-id"
	withCapabilityOpt    = "with-capability"
	withSubscriptionOpt  = "with-subscription"
	allowDeleteByUserOpt = "allow-delete-by-user"
	connectionsURIOpt    = "connections-uri"
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

		tenantIDParsed, err := uuidFlagOrEnv(cmd, tenantIDOpt, "ART_TENANT_ID")
		if err != nil {
			return err
		}

		applicationIDParsed, err := uuidFlagOrEnv(cmd, applicationIDOpt, "ART_APPLICATION_ID")
		if err != nil {
			return err
		}

		endpointType, err := cmd.Flags().GetString(endpointTypeOpt)
		if err != nil {
			return fmt.Errorf("failed to get endpoint-type flag: %w", err)
		}
		if endpointType == "" {
			return fmt.Errorf("endpoint-type flag is required")
		}

		softwareVersionIDParsed, err := uuidFlagOrEnv(cmd, softwareVersionIDOpt, "ART_SOFTWARE_VERSION_ID")
		if err != nil {
			return err
		}

		switch endpointType {
		case string(agrirouter.VirtualCommunicationUnit), string(agrirouter.CloudSoftware):
			// valid
		default:
			return fmt.Errorf("invalid endpoint-type '%s', must be one of: %s, %s", endpointType, agrirouter.VirtualCommunicationUnit, agrirouter.CloudSoftware)
		}

		capabilities := make([]agrirouter.EndpointCapability, 0)
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

		name, err := cmd.Flags().GetString(nameOpt)
		if err != nil {
			return fmt.Errorf("failed to get name flag: %w", err)
		}
		namePtr := &name
		if name == "" {
			namePtr = nil
		}

		connectionsURI, err := cmd.Flags().GetString(connectionsURIOpt)
		if err != nil {
			return fmt.Errorf("failed to get connections-uri flag: %w", err)
		}
		connectionsURIPtr := &connectionsURI
		if connectionsURI == "" {
			connectionsURIPtr = nil
		}

		var allowDeleteByUserPtr *bool
		if cmd.Flags().Changed(allowDeleteByUserOpt) {
			allowDeleteByUser, err := cmd.Flags().GetBool(allowDeleteByUserOpt)
			if err != nil {
				return fmt.Errorf("failed to get allow-delete-by-user flag: %w", err)
			}
			allowDeleteByUserPtr = &allowDeleteByUser
		}

		slog.Info("Putting endpoint",
			"externalID", externalID,
			"name", name,
			"tenantID", tenantIDParsed,
			"applicationID", applicationIDParsed,
			"softwareVersionID", softwareVersionIDParsed,
			"endpointType", endpointType,
			"capabilities", capabilities,
			"subscriptions", subscriptions,
			"allowDeleteByUser", allowDeleteByUserPtr,
			"connectionsURI", connectionsURIPtr,
		)

		epResult, err := client.PutEndpoint(ctx, externalID, &agrirouter.PutEndpointParams{
			XAgrirouterTenantId: tenantIDParsed,
		}, &agrirouter.PutEndpointRequest{
			Name:              namePtr,
			ApplicationId:     applicationIDParsed,
			SoftwareVersionId: softwareVersionIDParsed,
			EndpointType:      agrirouter.EndpointType(endpointType),
			Capabilities:      capabilities,
			Subscriptions:     subscriptions,
			AllowDeleteByUser: allowDeleteByUserPtr,
			ConnectionsUri:    connectionsURIPtr,
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

	putEndpointCmd.Flags().String(nameOpt, "", "Optional name of endpoint")

	putEndpointCmd.Flags().StringP(tenantIDOpt, "t", "", "ID of the tenant to create endpoint in (default: $ART_TENANT_ID)")

	putEndpointCmd.Flags().String(applicationIDOpt, "", "The application ID of the endpoint (default: $ART_APPLICATION_ID)")

	putEndpointCmd.Flags().String(softwareVersionIDOpt, "", "The software version ID of the endpoint (default: $ART_SOFTWARE_VERSION_ID)")

	putEndpointCmd.Flags().String(
		endpointTypeOpt,
		string(agrirouter.CloudSoftware),
		fmt.Sprintf("The type of the endpoint, available types: %s, %s",
			agrirouter.VirtualCommunicationUnit,
			agrirouter.CloudSoftware,
		),
	)
	putEndpointCmd.MarkFlagRequired(endpointTypeOpt)
	_ = putEndpointCmd.RegisterFlagCompletionFunc(endpointTypeOpt, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{
			string(agrirouter.VirtualCommunicationUnit),
			string(agrirouter.CloudSoftware),
		}, cobra.ShellCompDirectiveNoFileComp
	})

	putEndpointCmd.Flags().StringSlice(withCapabilityOpt, []string{}, `Capabilities to assign to the endpoint, 
	every capability should be formatted as '<messageType>=<direction>', 
	where direction is either 'SEND', 'RECEIVE' or 'SEND_RECEIVE', for example: 'iso:11783:-10:taskdata:zip=SEND'`)

	putEndpointCmd.Flags().StringSlice(withSubscriptionOpt, []string{}, `Subscriptions to assign to the endpoint,
	every subscription should be simply a message type, for example: 'iso:11783:-10:taskdata:zip'`)

	putEndpointCmd.Flags().Bool(allowDeleteByUserOpt, false, `Whether the user is allowed to delete this endpoint from agrirouter web interface.
	Note that even when this flag is not set, the user can still force deletion of the endpoint,
	so applications must handle ENDPOINT_DELETED event on a best-effort basis.
	If the flag is not passed at all, the field is omitted from the request and server default applies.`)

	putEndpointCmd.Flags().String(connectionsURIOpt, "", `Optional URI pointing to where the user can manage the entity connected to this endpoint,
	e.g. to disconnect or delete equipment from an equipment vendor. When provided, this URI will be
	shown when the user attempts to delete the endpoint instead of the usual deletion dialog.`)
}
