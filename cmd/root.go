package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "boojo",
	Short: "Boojo is a cli tool for maintaining digital and extended bullet lists",
	Long:  "Boojo is a cli tool for maintaining digital and extended bullet lists - take care of your tasks, events and notes.",
	Run: func(cmd *cobra.Command, args []string) {

	},
	Version: "0.1.0",
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Oops. An error while executing Boojoo '%s'\n", err)
		os.Exit(1)
	}
}
