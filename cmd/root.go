/*
Copyright © 2022 Miroslav Safar <msafar@redhat.com>
*/
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "apicurio-sr-sync-cli",
	Short: "apicurio poc sync cli",
	Long:  "",
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(NewSyncCommand())
}
