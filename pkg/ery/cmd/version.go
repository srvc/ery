package cmd

import (
	"bytes"
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
			buf := bytes.NewBufferString("ery " + cfg.Version)
			if cfg.Revision != "" && cfg.BuildDate != "" {
				buf.WriteString(" (" + cfg.BuildDate + " " + cfg.Revision + ")")
			}
			fmt.Fprintln(cfg.OutWriter, buf.String())
		},
	}
}
