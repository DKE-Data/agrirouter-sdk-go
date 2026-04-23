package cmd

import (
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/spf13/cobra"
)

// uuidFlagOrEnv reads a string flag and, if empty, falls back to the named
// environment variable, then parses the result as UUID. Errors if neither is
// set, or if the value is not a valid UUID.
func uuidFlagOrEnv(cmd *cobra.Command, flag, envVar string) (uuid.UUID, error) {
	s, err := cmd.Flags().GetString(flag)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get %s flag: %w", flag, err)
	}
	if s == "" {
		s = os.Getenv(envVar)
	}
	if s == "" {
		return uuid.Nil, fmt.Errorf("%s is required (pass --%s or set $%s)", flag, flag, envVar)
	}
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to parse %s '%s' as UUID: %w", flag, s, err)
	}
	return id, nil
}
