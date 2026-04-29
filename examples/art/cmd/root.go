package cmd

import (
	"errors"
	"fmt"
	"io/fs"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

var rootCmd = &cobra.Command{
	Use:   "art",
	Short: "A small tool using SDK to communicate with agrirouter",
	Long: `agrirouter is data exchange platform for the agricultural industry.

agrirouter G4 is a new API gateway allowing broader access to agrirouter.
'art' is a simple command line tool demonstrating the usage of
the agrirouter-sdk-go to interact with agrirouter. It can be useful
by itself in order to test and debug agrirouter via G4 API.

Configuration values (AGRIROUTER_OAUTH_*, ART_*) are read from the process
environment, which is automatically populated from a '.env' file in the
current working directory if one exists. Existing environment values are
not overridden by the file.`,
	SilenceUsage: true,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

// loadDotenv reads .env from the current working directory and exports each
// variable that is not already set in the environment. It is intentionally
// silent when the file is missing.
func loadDotenv() {
	if err := gotenv.Load(".env"); err != nil && !errors.Is(err, fs.ErrNotExist) {
		fmt.Fprintf(os.Stderr, "warning: failed to read .env: %v\n", err)
	}
}

func init() {
	loadDotenv()
	viper.AutomaticEnv()
}
