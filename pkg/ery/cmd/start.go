package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/ery/di"
)

func newCmdStart(c di.AppComponent) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start ery server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStartCommand(c)
		},
	}

	return cmd
}

func runStartCommand(c di.AppComponent) error {
	svrs := []app.Server{
		c.APIServer(),
		c.DNSServer(),
		c.ProxyServer(),
	}

	cctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(cctx)

	for _, s := range svrs {
		s := s
		eg.Go(func() error { return s.Serve(ctx) })
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

	return eg.Wait()
}
