package daemon

import (
	"sync"

	"github.com/takama/daemon"
)

// Factory is a factory object for daemon.
type Factory interface {
	Get() (daemon.Daemon, error)
}

// NewFactory creates a factory instance.
func NewFactory(name, description string) Factory {
	return &factoryImpl{
		name:        name,
		description: description,
	}
}

type factoryImpl struct {
	name, description string
	daemon            daemon.Daemon
	initOnce          sync.Once
}

func (f *factoryImpl) Get() (daemon.Daemon, error) {
	var err error
	f.initOnce.Do(func() {
		f.daemon, err = daemon.New(f.name, f.description)
	})
	return f.daemon, err
}
