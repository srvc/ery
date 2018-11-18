package cmd

import (
	"context"
	"os"
	"os/signal"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/ery/pkg/ery/di"
)

// NewEryCommand creates a new cobra.Command instance.
func NewEryCommand(c di.AppComponent) *cobra.Command {
	var (
		verbose bool
	)

	cmd := &cobra.Command{
		Use:  "ery",
		Args: cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cmd.SilenceUsage = true
			return errors.WithStack(runCommand(c, args[0], args[1:]))
		},
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose lovel output")
	cmd.PersistentFlags().Uint16Var(&c.Config().DNS.Port, "dns-port", 53, "DNS server runs on the specified port")
	cmd.PersistentFlags().Uint16Var(&c.Config().Proxy.DefaultPort, "proxy-port", 80, "Proxy server runs on the specified port in default")

	cmd.AddCommand(
		newCmdDaemon(c),
		newCmdStart(c),
		newCmdPS(c),
		newCmdVersion(c),
	)

	cobra.OnInitialize(func() {
		setupLogger(verbose)
	})

	return cmd
}

func setupLogger(verbose bool) {
	logger := zap.NewNop()

	if verbose {
		cfg := zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.Level = zap.NewAtomicLevelAt(zapcore.DebugLevel)
		cfg.EncoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
			enc.AppendString(t.Local().Format("2006-01-02 15:04:05 MST"))
		}
		vLogger, err := cfg.Build()
		if err == nil {
			logger = vLogger.Named("ery")
		}
	}

	zap.ReplaceGlobals(logger)
}

func runCommand(c di.AppComponent, name string, args []string) error {
	cctx, cancel := context.WithCancel(context.Background())
	eg, ctx := errgroup.WithContext(cctx)

	eg.Go(func() error {
		return errors.WithStack(c.CommandRunner().Run(ctx, name, args))
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

	err := errors.WithStack(eg.Wait())

	if errors.Cause(err) == context.Canceled {
		return nil
	}

	return errors.WithStack(err)
}
