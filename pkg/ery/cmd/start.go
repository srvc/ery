package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/app/api"
	"github.com/srvc/ery/pkg/app/dns"
	"github.com/srvc/ery/pkg/app/proxy"
	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/util/netutil"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

func newCmdStart() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start ery server",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runStartCommand()
		},
	}
	return cmd
}

func runStartCommand() error {
	// Start servers
	localIP := netutil.LocalhostIP()
	mapper := domain.NewMapper(localIP)

	svrs := []app.Server{
		api.NewServer(mapper),
		dns.NewServer(mapper, localIP),
		proxy.NewServer(mapper),
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
