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

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/ery/di"
)

func newCmdStart(c di.AppComponent) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start ery server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return errors.WithStack(runStartCommand(c))
		},
	}

	cmd.PersistentFlags().Uint16Var(&c.Config().DNS.Port, "dns-port", 53, "DNS server runs on the specified port")
	cmd.PersistentFlags().Uint16Var(&c.Config().Proxy.DefaultPort, "proxy-port", 80, "Proxy server runs on the specified port in default")

	return cmd
}

func runStartCommand(c di.AppComponent) error {
	svrs := []app.Server{
		c.ProxyServer(),
		c.DNSServer(),
		c.APIServer(),
	}

	cctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(cctx)

	for _, s := range svrs {
		s := s
		time.Sleep(30 * time.Millisecond) // wait for starting proxy server manager
		eg.Go(func() error { return errors.WithStack(s.Serve(ctx)) })
	}

	eg.Go(func() error {
		return errors.WithStack(c.ContainerWatcher().Watch(ctx))
	})

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

	return errors.WithStack(eg.Wait())
}
