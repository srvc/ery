package cmd

import (
	"context"
	"os"
	"os/signal"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/ery"
	"github.com/srvc/ery/pkg/ery/di"
	"github.com/srvc/ery/pkg/util/cliutil"
)

// NewEryCommand creates a new cobra.Command instance.
func NewEryCommand(cfg *ery.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   cfg.Name,
		Short: cfg.Summary,
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			app := di.NewClientApp(cfg)
			return errors.WithStack(runCommand(app, args[0], args[1:]))
		},
	}

	var (
		dnsPort, apiPort uint16
		apiHostname      string
	)

	cliutil.AddLoggingFlags(cmd)
	cmd.PersistentFlags().Uint16Var(&dnsPort, "dns-port", 53, "DNS server runs on the specified port")
	cmd.PersistentFlags().Uint16Var(&apiPort, "api-port", 80, "API server runs on the specified port")
	cmd.PersistentFlags().StringVar(&apiHostname, "api-host", "api.ery", "API server runs on the specified hostname")
	cmd.Flags().SetInterspersed(false)

	cobra.OnInitialize(func() {
		cfg.DNS.Port = domain.Port(dnsPort)
		cfg.API.Port = domain.Port(apiPort)
		cfg.API.Hostname = apiHostname
	})

	cmd.AddCommand(
		newCmdInit(cfg),
		newCmdDaemon(cfg),
		newCmdStart(cfg),
		newCmdPS(cfg),
		newCmdVersion(cfg),
	)

	return cmd
}

func runCommand(app *di.ClientApp, name string, args []string) error {
	cctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(cctx)

	eg.Go(func() error {
		err := app.CommandRunner.Run(ctx, name, args)
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
