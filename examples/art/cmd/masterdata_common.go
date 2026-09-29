package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"strings"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// Flag names shared by the master-data (AgmaSync) commands. tenantIDOpt and
// endpointIDOpt are defined by other commands and reused here.
const (
	localIDOpt            = "local-id"
	agrirouterIDOpt       = "agrirouter-id"
	baseRevisionOpt       = "base-revision"
	bodyOpt               = "body"
	externalEndpointIDOpt = "external-endpoint-id"
	stateOpt              = "state"
	awaitingUserOpt       = "awaiting-user"
	lastEventIDOpt        = "last-event-id"
	idMappingsOpt         = "id-mappings"
)

// humanize turns a kebab-case entity name into a space-separated label.
func humanize(entity string) string {
	return strings.ReplaceAll(entity, "-", " ")
}

// withArticle prefixes a noun with the appropriate indefinite article.
func withArticle(noun string) string {
	if noun == "" {
		return noun
	}
	switch noun[0] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return "an " + noun
	default:
		return "a " + noun
	}
}

// requiredStringFlag reads a string flag and errors if it is empty.
func requiredStringFlag(cmd *cobra.Command, flag string) (string, error) {
	s, err := cmd.Flags().GetString(flag)
	if err != nil {
		return "", fmt.Errorf("failed to get %s flag: %w", flag, err)
	}
	if s == "" {
		return "", fmt.Errorf("%s is required", flag)
	}
	return s, nil
}

// requiredUUIDFlag reads a required string flag and parses it as a UUID.
func requiredUUIDFlag(cmd *cobra.Command, flag string) (uuid.UUID, error) {
	s, err := requiredStringFlag(cmd, flag)
	if err != nil {
		return uuid.Nil, err
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse %s '%s' as UUID: %w", flag, s, err)
	}
	return id, nil
}

// resolveEndpointID reads the acting endpoint ID from --endpoint-id or $ART_ENDPOINT_ID.
func resolveEndpointID(cmd *cobra.Command) (uuid.UUID, error) {
	return uuidFlagOrEnv(cmd, endpointIDOpt, "ART_ENDPOINT_ID")
}

// resolveTenantID reads the tenant ID from --tenant-id or $ART_TENANT_ID.
func resolveTenantID(cmd *cobra.Command) (uuid.UUID, error) {
	return uuidFlagOrEnv(cmd, tenantIDOpt, "ART_TENANT_ID")
}

// resolveExternalEndpointID reads the external endpoint ID from
// --external-endpoint-id or $ART_EXTERNAL_ENDPOINT_ID.
func resolveExternalEndpointID(cmd *cobra.Command) (string, error) {
	s, err := cmd.Flags().GetString(externalEndpointIDOpt)
	if err != nil {
		return "", fmt.Errorf("failed to get %s flag: %w", externalEndpointIDOpt, err)
	}
	if s == "" {
		s = os.Getenv("ART_EXTERNAL_ENDPOINT_ID")
	}
	if s == "" {
		return "", fmt.Errorf("%s is required (pass --%s or set $ART_EXTERNAL_ENDPOINT_ID)", externalEndpointIDOpt, externalEndpointIDOpt)
	}
	return s, nil
}

// baseRevisionFlag returns the --base-revision value as a pointer, or nil when
// the flag was not set (i.e. a create rather than an update).
func baseRevisionFlag(cmd *cobra.Command) (*int, error) {
	if !cmd.Flags().Changed(baseRevisionOpt) {
		return nil, nil
	}
	v, err := cmd.Flags().GetInt(baseRevisionOpt)
	if err != nil {
		return nil, fmt.Errorf("failed to get %s flag: %w", baseRevisionOpt, err)
	}
	return &v, nil
}

// readJSONBody reads a JSON flag value, either inline or as @path to a file, and
// unmarshals it into T.
func readJSONBody[T any](cmd *cobra.Command, flag string) (*T, error) {
	raw, err := requiredStringFlag(cmd, flag)
	if err != nil {
		return nil, fmt.Errorf("%w (JSON inline or @path)", err)
	}
	data := []byte(raw)
	if path, ok := strings.CutPrefix(raw, "@"); ok {
		data, err = os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s file: %w", flag, err)
		}
	}
	var v T
	if err := json.Unmarshal(data, &v); err != nil {
		return nil, fmt.Errorf("failed to parse %s as JSON: %w", flag, err)
	}
	return &v, nil
}

// printJSON pretty-prints v as JSON to stdout.
func printJSON(v any) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal result as JSON: %w", err)
	}
	fmt.Println(string(b))
	return nil
}

func addLocalIDFlag(cmd *cobra.Command) {
	cmd.Flags().String(localIDOpt, "", "The application's own local identifier for the entity")
	_ = cmd.MarkFlagRequired(localIDOpt)
}

func addAgrirouterIDFlag(cmd *cobra.Command) {
	cmd.Flags().String(agrirouterIDOpt, "", "The canonical agrirouter ID of the entity")
	_ = cmd.MarkFlagRequired(agrirouterIDOpt)
}

func addEndpointIDFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(endpointIDOpt, "e", "", "Acting agrirouter endpoint ID (default: $ART_ENDPOINT_ID)")
}

func addTenantIDFlag(cmd *cobra.Command) {
	cmd.Flags().StringP(tenantIDOpt, "t", "", "Tenant ID the operation is performed in (default: $ART_TENANT_ID)")
}

func addBaseRevisionFlag(cmd *cobra.Command) {
	cmd.Flags().Int(baseRevisionOpt, 0, "Revision this write is based on; required when updating an existing object, omit when creating")
}

func addBodyFlag(cmd *cobra.Command, entity string) {
	cmd.Flags().String(bodyOpt, "", fmt.Sprintf("The %s as JSON, inline or @path to a file", humanize(entity)))
	_ = cmd.MarkFlagRequired(bodyOpt)
}

func addExternalEndpointIDFlag(cmd *cobra.Command) {
	cmd.Flags().String(externalEndpointIDOpt, "", "The endpoint's application-defined external ID (default: $ART_EXTERNAL_ENDPOINT_ID)")
}

// mdPutCmd builds a "put-<entity>" command that sends the entity JSON body and
// prints the canonical object agrirouter returns.
func mdPutCmd[T any](
	entity string,
	setLocalID func(*T, string),
	put func(ctx context.Context, c *agrirouter.Client, localID string, endpointID, tenantID uuid.UUID, baseRevision *int, body *T) (*T, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put-" + entity,
		Short: fmt.Sprintf("Send (create or update) %s", withArticle(humanize(entity))),
		Long: fmt.Sprintf(`Sends %[1]s to agrirouter under the application's own --%[2]s.

The %[5]s body is supplied as JSON via --%[3]s, either inline or as @path to a
file; --%[2]s is written into the body automatically. When updating an existing
object, pass --%[4]s with the revision it was edited from.`,
			withArticle(humanize(entity)), localIDOpt, bodyOpt, baseRevisionOpt, humanize(entity)),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			localID, err := requiredStringFlag(cmd, localIDOpt)
			if err != nil {
				return err
			}
			endpointID, err := resolveEndpointID(cmd)
			if err != nil {
				return err
			}
			tenantID, err := resolveTenantID(cmd)
			if err != nil {
				return err
			}
			baseRev, err := baseRevisionFlag(cmd)
			if err != nil {
				return err
			}
			body, err := readJSONBody[T](cmd, bodyOpt)
			if err != nil {
				return err
			}
			setLocalID(body, localID)

			client, err := getClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to create agrirouter client: %w", err)
			}

			slog.Info("Sending "+humanize(entity), "localID", localID, "endpointID", endpointID, "tenantID", tenantID, "baseRevision", baseRev)

			result, err := put(ctx, client, localID, endpointID, tenantID, baseRev, body)
			if err != nil {
				return fmt.Errorf("failed to put %s: %w", humanize(entity), err)
			}
			return printJSON(result)
		},
	}
	addLocalIDFlag(cmd)
	addEndpointIDFlag(cmd)
	addTenantIDFlag(cmd)
	addBaseRevisionFlag(cmd)
	addBodyFlag(cmd, entity)
	return cmd
}

// mdMappingCmd builds a bind/unbind mapping command (both share the same shape).
func mdMappingCmd(
	use, short, long string,
	fn func(ctx context.Context, c *agrirouter.Client, localID string, agrirouterID, endpointID, tenantID uuid.UUID) error,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		Long:  long,
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			localID, err := requiredStringFlag(cmd, localIDOpt)
			if err != nil {
				return err
			}
			agrirouterID, err := requiredUUIDFlag(cmd, agrirouterIDOpt)
			if err != nil {
				return err
			}
			endpointID, err := resolveEndpointID(cmd)
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

			slog.Info(use, "localID", localID, "agrirouterID", agrirouterID, "endpointID", endpointID, "tenantID", tenantID)

			if err := fn(ctx, client, localID, agrirouterID, endpointID, tenantID); err != nil {
				return fmt.Errorf("%s failed: %w", use, err)
			}
			fmt.Printf("%s: ok (local_id=%s agrirouter_id=%s)\n", use, localID, agrirouterID)
			return nil
		},
	}
	addLocalIDFlag(cmd)
	addAgrirouterIDFlag(cmd)
	addEndpointIDFlag(cmd)
	addTenantIDFlag(cmd)
	return cmd
}

// mdDeactivateCmd builds a "deactivate-<entity>" command.
func mdDeactivateCmd[T any](
	entity string,
	fn func(ctx context.Context, c *agrirouter.Client, localID string, endpointID, tenantID uuid.UUID, baseRevision *int) (*T, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deactivate-" + entity,
		Short: fmt.Sprintf("Deactivate %s", withArticle(humanize(entity))),
		Long: fmt.Sprintf(`Signals that the %[1]s was deactivated in the source system. The canonical
object is kept but marked inactive. When the object already exists, pass --%[2]s
with the revision it was edited from.`, humanize(entity), baseRevisionOpt),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			localID, err := requiredStringFlag(cmd, localIDOpt)
			if err != nil {
				return err
			}
			endpointID, err := resolveEndpointID(cmd)
			if err != nil {
				return err
			}
			tenantID, err := resolveTenantID(cmd)
			if err != nil {
				return err
			}
			baseRev, err := baseRevisionFlag(cmd)
			if err != nil {
				return err
			}

			client, err := getClient(ctx)
			if err != nil {
				return fmt.Errorf("failed to create agrirouter client: %w", err)
			}

			slog.Info("Deactivating "+humanize(entity), "localID", localID, "endpointID", endpointID, "tenantID", tenantID, "baseRevision", baseRev)

			result, err := fn(ctx, client, localID, endpointID, tenantID, baseRev)
			if err != nil {
				return fmt.Errorf("failed to deactivate %s: %w", humanize(entity), err)
			}
			return printJSON(result)
		},
	}
	addLocalIDFlag(cmd)
	addEndpointIDFlag(cmd)
	addTenantIDFlag(cmd)
	addBaseRevisionFlag(cmd)
	return cmd
}

// mdRequestCmd builds a "request-<entity>" command (lazy loading by canonical id).
func mdRequestCmd(
	entity string,
	fn func(ctx context.Context, c *agrirouter.Client, endpointID, tenantID, agrirouterID uuid.UUID) error,
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-" + entity,
		Short: fmt.Sprintf("Request %s (lazy loading)", withArticle(humanize(entity))),
		Long: fmt.Sprintf(`Refetches %s the calling endpoint is entitled to but does not currently
hold. The object arrives asynchronously on the master-data event stream
(see stream-masterdata-events).`, withArticle(humanize(entity))),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			agrirouterID, err := requiredUUIDFlag(cmd, agrirouterIDOpt)
			if err != nil {
				return err
			}
			endpointID, err := resolveEndpointID(cmd)
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

			slog.Info("Requesting "+humanize(entity), "agrirouterID", agrirouterID, "endpointID", endpointID, "tenantID", tenantID)

			if err := fn(ctx, client, endpointID, tenantID, agrirouterID); err != nil {
				return fmt.Errorf("failed to request %s: %w", humanize(entity), err)
			}
			fmt.Printf("request-%s: accepted; the %s will arrive on the master-data event stream if entitled\n", entity, humanize(entity))
			return nil
		},
	}
	addAgrirouterIDFlag(cmd)
	addEndpointIDFlag(cmd)
	addTenantIDFlag(cmd)
	return cmd
}
