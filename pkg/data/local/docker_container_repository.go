package local

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/filters"
	"github.com/moby/moby/client"
	"github.com/pkg/errors"
	"github.com/srvc/ery/pkg/domain"
	"go.uber.org/zap"
)

// NewDockerContainerRepository creates a new ContainerRepository instance concerned to local docker containers.
func NewDockerContainerRepository() domain.ContainerRepository {
	return &dockerContainerRepository{
		log: zap.L().Named("docker"),
	}
}

type dockerContainerRepository struct {
	log *zap.Logger
}

func (r *dockerContainerRepository) ListenEvent(ctx context.Context) (_evCh <-chan domain.ContainerEvent, _errCh <-chan error) {
	evCh := make(chan domain.ContainerEvent)
	errCh := make(chan error, 1)
	_evCh, _errCh = evCh, errCh

	go func() {
		defer close(evCh)
		defer close(errCh)

		client, err := client.NewEnvClient()
		if err != nil {
			errCh <- errors.WithStack(err)
			return
		}
		defer client.Close()

		dockerEvCh, dockerErrCh := r.listenDockerEvent(ctx, client)

		for {
			select {
			case ev := <-dockerEvCh:
				r.log.Debug("receive event", zap.Any("message", ev))

				switch ev.Action {
				case "start":
					evCh <- *r.handleStart(ctx, client, ev)
				case "die":
					evCh <- *r.handleDie(ctx, client, ev)
				}
			case err := <-dockerErrCh:
				if err != context.Canceled {
					r.log.Warn("receive error", zap.Error(err))
					errCh <- errors.WithStack(err)
				}
				return
			}
		}
	}()

	return
}

func (r *dockerContainerRepository) listenDockerEvent(ctx context.Context, cli client.APIClient) (<-chan events.Message, <-chan error) {
	args := filters.NewArgs()
	args.Add("type", "container")
	for _, a := range []string{"start", "die"} {
		args.Add("event", a)
	}
	return cli.Events(ctx, types.EventsOptions{Filters: args})
}

func (r *dockerContainerRepository) handleStart(ctx context.Context, cli client.APIClient, msg events.Message) (ev *domain.ContainerEvent) {
	ev = &domain.ContainerEvent{
		Type: domain.ContainerEventCreated,
		Container: domain.Container{
			ID:           msg.ID,
			Platform:     domain.ContainerPlatformDocker,
			PortBindings: map[domain.Port][]domain.Port{},
		},
	}

	data, err := cli.ContainerInspect(ctx, msg.ID)
	if err != nil {
		r.log.Warn("failed to inspect container", zap.String("id", msg.ID), zap.Error(err))
		ev.Error = errors.Wrap(err, "failed to inspect container")
		return
	}

	ev.Container.Name = strings.TrimPrefix(data.Name, "/")
	ev.Container.Labels = data.Config.Labels

	for k, v := range data.NetworkSettings.Ports {
		if v == nil {
			continue
		}

		cport := domain.Port(k.Int())

		for _, b := range v {
			hport, err := domain.PortFromString(b.HostPort)
			if err != nil {
				r.log.Warn("failed to find the port number", zap.String("id", msg.ID), zap.Any("binding", b), zap.Error(err))
				continue
			}
			ev.Container.PortBindings[cport] = append(ev.Container.PortBindings[cport], hport)
		}
	}

	return ev
}

func (r *dockerContainerRepository) handleDie(ctx context.Context, cli client.APIClient, msg events.Message) (ev *domain.ContainerEvent) {
	ev = &domain.ContainerEvent{
		Type: domain.ContainerEventDestroyed,
		Container: domain.Container{
			ID:       msg.ID,
			Platform: domain.ContainerPlatformDocker,
		},
	}
	return
}
