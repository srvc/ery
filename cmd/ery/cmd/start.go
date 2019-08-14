package cmd

import "github.com/spf13/cobra"

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start server",
		RunE: func(c *cobra.Command, args []string) error {
			return nil
		},
	}

	return cmd
}
