package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/ery/pkg/app"
	"github.com/srvc/ery/pkg/ery/di"
)

func newCmdStart(c di.AppComponent) *cobra.Command {
	var asDaemon bool

	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start ery server",
		RunE: func(cmd *cobra.Command, args []string) error {
			if asDaemon {
				return runDaemon(c)
			}
			return runStartCommand(c)
		},
	}

	cmd.PersistentFlags().BoolVarP(&asDaemon, "daemon", "d", false, "Start as daemon")

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

func runDaemon(c di.AppComponent) error {
	d, err := c.DaemonFactory().Get()
	if err == nil {
		var msg string
		msg, err = d.Start()
		zap.L().Debug("start daemon", zap.String("message", msg), zap.Error(err))
		fmt.Fprintln(c.Config().OutWriter, msg)
	}
	if err != nil {
		fmt.Fprintln(c.Config().ErrWriter, err)
	}
	return err
}
