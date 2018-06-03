package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/srvc/ery/pkg/ery/di"
)

func newCmdVersion(c di.AppComponent) *cobra.Command {
	return &cobra.Command{
		Use:           "version",
		Short:         "Print version information",
		Long:          "Print version information",
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, _ []string) {
			cfg := c.Config()
			fmt.Fprintf(cfg.OutWriter, "ery %s %s (%s %s)\n", cfg.Version, cfg.ReleaseType, cfg.BuildDate, cfg.Revision)
		},
	}
}
