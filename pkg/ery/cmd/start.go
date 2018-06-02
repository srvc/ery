package cmd

import "github.com/spf13/cobra"

func newCmdStart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start ery server",
	}
	return cmd
}
