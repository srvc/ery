package domain

// MappingRepository is an interface for accessing <hostname>-<port> mappings.
type MappingRepository interface {
	List() ([]*Mapping, error)
	GetBySourceHost(host string) (targetHost string, err error)
	Create(port uint32, host string) error
	Delete(port uint32) error
	DeleteAll() error
}
