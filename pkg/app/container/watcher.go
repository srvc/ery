package container

import (
	"context"
	"strings"
	"sync"

	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/srvc/ery/pkg/domain"
)

// Manager is an interface of proxy servers' manager.
type Watcher interface {
	ListenEvents(context.Context) error
}

// NewWatcher creates a new Watcher instance concerned to containers.
func NewWatcher(
	mappingRepo domain.MappingRepository,
	containerRepos []domain.ContainerRepository,
	tld, labelHostname string,
) Watcher {
	return &watcherImpl{
		mappingRepo:    mappingRepo,
		containerRepos: containerRepos,
		tld:            tld,
		labelHostname:  labelHostname,
		hostsByCID:     new(sync.Map),
		log:            zap.L().Named("watcher"),
	}
}

type watcherImpl struct {
	hostsByCID     *sync.Map
	mappingRepo    domain.MappingRepository
	containerRepos []domain.ContainerRepository
	tld            string
	labelHostname  string
	log            *zap.Logger
}

func (w *watcherImpl) ListenEvents(pctx context.Context) error {
	evCh := make(chan domain.ContainerEvent)
	defer close(evCh)

	eg, ctx := errgroup.WithContext(pctx)
	w.log.Debug("start watching container events")

	// collectors
	for _, containerRepo := range w.containerRepos {
		repo := containerRepo
		eg.Go(func() error {
			origEvCh, origErrCh := repo.ListenEvent(ctx)
			for {
				select {
				case ev := <-origEvCh:
					evCh <- ev
				case err := <-origErrCh:
					return errors.WithStack(err)
				case <-ctx.Done():
					w.log.Debug("stop watching container events", zap.Error(ctx.Err()))
					return errors.WithStack(ctx.Err())
				}
			}
		})
	}

	// processor
	eg.Go(func() error {
		for {
			select {
			case ev := <-evCh:
				w.log.Debug("receive event", zap.Any("message", ev))

				switch ev.Type {
				case domain.ContainerEventCreated:
					w.handleCreated(ctx, ev.Container)
				case domain.ContainerEventDestroyed:
					w.handleDestroyed(ctx, ev.Container)
				}
			case <-ctx.Done():
				w.log.Debug("stop processing container events", zap.Error(ctx.Err()))
				return errors.WithStack(ctx.Err())
			}
		}
	})

	return errors.WithStack(eg.Wait())
}

func (w *watcherImpl) handleCreated(ctx context.Context, c domain.Container) {
	hostnames := []string{}
	for _, n := range c.Networks {
		hostname := strings.Join([]string{c.Name, n.Name, c.Platform.String(), w.tld}, ".")
		hostnames = append(hostnames, hostname)
	}
	if n, ok := c.Labels[w.labelHostname]; ok {
		hostnames = append(hostnames, n)
	}

	w.hostsByCID.Store(c.ID, hostnames)

	for cport, hports := range c.PortBindings {
		for _, hport := range hports {
			for _, host := range hostnames {
				lAddr := domain.Addr{Host: host, Port: cport}
				rAddr, err := w.mappingRepo.Create(ctx, lAddr, hport)
				if err == nil {
					w.log.Info("created a new mapping", zap.Stringer("src_addr", &lAddr), zap.Stringer("dest_addr", &rAddr), zap.String("container_id", c.ID))
				} else {
					w.log.Warn("failed to create a new mapping", zap.Error(err), zap.Stringer("src_addr", &lAddr), zap.Any("dest_port", hport), zap.String("container_id", c.ID))
				}
			}
		}
	}
}

func (w *watcherImpl) handleDestroyed(ctx context.Context, c domain.Container) {
	if v, ok := w.hostsByCID.Load(c.ID); ok {
		if hosts, ok := v.([]string); ok {
			for _, h := range hosts {
				err := w.mappingRepo.DeleteByHost(ctx, h)
				if err != nil {
					w.log.Warn("failed to delete a mapping", zap.Error(err), zap.String("host", h), zap.String("container_id", c.ID))
				}
			}
		}
		w.hostsByCID.Delete(c.ID)
	}
}
