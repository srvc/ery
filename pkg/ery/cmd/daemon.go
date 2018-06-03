package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/takama/daemon"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/ery/di"
)

func newDaemonCmds(c di.AppComponent) (cmds []*cobra.Command) {
	funcs := []struct {
		name string
		desc string
		run  func(d daemon.Daemon) (string, error)
	}{
		{
			name: "install",
			run:  func(d daemon.Daemon) (string, error) { return d.Install("start") },
		},
		{
			name: "remove",
			run:  func(d daemon.Daemon) (string, error) { return d.Remove() },
		},
		{
			name: "stop",
			run:  func(d daemon.Daemon) (string, error) { return d.Stop() },
		},
		{
			name: "status",
			run:  func(d daemon.Daemon) (string, error) { return d.Status() },
		},
	}

	for _, f := range funcs {
		f := f
		cmds = append(cmds, &cobra.Command{
			Use:           f.name,
			Short:         f.desc,
			SilenceErrors: true,
			SilenceUsage:  true,
			RunE: func(*cobra.Command, []string) error {
				log := zap.L().Named("daemon")

				d, err := c.DaemonFactory().Get()
				if err != nil {
					log.Error("failed to init daemon", zap.Error(err))
					return err
				}

				msg, err := f.run(d)
				if err == nil {
					log.Debug(f.name, zap.String("message", msg))
				} else {
					log.Error(f.name, zap.String("message", msg), zap.Error(err))
					fmt.Fprintln(c.Config().ErrWriter, err)
				}
				fmt.Fprintln(c.Config().OutWriter, msg)
				return err
			},
		})
	}

	return
}
