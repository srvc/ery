package cmd

import (
	"context"
	"fmt"
	"io/ioutil"
	"os/exec"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/ery/pkg/ery/di"
	"github.com/srvc/ery/pkg/util/netutil"
)

// NewEryCommand creates a new cobra.Command instance.
func NewEryCommand(c di.AppComponent) *cobra.Command {
	var (
		verbose bool
	)

	cmd := &cobra.Command{
		Use:  "ery",
		Args: cobra.ArbitraryArgs,
		RunE: func(_ *cobra.Command, args []string) error {
			return errors.WithStack(runCommand(c, args[0], args[1:]))
		},
	}

	cmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Verbose lovel output")

	cmd.AddCommand(
		newCmdDaemon(c),
		newCmdStart(c),
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
	log := zap.L().Named("exec")
	eg, ctx := errgroup.WithContext(context.Background())

	port, err := netutil.GetFreePort()
	if err != nil {
		return errors.WithStack(err)
	}
	log.Debug("found free port", zap.Int("port", port))

	eg.Go(func() error {
		cmd := exec.CommandContext(ctx, name, args...)
		cmd.Stdin = c.Config().InReader
		cmd.Stdout = c.Config().OutWriter
		cmd.Stderr = c.Config().ErrWriter
		cmd.Env = append(cmd.Env, fmt.Sprintf("PORT=%d", port))
		log.Debug("execute command", zap.String("name", name), zap.Strings("args", args))
		return errors.WithStack(cmd.Run())
	})

	eg.Go(func() error {
		data, err := ioutil.ReadFile("localhost")
		if err != nil {
			return errors.WithStack(err)
		}

		hostnames := []string{}
		for _, h := range strings.Split(string(data), "\n") {
			if h != "" {
				hostnames = append(hostnames, h)
			}
		}
		log.Debug("found hostnames", zap.Strings("hostnames", hostnames))

		return errors.WithStack(c.RemoteMappingRepository().Create(uint32(port), hostnames...))
	})

	defer func() {
		err := c.RemoteMappingRepository().Delete(uint32(port))
		log.Warn("deleting mappings returned error", zap.Int("port", port), zap.Error(err))
	}()

	return errors.WithStack(eg.Wait())
}
