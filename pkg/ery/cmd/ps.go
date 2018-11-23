package cmd

import (
	"context"
	"fmt"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/srvc/ery/pkg/ery/di"
)

func newCmdPS(c di.AppComponent) *cobra.Command {
	cmd := &cobra.Command{
		Use: "ps",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return errors.WithStack(runPSCommand(c))
		},
	}

	return cmd
}

func runPSCommand(c di.AppComponent) error {
	ctx := context.Background()

	mappings, err := c.RemoteMappingRepository().List(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	w := tabwriter.NewWriter(c.Config().OutWriter, 0, 0, 3, ' ', 0)

	fmt.Fprintln(w, "HOST\tPORT\tTARGET")

	for _, m := range mappings {
		for sPort, dPort := range m.PortMap {
			fmt.Fprintf(w, "%s\t%d\t%s:%d\n", m.VirtualHost, sPort, m.ProxyHost, dPort)
		}
	}

	return errors.WithStack(w.Flush())
}
