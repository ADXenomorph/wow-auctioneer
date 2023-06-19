package cmd

import (
    "github.com/spf13/cobra"
)

var (
    rootCmd = &cobra.Command{
        Use:          "wow-auctioneer",
        Short:        "wowauc",
        SilenceUsage: true,
    }
)

func ExecuteRootCmd() error {
    return rootCmd.Execute()
}
