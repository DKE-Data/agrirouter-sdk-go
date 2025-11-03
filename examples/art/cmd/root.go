package cmd

import (
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var rootCmd = &cobra.Command{
	Use:   "art",
	Short: "A small tool using SDK to communicate with agrirouter",
	Long: `agrirouter is data exchange platform for the agricultural industry.

agrirouter G4 is a new API gateway allowing broader access to agrirouter.
'art' is a simple command line tool demonstrating the usage of 
the agrirouter-sdk-go to interact with agrirouter. It can be useful
by itself in order to test and debug agrirouter via G4 API.`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	viper.AutomaticEnv()
}
