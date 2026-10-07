package cmd

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/DKE-Data/agrirouter-sdk-go"
	"github.com/spf13/cobra"
)

var allInitialLoadStates = []agrirouter.InitialLoadState{
	agrirouter.InitialLoadStateLoadingFromAgrirouter,
	agrirouter.InitialLoadStateReconciling,
	agrirouter.InitialLoadStateLoadingToAgrirouter,
	agrirouter.InitialLoadStateCompleted,
}

var setInitialLoadStateCmd = &cobra.Command{
	Use:   "set-initial-load-state",
	Short: "Set an endpoint's initial-load state",
	Long: fmt.Sprintf(`Advances the endpoint's initial-load state and, when confirming reconciliation,
carries the identifier bindings reconciliation produced via --%s (a JSON array
of {"agrirouter_id","local_id"} objects, inline or @path).

Valid --%s values: %s.`, idMappingsOpt, stateOpt, joinInitialLoadStates()),
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

		stateRaw, err := requiredStringFlag(cmd, stateOpt)
		if err != nil {
			return err
		}
		state, err := parseInitialLoadState(stateRaw)
		if err != nil {
			return err
		}

		update := agrirouter.InitialLoadStateUpdate{State: state}

		if raw, _ := cmd.Flags().GetString(idMappingsOpt); raw != "" {
			mappings, err := readJSONBody[[]agrirouter.IDMappingBinding](cmd, idMappingsOpt)
			if err != nil {
				return err
			}
			update.IdMappings = mappings
		}

		client, err := getClient(ctx)
		if err != nil {
			return fmt.Errorf("failed to create agrirouter client: %w", err)
		}

		slog.Info("Setting initial load state", "externalEndpointID", externalEndpointID, "tenantID", tenantID, "state", state)

		status, err := client.SetInitialLoadState(ctx, externalEndpointID, tenantID, update)
		if err != nil {
			return fmt.Errorf("failed to set initial load state: %w", err)
		}
		return printJSON(status)
	},
}

func parseInitialLoadState(raw string) (agrirouter.InitialLoadState, error) {
	want := strings.ToUpper(strings.TrimSpace(raw))
	for _, s := range allInitialLoadStates {
		if string(s) == want {
			return s, nil
		}
	}
	return "", fmt.Errorf("invalid %s %q (valid: %s)", stateOpt, raw, joinInitialLoadStates())
}

func joinInitialLoadStates() string {
	s := make([]string, len(allInitialLoadStates))
	for i, st := range allInitialLoadStates {
		s[i] = string(st)
	}
	return strings.Join(s, ", ")
}

func init() {
	rootCmd.AddCommand(setInitialLoadStateCmd)
	addExternalEndpointIDFlag(setInitialLoadStateCmd)
	addTenantIDFlag(setInitialLoadStateCmd)
	setInitialLoadStateCmd.Flags().String(stateOpt, "", "Target initial-load state: "+joinInitialLoadStates())
	_ = setInitialLoadStateCmd.MarkFlagRequired(stateOpt)
	setInitialLoadStateCmd.Flags().String(idMappingsOpt, "", `Reconciliation bindings as a JSON array of {"agrirouter_id","local_id"} objects, inline or @path`)
	_ = setInitialLoadStateCmd.RegisterFlagCompletionFunc(stateOpt, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		out := make([]string, len(allInitialLoadStates))
		for i, s := range allInitialLoadStates {
			out[i] = string(s)
		}
		return out, cobra.ShellCompDirectiveNoFileComp
	})
}
