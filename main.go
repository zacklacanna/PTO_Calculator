package main

import (
	"pto_calculator/tli"
	"strings"

	"github.com/spf13/cobra"
)

func main() {

	// check if user set flag for TUI or CLI
	// default to TUI

	var appType string

	rootCmd := &cobra.Command{
		Use:   "pto_calculator",
		Short: "Calculate your remaining PTO at a certain date",
	}

	rootCmd.RunE = func(cmd *cobra.Command, args []string) error {
		switch strings.ToLower(appType) {
		case "tui":
			return tli.InitTUI()
		case "cli":
			return tli.InitTUI()
		default:
			return tli.InitTUI()
		}
	}

	rootCmd.PersistentFlags().StringVar(&appType, "type", "tui", "ui type: tui|cli")

}
