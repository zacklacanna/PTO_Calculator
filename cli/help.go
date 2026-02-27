package cli

import (
	"github.com/spf13/cobra"
)

func GetMainDisplay() *cobra.Command {

	cmd := &cobra.Command{
		Use:   "display",
		Short: "Show display of CLI",
		RunE: func(cmd *cobra.Command, args []string) error {

		},
	}

	return cmd

}
