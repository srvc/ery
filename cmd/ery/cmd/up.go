package cmd

import (
	"context"
	"fmt"
	"sync"

	"github.com/izumin5210/clig/pkg/clib"
	"github.com/spf13/cobra"
	"github.com/srvc/ery/cmd/ery/cmd/up"
	cliutil "github.com/srvc/ery/pkg/util/cli"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/srvc/ery"
	api_pb "github.com/srvc/ery/api"
)

func newUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Up server",
		Args:  cobra.ExactArgs(1),
		RunE: cliutil.CobraRunE(func(ctx context.Context, c *cobra.Command, args []string) error {
			fs := ery.NewFs()
			uFs, err := ery.NewUnionFs(fs)
			if err != nil {
				return err
			}
			viper := ery.NewViper(uFs)
			cfg, err := ery.NewConfig(viper)
			if err != nil {
				return err
			}

			conn, err := grpc.DialContext(ctx, "api.ery.local:80", grpc.WithInsecure())
			if err != nil {
				return err
			}
			appAPI := api_pb.NewAppServiceClient(conn)

			proj := cfg.FindProject(args[0])
			if proj == nil {
				return fmt.Errorf("Project %q was not found", args[0])
			}

			io := &clib.IOContainer{
				InR:  c.InOrStdin(),
				OutW: c.OutOrStdout(),
				ErrW: c.OutOrStderr(),
			}

			runner := up.New(
				appAPI,
				up.NewLocalRunnerFactory(
					cfg.Root,
					io,
				),
			)

			var wg sync.WaitGroup
			for _, app := range proj.Apps {
				app := app
				wg.Add(1)
				go func() {
					defer wg.Done()

					err := runner.Run(ctx, app)
					if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
						zap.L().Error("unexpected exit app", zap.Any("app", app), zap.Error(err))
					}
					// TODO: auto restart?
				}()
			}

			wg.Wait()

			return nil
		}),
	}

	return cmd
}
