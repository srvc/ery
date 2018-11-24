package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/ery/pkg/ery/di"
)

func newCmdStart(c di.AppComponent) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start ery server",
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return errors.WithStack(runStartCommand(c))
		},
	}

	return cmd
}

func runStartCommand(c di.AppComponent) error {
	cctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(cctx)

	runFuncs := []func(context.Context) error{
		c.DNSServer().Serve,
		c.ProxyServer().ListenMappingEvents,
		c.APIServer().Serve,
		c.ContainerWatcher().ListenEvents,
	}

	for _, f := range runFuncs {
		f := f
		time.Sleep(30 * time.Millisecond) // wait for starting proxy server manager
		eg.Go(func() error { return errors.WithStack(f(ctx)) })
	}

	// Observe os signals
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	select {
	case sig := <-sigCh:
		zap.L().Debug("received signal", zap.Stringer("signal", sig))
		cancel()
	case <-ctx.Done():
		// do nothing
	}

	signal.Stop(sigCh)
	close(sigCh)

	err := errors.WithStack(eg.Wait())

	if errors.Cause(err) == context.Canceled {
		return nil
	}

	return err
}
