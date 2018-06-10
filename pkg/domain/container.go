package domain

// Container contains meta data of a container.
type Container struct {
	ID           string
	Name         string
	Platform     ContainerPlatform
	Labels       map[string]string
	Networks     []ContainerNetwork
	PortBindings map[Port][]Port
}

// ContainerNetwork contains meta data of a container network.
type ContainerNetwork struct {
	Name string
}

// ContainerPlatform represents each container implementation, such as "docker".
type ContainerPlatform int

// Enum values of ContainerPlatform.
const (
	ContainerPlatformUnknown ContainerPlatform = iota
	ContainerPlatformDocker
)

var (
	nameByContainerPlatform = map[ContainerPlatform]string{
		ContainerPlatformDocker: "docker",
	}
)

func (p ContainerPlatform) String() string {
	n, ok := nameByContainerPlatform[p]
	if !ok {
		n = "unknown"
	}
	return n
}
