package cmd

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/takama/daemon"
	"go.uber.org/zap"

	"github.com/srvc/ery/pkg/ery"
	"github.com/srvc/ery/pkg/ery/di"
)

func newCmdDaemon(cfg *ery.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage daemon",
	}

	cmd.AddCommand(newDaemonCmds(cfg)...)

	return cmd
}

func newDaemonCmds(cfg *ery.Config) (cmds []*cobra.Command) {
	funcs := []struct {
		name string
		desc string
		run  func(d daemon.Daemon) (string, error)
	}{
		{
			name: "install",
			desc: "Register daemon to system",
			run:  func(d daemon.Daemon) (string, error) { return d.Install("start") },
		},
		{
			name: "remove",
			desc: "Unregister daemon to system",
			run:  func(d daemon.Daemon) (string, error) { return d.Remove() },
		},
		{
			name: "start",
			desc: "Start servers as daemon",
			run:  func(d daemon.Daemon) (string, error) { return d.Start() },
		},
		{
			name: "stop",
			desc: "Stop servers daemon",
			run:  func(d daemon.Daemon) (string, error) { return d.Stop() },
		},
		{
			name: "status",
			desc: "Show daemon status",
			run:  func(d daemon.Daemon) (string, error) { return d.Status() },
		},
	}

	log := zap.L().Named("daemon")

	for _, f := range funcs {
		f := f
		cmds = append(cmds, &cobra.Command{
			Use:           f.name,
			Short:         f.desc,
			SilenceErrors: true,
			SilenceUsage:  true,
			RunE: func(*cobra.Command, []string) error {
				app := di.NewDaemonApp(cfg)

				d, err := app.DaemonFactory.Get()
				if err != nil {
					log.Error("failed to init daemon", zap.Error(err))
					return errors.WithStack(err)
				}

				msg, err := f.run(d)
				if err == nil {
					log.Debug(f.name, zap.String("message", msg))
				} else {
					log.Error(f.name, zap.String("message", msg), zap.Error(err))
					fmt.Fprintln(cfg.ErrWriter, err)
				}
				fmt.Fprintln(cfg.OutWriter, msg)
				return errors.WithStack(err)
			},
		})
	}

	return
}
