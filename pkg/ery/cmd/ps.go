package cmd

import (
	"context"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/srvc/ery/pkg/ery"
	"github.com/srvc/ery/pkg/ery/di"
)

func newCmdPS(cfg *ery.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "ps",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			app := di.NewClientApp(cfg)
			return errors.WithStack(runPSCommand(app, cfg.OutWriter))
		},
	}

	return cmd
}

func runPSCommand(app *di.ClientApp, w io.Writer) error {
	ctx := context.Background()

	mappings, err := app.MappingRepo.List(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	tw := tabwriter.NewWriter(w, 0, 0, 3, ' ', 0)

	fmt.Fprintln(tw, "HOST\tPORT\tTARGET")

	for _, m := range mappings {
		for sPort, dPort := range m.PortMap {
			fmt.Fprintf(tw, "%s\t%d\t%s:%d\n", m.VirtualHost, sPort, m.ProxyHost, dPort)
		}
	}

	return errors.WithStack(tw.Flush())
}
