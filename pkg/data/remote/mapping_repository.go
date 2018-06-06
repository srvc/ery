package remote

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"

	"github.com/srvc/ery/pkg/domain"
)

// NewMappingRepository creates a new MappingRepository instance that can access remote data.
func NewMappingRepository(baseURL string) domain.MappingRepository {
	m := &mappingRepositoryImpl{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		client:  &http.Client{},
	}
	return m
}

type mappingRepositoryImpl struct {
	baseURL string
	client  *http.Client
}

func (m *mappingRepositoryImpl) List() ([]*domain.Mapping, error) {
	resp, err := m.client.Get(m.baseURL + "/mappings")
	if err != nil {
		return nil, err
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	body := struct {
		Mappings []struct {
			IP        net.IP   `json:"ip"`
			Port      uint32   `json:"prot"`
			Hostnames []string `json:"hostnames"`
		} `json:"mappings"`
	}{}

	err = json.Unmarshal(data, &body)
	if err != nil {
		return nil, err
	}

	mappings := make([]*domain.Mapping, 0, len(body.Mappings))
	for _, m := range body.Mappings {
		mappings = append(mappings, &domain.Mapping{
			Addr: domain.Addr{
				IP:   m.IP,
				Port: m.Port,
			},
			Hostnames: m.Hostnames,
		})
	}

	return mappings, nil
}

func (m *mappingRepositoryImpl) GetBySourceHost(host string) (targetHost string, err error) {
	err = errors.New("remote.MappingRepository.GetBySourceHost() has not been implemented yet")
	return
}

func (m *mappingRepositoryImpl) Create(port uint32, hosts ...string) error {
	data, err := json.Marshal(struct {
		Port      uint32   `json:"port"`
		Hostnames []string `json:"hostnames"`
	}{
		Port:      port,
		Hostnames: hosts,
	})
	if err != nil {
		return err
	}

	_, err = m.client.Post(m.baseURL+"/mappings", "application/json", bytes.NewBuffer(data))
	return err
}

func (m *mappingRepositoryImpl) Delete(port uint32) error {
	req, err := http.NewRequest("DELETE", fmt.Sprintf("%s/mappings/%d", m.baseURL, port), nil)
	if err != nil {
		return err
	}

	_, err = m.client.Do(req)
	return err
}

func (m *mappingRepositoryImpl) DeleteAll() error {
	return errors.New("remote.MappingRepository.DeleteAll() has not been implemented yet")
}
