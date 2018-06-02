package cmd

import (
	"io"

	"github.com/spf13/cobra"
)

// NewEryCommand creates a new cobra.Command instance.
func NewEryCommand(in io.Reader, out, err io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use: "ery",
	}

	cmd.AddCommand(newCmdStart())

	return cmd
}
