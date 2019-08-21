package cmd

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
	"google.golang.org/grpc"

	api_pb "github.com/srvc/ery/api"
)

func newPsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ps",
		Short: "List apps",
		RunE: func(c *cobra.Command, args []string) error {
			ctx := context.Background()
			conn, err := grpc.DialContext(ctx, "api.ery.local:80", grpc.WithInsecure())
			if err != nil {
				return err
			}
			appAPI := api_pb.NewAppServiceClient(conn)
			resp, err := appAPI.ListApps(ctx, new(api_pb.ListAppsRequest))
			if err != nil {
				return err
			}

			for _, app := range resp.GetApps() {
				fmt.Fprintf(c.OutOrStdout(), "%s\t%s\t%s -> %s\n", app.GetAppId()[:7], app.GetName(), app.GetHostname(), app.GetIp())
			}
			return nil
		},
	}

	return cmd
}
