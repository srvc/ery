package up

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/cli/cli/compose/loader"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/volume"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/izumin5210/clig/pkg/clib"
	"github.com/izumin5210/execx"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/ery"
	api_pb "github.com/srvc/ery/api"
)

func NewDockerRunnerFactory(
	rootDir string,
	io clib.IO,
	docker *client.Client,
) *DockerRunnerFactory {
	return &DockerRunnerFactory{
		rootDir: rootDir,
		io:      io,
		docker:  docker,
	}
}

type DockerRunnerFactory struct {
	rootDir string
	io      clib.IO
	docker  *client.Client
}

func (f *DockerRunnerFactory) GetRunner(app *ery.App, appPb *api_pb.App) Runner {
	return &DockerRunner{
		DockerRunnerFactory: f,
		app:                 app,
		appPb:               appPb,
		log: zap.L().With(
			zap.String("app_name", app.Name),
			zap.String("app_type", "docker"),
			zap.Any("app", appPb),
		),
	}
}

type DockerRunner struct {
	*DockerRunnerFactory
	app   *ery.App
	appPb *api_pb.App
	log   *zap.Logger
}

func (r *DockerRunner) Run(ctx context.Context) error {
	imageTag := "ery--" + r.app.Name

	err := r.build(ctx, imageTag)
	if err != nil {
		r.log.Error("failed to build image", zap.Error(err))
		return err
	}

	r.docker, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		r.log.Error("failed to create docker client", zap.Error(err))
		return err
	}

	mounts, err := r.createVolumes(ctx, r.app.Docker.Run.Volumes)
	if err != nil {
		r.log.Error("failed to create docker volumes", zap.Error(err))
		return err
	}

	containerID, err := r.createContainer(ctx, imageTag, mounts)
	if err != nil {
		r.log.Error("failed to create docker container", zap.Error(err))
		return err
	}

	eg, ctx := errgroup.WithContext(ctx)

	for _, cmd := range r.app.Docker.Run.Cmd {
		cmd := cmd
		eg.Go(func() error { return r.execCommand(ctx, containerID, cmd) })
	}

	return eg.Wait()
}

func (r *DockerRunner) build(ctx context.Context, imageTag string) error {
	// build image
	args := []string{"build", "--tag", imageTag}
	if f := r.app.Docker.Build.Dockerfile; f != "" {
		args = append(args, "--file", f)
	}
	if t := r.app.Docker.Build.Target; t != "" {
		args = append(args, "--target", t)
	}
	args = append(args, ".")
	cmd := execx.CommandContext(ctx, "docker", args...)
	cmd.Dir = filepath.Join(r.rootDir, r.app.Docker.Path)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "DOCKER_BUILDKIT=1")
	cmd.Stdin = r.io.In()
	cmd.Stdout = r.io.Out()
	cmd.Stderr = r.io.Err()
	r.log.Debug("start building docker image", zap.Strings("command", cmd.Args))
	return cmd.Run()
}

func (r *DockerRunner) createVolumes(ctx context.Context, specs []string) ([]*mount.Mount, error) {
	res := make([]*mount.Mount, len(specs))

	for i, spec := range specs {
		m, err := r.createVolume(ctx, spec)
		if err != nil {
			return nil, err
		}
		res[i] = m
	}

	return res, nil
}

func (r *DockerRunner) createVolume(ctx context.Context, spec string) (*mount.Mount, error) {
	vol, err := loader.ParseVolume(spec)
	if err != nil {
		r.log.Error("failed to parse volume", zap.Error(err))
		return nil, err
	}
	switch vol.Type {
	case "volume":
		volName := strings.Join([]string{"ery", r.app.Name, vol.Source}, "--")
		_, err := r.docker.VolumeInspect(ctx, volName)
		if client.IsErrNotFound(err) {
			_, err := r.docker.VolumeCreate(ctx, volume.VolumeCreateBody{
				Driver:     "local",
				DriverOpts: map[string]string{},
				Labels:     map[string]string{},
				Name:       volName,
			})
			if err != nil {
				r.log.Error("failed to create volume", zap.Error(err))
				return nil, err
			}
		} else if err != nil {
			r.log.Error("failed to inspectvolume", zap.Error(err))
			return nil, err
		}
		vol.Source = volName
	case "bind":
		src := vol.Source
		switch src[0] {
		case '.':
			src = filepath.Join(r.rootDir, r.app.Docker.Path, src[1:])
		case '~':
			src = filepath.Join(os.Getenv("HOME"), string(src[1:])) // TODO: extract homedir
		}
		vol.Source = src
	}
	m := &mount.Mount{
		Type:        mount.Type(vol.Type),
		Source:      vol.Source,
		Target:      vol.Target,
		ReadOnly:    vol.ReadOnly,
		Consistency: mount.Consistency(vol.Consistency),
	}
	if b := vol.Bind; b != nil {
		m.BindOptions = &mount.BindOptions{
			Propagation: mount.Propagation(b.Propagation),
		}
	}
	if v := vol.Volume; v != nil {
		m.VolumeOptions = &mount.VolumeOptions{
			NoCopy: v.NoCopy,
		}
	}
	if f := vol.Tmpfs; f != nil {
		m.TmpfsOptions = &mount.TmpfsOptions{
			SizeBytes: f.Size,
		}
	}

	return m, nil
}

func (r *DockerRunner) createContainer(ctx context.Context, imageTag string, mounts []*mount.Mount) (string, error) {
	containers, err := r.docker.ContainerList(ctx, types.ContainerListOptions{
		Quiet:   true,
		Filters: filters.NewArgs(filters.KeyValuePair{Key: "name", Value: r.appPb.GetHostname()}),
	})
	if err != nil {
		return "", err
	}
	if len(containers) > 0 {
		r.log.Debug("container has existed")
		return containers[0].ID, nil
	}

	cfg := &container.Config{
		Hostname:     r.appPb.GetHostname(),
		Domainname:   r.appPb.GetHostname(),
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"tail", "-f", "/dev/null"},
		Image:        imageTag,
		ExposedPorts: nat.PortSet{},
		Labels: map[string]string{
			"ery-app-id":   r.appPb.GetAppId(),
			"ery-app-name": r.appPb.GetName(),
		},
	}
	hostCfg := &container.HostConfig{
		NetworkMode:  container.NetworkMode("srvc/ery"),
		PortBindings: nat.PortMap{},
		AutoRemove:   true,
	}
	for _, m := range mounts {
		hostCfg.Mounts = append(hostCfg.Mounts, *m)
	}
	nwCfg := &network.NetworkingConfig{}
	for _, p := range r.appPb.GetPorts() {
		ePort := nat.Port(fmt.Sprintf("%d/%s", p.GetExposedPort(), strings.ToLower(p.GetNetwork().String())))
		cfg.ExposedPorts[ePort] = struct{}{}
		hostCfg.PortBindings[ePort] = append(hostCfg.PortBindings[ePort], nat.PortBinding{
			HostIP:   r.appPb.GetIp(),
			HostPort: fmt.Sprintf("%d/%s", p.GetExposedPort(), strings.ToLower(p.GetNetwork().String())),
		})
	}
	resp, err := r.docker.ContainerCreate(ctx, cfg, hostCfg, nwCfg, r.app.Hostname)
	if err != nil {
		r.log.Error("failed to create container", zap.Error(err))
		return "", err
	}
	containerID := resp.ID

	err = r.docker.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	if err != nil {
		r.log.Error("failed to start container", zap.Error(err))
		return "", err
	}

	return containerID, nil
}

func (r *DockerRunner) execCommand(ctx context.Context, containerID string, cmd []string) error {
	resp, err := r.docker.ContainerExecCreate(ctx, containerID, types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          cmd,
	})
	if err != nil {
		r.log.Error("failed to create exec", zap.Error(err))
		return err
	}

	attachResp, err := r.docker.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{})
	if err != nil {
		r.log.Error("failed to attach exec", zap.Error(err))
		return err
	}
	defer attachResp.Close()

	io.Copy(r.io.Out(), attachResp.Reader)

	return err
}
