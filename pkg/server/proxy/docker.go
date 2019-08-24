package proxy

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	api_pb "github.com/srvc/ery/api"
	logutil "github.com/srvc/ery/pkg/util/log"
)

type DockerServer struct {
	client *client.Client
	app    *api_pb.App
	log    *zap.Logger
}

func NewDockerServer(
	client *client.Client,
	app *api_pb.App,
) *DockerServer {
	return &DockerServer{
		client: client,
		app:    app,
		log:    zap.L().Named("proxy").Named("docker").With(logutil.Proto("app", app)),
	}
}

func (s *DockerServer) Serve(ctx context.Context) error {
	// TODO: pull srvc/ery-proxy
	// TODO: create network

	cfg := &container.Config{
		Hostname:     s.app.GetHostname(),
		Domainname:   s.app.GetHostname(),
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"tail", "-f", "/dev/null"},
		Image:        "srvc/ery-proxy",
		ExposedPorts: nat.PortSet{},
		Labels: map[string]string{
			"ery-app-id":   s.app.GetAppId(),
			"ery-app-name": s.app.GetName(),
		},
	}
	hostCfg := &container.HostConfig{
		NetworkMode:  container.NetworkMode("srvc/ery"),
		PortBindings: nat.PortMap{},
		AutoRemove:   true,
	}
	nwCfg := &network.NetworkingConfig{}

	for _, p := range s.app.GetPorts() {
		ePort := nat.Port(fmt.Sprintf("%d/%s", p.GetExposedPort(), strings.ToLower(p.GetNetwork().String())))
		cfg.ExposedPorts[ePort] = struct{}{}
		hostCfg.PortBindings[ePort] = append(hostCfg.PortBindings[ePort], nat.PortBinding{
			HostIP:   s.app.GetIp(),
			HostPort: fmt.Sprintf("%d/%s", p.GetExposedPort(), strings.ToLower(p.GetNetwork().String())),
		})
	}

	resp, err := s.client.ContainerCreate(ctx, cfg, hostCfg, nwCfg, s.app.GetHostname())
	if err != nil {
		return err
	}
	containerID := resp.ID
	s.log.Debug("ery-proxy container has started", zap.String("container_id", containerID))

	err = s.client.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)

	for _, p := range s.app.GetPorts() {
		p := p
		eg.Go(func() error {
			cmd := []string{
				"ery-proxy",
				"--network", p.GetNetwork().String(),
				"--src-addr", fmt.Sprintf(":%d", p.GetExposedPort()),
				"--dest-addr", fmt.Sprintf("host.docker.internal:%d", p.GetInternalPort()),
			}
			resp, err := s.client.ContainerExecCreate(ctx, containerID, types.ExecConfig{
				AttachStderr: true,
				AttachStdout: true,
				Cmd:          cmd,
			})
			if err != nil {
				return err
			}
			s.log.Debug("execute proxy", zap.Strings("cmd", cmd), logutil.Proto("port", p))
			err = s.client.ContainerExecStart(ctx, resp.ID, types.ExecStartCheck{})
			if err != nil {
				return err
			}
			return nil
		})
	}

	eg.Go(func() error {
		<-ctx.Done()
		to := 30 * time.Second
		err := s.client.ContainerStop(
			context.Background(),
			containerID,
			&to,
		)
		if err != nil {
			s.log.Warn("failed to stop the container", zap.Error(err))
			return err
		}
		return nil
	})

	return eg.Wait()
}
