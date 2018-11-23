package remote

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"

	"github.com/pkg/errors"

	"github.com/srvc/ery/pkg/domain"
)

// NewMappingRepository creates a new MappingRepository instance that can access remote data.
func NewMappingRepository(url *url.URL, client *http.Client) domain.MappingRepository {
	m := &mappingRepositoryImpl{
		baseURL: url,
		client:  client,
	}
	return m
}

type mappingRepositoryImpl struct {
	baseURL *url.URL
	client  *http.Client
}

func (m *mappingRepositoryImpl) List(ctx context.Context) ([]*domain.Mapping, error) {
	req, err := http.NewRequest("GET", m.baseURL.String()+"/mappings", nil)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	resp, err := m.client.Do(req.WithContext(ctx))
	if err != nil {
		return nil, errors.WithStack(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	body := struct {
		Mappings []*domain.Mapping `json:"mappings"`
	}{}

	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return body.Mappings, nil
}

func (m *mappingRepositoryImpl) LookupIP(ctx context.Context, host string) (net.IP, bool) {
	panic("not implemented")
}

func (m *mappingRepositoryImpl) MapAddr(ctx context.Context, addr domain.Addr) (domain.Addr, error) {
	panic("not implemented")
}

func (m *mappingRepositoryImpl) Create(ctx context.Context, addr domain.Addr, rPort domain.Port) (domain.Addr, error) {
	var rAddr domain.Addr

	if rPort != 0 {
		return rAddr, errors.Errorf("cannot specify rPort: %v, %d", addr, rPort)
	}

	data, err := json.Marshal(addr)
	if err != nil {
		return rAddr, errors.WithStack(err)
	}

	req, err := http.NewRequest("POST", m.baseURL.String()+"/mappings", bytes.NewBuffer(data))
	if err != nil {
		return rAddr, errors.WithStack(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := m.client.Do(req.WithContext(ctx))
	if err != nil {
		return rAddr, errors.WithStack(err)
	}
	defer resp.Body.Close()

	err = json.NewDecoder(resp.Body).Decode(&rAddr)

	return rAddr, errors.WithStack(err)
}

func (m *mappingRepositoryImpl) DeleteByHost(ctx context.Context, host string) error {
	req, err := http.NewRequest("DELETE", m.baseURL.String()+"/mappings/"+host, nil)
	if err != nil {
		return errors.WithStack(err)
	}

	_, err = m.client.Do(req.WithContext(ctx))
	return errors.WithStack(err)
}

func (m *mappingRepositoryImpl) ListenEvent(ctx context.Context) (<-chan domain.MappingEvent, <-chan error) {
	evCh := make(chan domain.MappingEvent)
	errCh := make(chan error, 1)
	errCh <- errors.New("remote.MappingRepository.HasHost() has not been implemented yet")
	return evCh, errCh
}
