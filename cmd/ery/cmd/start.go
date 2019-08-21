package cmd

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/srvc/ery/pkg/ery/infra/local"
	"github.com/srvc/ery/pkg/ery/infra/mem"
	"github.com/srvc/ery/pkg/server/api"
	"github.com/srvc/ery/pkg/server/dns"
	"golang.org/x/sync/errgroup"
)

func newStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "Start server",
		RunE: func(c *cobra.Command, args []string) error {
			ipPool, err := local.NewIPPool()
			if err != nil {
				return err
			}
			portPool := local.NewPortPool()
			appRepo := mem.NewAppRepository(ipPool, portPool)
			dns := dns.NewServer(appRepo)
			api := api.NewServer(appRepo)

			eg, ctx := errgroup.WithContext(context.Background())

			eg.Go(func() error { return dns.Serve(ctx) })
			eg.Go(func() error { return api.Serve(ctx) })

			return eg.Wait()
		},
	}

	return cmd
}
