package cmd

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/ery/pkg/ery/infra/local"
	"github.com/srvc/ery/pkg/ery/infra/mem"
	"github.com/srvc/ery/pkg/server/api"
	"github.com/srvc/ery/pkg/server/dns"
	"github.com/srvc/ery/pkg/server/proxy"
	cliutil "github.com/srvc/ery/pkg/util/cli"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start server",
		RunE: cliutil.CobraRunE(func(ctx context.Context, c *cobra.Command, args []string) error {
			ipPool, err := local.NewIPPool()
			if err != nil {
				return err
			}
			portPool := local.NewPortPool()
			appRepo := mem.NewAppRepository(ipPool, portPool)
			dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
			if err != nil {
				return err
			}
			proxies := proxy.NewManager(dockerClient)
			dns := dns.NewServer(appRepo)
			api := api.NewServer(appRepo, proxies)

			eg, ctx := errgroup.WithContext(ctx)
			eg.Go(func() error { return proxies.Serve(ctx) })
			eg.Go(func() error { return dns.Serve(ctx) })
			eg.Go(func() error { return api.Serve(ctx) })

			return eg.Wait()
		}),
	}

	return cmd
}
