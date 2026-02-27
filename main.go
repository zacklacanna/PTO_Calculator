package main

import (
	"github.com/spf13/cobra"

	"pto_calculator/cli"
)

func main() {

	// check if user set flag for TUI or CLI
	// default to TUI

	var flag = false

	if flag {

	}

	rootCmd := &cobra.Command{
		Use:   "pto_calculator",
		Short: "Calculate your remaining PTO at a certain date",
	}

	rootCmd.AddCommand(cli.GetMainDisplay())

}
