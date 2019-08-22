package cmd

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/izumin5210/execx"
	"github.com/spf13/cobra"
	"github.com/srvc/ery"
	api_pb "github.com/srvc/ery/api"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func newUpCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "up",
		Short: "Up server",
		Args:  cobra.ExactArgs(1),
		RunE: func(c *cobra.Command, args []string) error {
			ctx := context.Background()

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

			var wg sync.WaitGroup
			for _, app := range proj.Apps {
				app := app
				wg.Add(1)
				go func() {
					defer wg.Done()
					log := zap.L().With(zap.String("app_name", app.Name))

					appPb := &api_pb.App{
						Name:     app.Name,
						Hostname: app.Hostname,
					}
					switch {
					case app.Local != nil:
						appPb.Type = api_pb.App_TYPE_LOCAL
						for name, port := range app.Local.PortEnv {
							appPb.Ports = append(appPb.Ports, &api_pb.App_Port{
								Network:     api_pb.App_Port_TCP, // TODO
								ExposedPort: uint32(port),
								Env:         name,
							})
						}
					case app.Docker != nil:
						appPb.Type = api_pb.App_TYPE_DOCKER
						// TODO: not yet supported
					case app.Kubernetes != nil:
						appPb.Type = api_pb.App_TYPE_KUBERNETES
						// TODO: not yet supported
					}

					resp, err := appAPI.CreateApp(ctx, &api_pb.CreateAppRequest{App: appPb})
					if err != nil {
						log.Error("failed to register app", zap.Error(err))
						return
					}
					appPb = resp

					log = log.With(zap.Any("app", appPb))

					switch {
					case app.Local != nil:
						cmd := execx.CommandContext(ctx, app.Local.Cmd[0], app.Local.Cmd[1:]...)
						cmd.Dir = filepath.Join(cfg.Root, app.Local.Path)
						cmd.Env = os.Environ()
						for _, port := range appPb.Ports {
							if port.Env != "" {
								cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%d", port.Env, port.InternalPort))
							}
						}
						cmd.Stdin = c.InOrStdin()
						cmd.Stdout = c.OutOrStdout()
						cmd.Stderr = c.ErrOrStderr()
						log.Info("start")
						err = cmd.Run()
						if err != nil {
							log.Warn("shutdown", zap.Error(err))
						}

					case app.Docker != nil:
						// TODO: not yet supported
					case app.Kubernetes != nil:
						// TODO: not yet supported
					}
				}()
			}

			wg.Wait()

			return nil
		},
	}

	return cmd
}
