package cmd

import (
	"time"

	"github.com/spf13/cobra"
	"github.com/srvc/ery/pkg/ery/di"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewEryCommand creates a new cobra.Command instance.
func NewEryCommand(c di.AppComponent) *cobra.Command {
	cmd := &cobra.Command{
		Use: "ery",
	}

	var (
		verbose bool
	)

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
