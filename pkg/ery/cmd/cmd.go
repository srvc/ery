package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/ery/pkg/ery/di"
	"github.com/srvc/ery/pkg/util/cliutil"
)

// NewEryCommand creates a new cobra.Command instance.
func NewEryCommand(c di.AppComponent) *cobra.Command {
	cmd := &cobra.Command{
		Use:  "ery",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return errors.WithStack(runCommand(c, args[0], args[1:]))
		},
	}
	cmd.FParseErrWhitelist.UnknownFlags = true

	cliutil.AddLoggingFlags(cmd)
	cmd.PersistentFlags().Uint16Var(&c.Config().DNS.Port, "dns-port", 53, "DNS server runs on the specified port")
	cmd.PersistentFlags().Uint16Var(&c.Config().Proxy.DefaultPort, "proxy-port", 80, "Proxy server runs on the specified port in default")
	cmd.Flags().SetInterspersed(false)

	cmd.AddCommand(
		newCmdInit(c),
		newCmdDaemon(c),
		newCmdStart(c),
		newCmdPS(c),
		newCmdVersion(c),
	)

	return cmd
}

func runCommand(c di.AppComponent, name string, args []string) error {
	cctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(cctx)

	eg.Go(func() error {
		err := c.CommandRunner().Run(ctx, name, args)
		cancel()
		return errors.WithStack(err)
	})

	// Observe os signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case sig := <-sigCh:
		zap.L().Debug("received signal", zap.Stringer("signal", sig))
	case <-ctx.Done():
		zap.L().Debug("done context", zap.Error(ctx.Err()))
	}

	cancel()

	signal.Stop(sigCh)
	close(sigCh)

	err := errors.WithStack(eg.Wait())

	if errors.Cause(err) == context.Canceled {
		return nil
	}

	return errors.WithStack(err)
}
