package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/takama/daemon"
	"go.uber.org/zap"
)

func newDaemonCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "daemon",
		Short: "Manage daemon",
	}

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
			run:  (daemon.Daemon).Remove,
		},
		{
			name: "start",
			desc: "Start servers as daemon",
			run:  (daemon.Daemon).Start,
		},
		{
			name: "stop",
			desc: "Stop servers daemon",
			run:  (daemon.Daemon).Stop,
		},
		{
			name: "status",
			desc: "Show daemon status",
			run:  (daemon.Daemon).Status,
		},
	}

	log := zap.L().Named("daemon")
	cmds := make([]*cobra.Command, len(funcs))

	for i, f := range funcs {
		f := f
		cmds[i] = &cobra.Command{
			Use:           f.name,
			Short:         f.desc,
			SilenceErrors: true,
			SilenceUsage:  true,
			RunE: func(cmd *cobra.Command, _ []string) error {
				d, err := daemon.New("ery", "Discover services in local")
				if err != nil {
					log.Error("failed to init daemon", zap.Error(err))
					return err
				}

				msg, err := f.run(d)
				if err == nil {
					log.Debug(f.name, zap.String("message", msg))
				} else {
					log.Error(f.name, zap.String("message", msg), zap.Error(err))
					fmt.Fprintln(cmd.ErrOrStderr(), err)
				}
				fmt.Fprintln(cmd.OutOrStdout(), msg)
				return err
			},
		}
	}

	cmd.AddCommand(cmds...)

	return cmd
}
