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
	archivedOpt           = "archived"
	bodyOpt               = "body"
	externalEndpointIDOpt = "external-endpoint-id"
	stateOpt              = "state"
	lastEventIDOpt        = "last-event-id"
	idMappingsOpt         = "id-mappings"
	entryOpt              = "entry"
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

func addArchivedFlag(cmd *cobra.Command) {
	cmd.Flags().Bool(archivedOpt, false, "Lifecycle state the write states: on a create it sets the initial state "+
		"(--archived creates the object archived); on an update it must equal the current state")
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
	put func(ctx context.Context, c *agrirouter.Client, localID string, endpointID, tenantID uuid.UUID,
		archived bool, baseRevision *int, body *T) (*T, error),
) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "put-" + entity,
		Short: fmt.Sprintf("Send (create or update) %s", withArticle(humanize(entity))),
		Long: fmt.Sprintf(`Sends %[1]s to agrirouter under the application's own --%[2]s.

The %[5]s body is supplied as JSON via --%[3]s, either inline or as @path to a
file; --%[2]s is written into the body automatically. When updating an existing
object, pass --%[4]s with the revision it was edited from.

--%[6]s (default false) is the lifecycle state the write states and overrides
any "archived" in the body. On a create it sets the initial state: --%[6]s
uploads an archived record. On an update it must equal the object's current
state, or the write is rejected; use archive-%[7]s or unarchive-%[7]s to change
it.`,
			withArticle(humanize(entity)), localIDOpt, bodyOpt, baseRevisionOpt, humanize(entity), archivedOpt, entity),
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
			archived, err := cmd.Flags().GetBool(archivedOpt)
			if err != nil {
				return fmt.Errorf("failed to get %s flag: %w", archivedOpt, err)
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

			slog.Info("Sending "+humanize(entity), "localID", localID, "endpointID", endpointID, "tenantID", tenantID,
				"archived", archived, "baseRevision", baseRev)

			result, err := put(ctx, client, localID, endpointID, tenantID, archived, baseRev, body)
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
	addArchivedFlag(cmd)
	addBodyFlag(cmd, entity)
	return cmd
}

// mdMappingCmd builds a "bind-<entity>-mapping" command.
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

// mdCorrectMappingsCmd builds a "correct-<entity>-mappings" command.
func mdCorrectMappingsCmd(entity string, entityType agrirouter.EntityType) *cobra.Command {
	use := "correct-" + entity + "-mappings"
	cmd := &cobra.Command{
		Use:   use,
		Short: fmt.Sprintf("Correct this endpoint's %s mappings", humanize(entity)),
		Long: fmt.Sprintf(`Replaces %[1]s bindings this endpoint's user found wrong. Each --%[2]s is
local-id=agrirouter-id, or local-id= to leave that local identifier mapped to
nothing. The entries are applied whole or not at all, so a crosswise swap is one
call. It creates no revision and is delivered to nobody; it is not a deletion.`,
			humanize(entity), entryOpt),
		Example: fmt.Sprintf("  art %s --%s a=<agrirouter-id-of-b> --%s b=<agrirouter-id-of-a>", use, entryOpt, entryOpt),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := cmd.Context()
			raw, err := cmd.Flags().GetStringArray(entryOpt)
			if err != nil {
				return fmt.Errorf("failed to get %s flag: %w", entryOpt, err)
			}
			entries, err := parseMappingCorrections(raw)
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

			slog.Info(use, "entries", len(entries), "endpointID", endpointID, "tenantID", tenantID)

			if err := client.CorrectMappings(ctx, entityType, entries, endpointID, tenantID); err != nil {
				return fmt.Errorf("%s failed: %w", use, err)
			}
			fmt.Printf("%s: ok (%d entries)\n", use, len(entries))
			return nil
		},
	}
	cmd.Flags().StringArray(entryOpt, nil, "local-id=agrirouter-id, or local-id= for no mapping (repeat for several)")
	_ = cmd.MarkFlagRequired(entryOpt)
	addEndpointIDFlag(cmd)
	addTenantIDFlag(cmd)
	return cmd
}

// parseMappingCorrections turns "local-id=agrirouter-id" or "local-id=" values
// into CorrectMappings entries.
func parseMappingCorrections(raw []string) ([]agrirouter.IDMappingCorrectionEntry, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("at least one --%s is required", entryOpt)
	}
	entries := make([]agrirouter.IDMappingCorrectionEntry, 0, len(raw))
	for _, r := range raw {
		localID, target, ok := strings.Cut(r, "=")
		if !ok || localID == "" {
			return nil, fmt.Errorf("invalid --%s %q: want local-id=agrirouter-id or local-id=", entryOpt, r)
		}
		if target == "" {
			entries = append(entries, agrirouter.MappingCorrection(localID, nil))
			continue
		}
		agrirouterID, err := uuid.Parse(target)
		if err != nil {
			return nil, fmt.Errorf("invalid --%s %q: agrirouter id: %w", entryOpt, r, err)
		}
		entries = append(entries, agrirouter.MappingCorrection(localID, &agrirouterID))
	}
	return entries, nil
}

// lifecycleFn changes the lifecycle state of the entity under localID and
// returns the canonical object as it now stands.
type lifecycleFn[T any] func(ctx context.Context, c *agrirouter.Client, localID string,
	endpointID, tenantID uuid.UUID, baseRevision *int) (*T, error)

// mdArchiveCmd builds an "archive-<entity>" command.
func mdArchiveCmd[T any](entity string, fn lifecycleFn[T]) *cobra.Command {
	return mdLifecycleCmd("archive", "Archiving", entity,
		fmt.Sprintf(`Signals that the %[1]s was archived in the source system. The canonical
object is kept but marked archived. When the object is not archived, pass
--%[2]s with the revision it was edited from; on an archived object it is
ignored. Use unarchive-%[3]s to undo it.`, humanize(entity), baseRevisionOpt, entity), fn)
}

// mdUnarchiveCmd builds an "unarchive-<entity>" command.
func mdUnarchiveCmd[T any](entity string, fn lifecycleFn[T]) *cobra.Command {
	return mdLifecycleCmd("unarchive", "Unarchiving", entity,
		fmt.Sprintf(`Signals that the %[1]s, archived earlier, is no longer archived in the
source system. The canonical object is no longer marked archived; its content
is unchanged. When the object is archived, pass --%[2]s with the revision it
was edited from; on an object that is not archived it is ignored. A put never
changes the lifecycle state, so this is the only way back from
archive-%[3]s.`, humanize(entity), baseRevisionOpt, entity), fn)
}

// mdLifecycleCmd builds a "<verb>-<entity>" command that changes the lifecycle
// state of an entity and prints the canonical object agrirouter returns.
func mdLifecycleCmd[T any](verb, progress, entity, long string, fn lifecycleFn[T]) *cobra.Command {
	cmd := &cobra.Command{
		Use:   verb + "-" + entity,
		Short: fmt.Sprintf("%s %s", strings.ToUpper(verb[:1])+verb[1:], withArticle(humanize(entity))),
		Long:  long,
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

			slog.Info(progress+" "+humanize(entity), "localID", localID, "endpointID", endpointID, "tenantID", tenantID, "baseRevision", baseRev)

			result, err := fn(ctx, client, localID, endpointID, tenantID, baseRev)
			if err != nil {
				return fmt.Errorf("failed to %s %s: %w", verb, humanize(entity), err)
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
