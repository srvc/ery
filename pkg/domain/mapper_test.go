package domain

import (
	"net"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/srvc/ery/pkg/util/netutil"
)

func Test_Mapper(t *testing.T) {
	var (
		ip                  = net.IPv4(127, 0, 0, 1)
		port1, port2 uint32 = 8081, 8082
		host1               = "foo.services.local"
		host2               = "bar.services.local"
		host3               = "baz.services.local"
	)

	assertFound := func(t *testing.T, m Mapper, inputHost string, wantPort uint32) {
		t.Helper()
		gotHost, ok := m.Lookup(inputHost)
		if got, want := ok, true; got != want {
			t.Errorf("Lookup(%q) returned %t, want %t", inputHost, got, want)
		}
		if got, want := gotHost, netutil.HostAndPort(ip.String(), wantPort); got != want {
			t.Errorf("Lookup(%q) returned %q, want %q", inputHost, got, want)
		}
	}

	assertNotFound := func(t *testing.T, m Mapper, inputHost string) {
		t.Helper()
		gotHost, ok := m.Lookup(inputHost)
		if got, want := ok, false; got != want {
			t.Errorf("Lookup(%q) returned %t, want %t", inputHost, got, want)
		}
		if got, want := gotHost, ""; got != want {
			t.Errorf("Lookup(%q) returned %q, want %q", inputHost, got, want)
		}
	}

	m := NewMapper(ip)

	// Test Add
	m.Add(port1, host1)
	assertFound(t, m, host1, port1)
	assertNotFound(t, m, host2)
	assertNotFound(t, m, host3)

	m.Add(port2, host2)
	assertFound(t, m, host2, port2)

	m.Add(port1, host3)
	assertFound(t, m, host1, port1)
	assertFound(t, m, host3, port1)

	// Test List
	got := m.List()
	want := []*Mapping{
		{Addr: Addr{IP: ip, Port: port1}, Hostnames: []string{host1, host3}},
		{Addr: Addr{IP: ip, Port: port2}, Hostnames: []string{host2}},
	}
	diffOpts := []cmp.Option{
		cmpopts.SortSlices(func(m1, m2 *Mapping) bool { return m1.Port < m2.Port }),
	}
	if diff := cmp.Diff(got, want, diffOpts...); diff != "" {
		t.Errorf("Returned list differs: (-got +want)\n%s", diff)
	}

	// Test Remove
	m.Remove(port1)
	assertFound(t, m, host2, port2)
	assertNotFound(t, m, host1)
	assertNotFound(t, m, host3)

	// Test Clear
	m.Clear()
	assertNotFound(t, m, host2)
}