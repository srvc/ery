package local

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/util/netutil"
)

func Test_MappingRepository(t *testing.T) {
	var (
		ip                  = net.IPv4(127, 0, 0, 1)
		port1, port2 uint32 = 8081, 8082
		host1               = "foo.services.local"
		host2               = "bar.services.local"
		host3               = "baz.services.local"
	)

	assertNoErr := func(t *testing.T, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
	}

	assertFound := func(t *testing.T, m domain.MappingRepository, inputHost string, wantPort uint32) {
		t.Helper()
		gotHost, err := m.GetBySourceHost(inputHost)
		assertNoErr(t, err)
		if got, want := gotHost, netutil.HostAndPort(ip.String(), wantPort); got != want {
			t.Errorf("Lookup(%q) returned %q, want %q", inputHost, got, want)
		}
	}

	assertNotFound := func(t *testing.T, m domain.MappingRepository, inputHost string) {
		t.Helper()
		gotHost, err := m.GetBySourceHost(inputHost)
		assertNoErr(t, err)
		if got, want := gotHost, ""; got != want {
			t.Errorf("Lookup(%q) returned %q, want %q", inputHost, got, want)
		}
	}

	repo := NewMappingRepository(ip)

	// Test Add
	assertNoErr(t, repo.Create(port1, host1))
	assertFound(t, repo, host1, port1)
	assertNotFound(t, repo, host2)
	assertNotFound(t, repo, host3)

	assertNoErr(t, repo.Create(port2, host2))
	assertFound(t, repo, host2, port2)

	assertNoErr(t, repo.Create(port1, host3))
	assertFound(t, repo, host1, port1)
	assertFound(t, repo, host3, port1)

	// Test List
	got, err := repo.List()
	assertNoErr(t, err)
	want := []*domain.Mapping{
		{Addr: domain.Addr{IP: ip, Port: port1}, Hostnames: []string{host1, host3}},
		{Addr: domain.Addr{IP: ip, Port: port2}, Hostnames: []string{host2}},
	}
	diffOpts := []cmp.Option{
		cmpopts.SortSlices(func(m1, m2 *domain.Mapping) bool { return m1.Port < m2.Port }),
	}
	if diff := cmp.Diff(got, want, diffOpts...); diff != "" {
		t.Errorf("Returned list differs: (-got +want)\n%s", diff)
	}

	// Test Remove
	assertNoErr(t, repo.Delete(port1))
	assertFound(t, repo, host2, port2)
	assertNotFound(t, repo, host1)
	assertNotFound(t, repo, host3)

	// Test Clear
	assertNoErr(t, repo.DeleteAll())
	assertNotFound(t, repo, host2)
}
