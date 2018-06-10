package local

import (
	"context"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/srvc/ery/pkg/domain"
)

func Test_MappingRepository(t *testing.T) {
	var (
		m1 = &domain.Mapping{
			Host: "web1.services.local",
			PortAddrMap: domain.PortAddrMap{
				80: domain.LocalAddr(8001),
			},
		}
		m2 = &domain.Mapping{
			Host: "web2.services.local",
			PortAddrMap: domain.PortAddrMap{
				80: domain.LocalAddr(8002),
			},
		}
		m3 = &domain.Mapping{
			Host: "web3.services.local",
			PortAddrMap: domain.PortAddrMap{
				80: domain.LocalAddr(8003),
			},
		}
		m4 = &domain.Mapping{
			Host: "web3.services.local",
			PortAddrMap: domain.PortAddrMap{
				80:   domain.LocalAddr(8003),
				5432: domain.LocalAddr(8004),
			},
		}
	)

	assertErr := func(t *testing.T, err error) {
		t.Helper()
		if err == nil {
			t.Error("returned nil, want an error")
		}
	}

	assertNoErr := func(t *testing.T, err error) {
		t.Helper()
		if err != nil {
			t.Errorf("returned %v, want nil", err)
		}
	}

	assertFound := func(t *testing.T, m domain.MappingRepository, in, want domain.Addr) {
		t.Helper()
		got, err := m.MapAddr(context.TODO(), in)
		assertNoErr(t, err)
		if got != want {
			t.Errorf("MapAddr(%v) returned %v, want %v", in, got, want)
		}
	}

	assertNotFound := func(t *testing.T, m domain.MappingRepository, in domain.Addr) {
		t.Helper()
		_, err := m.MapAddr(context.TODO(), in)
		assertErr(t, err)
	}

	repo := NewMappingRepository()

	// Test Add
	assertNoErr(t, repo.Create(context.TODO(), m1))
	assertFound(t, repo, domain.HTTPAddr(m1.Host), m1.Map(80))
	assertNotFound(t, repo, domain.NewAddr(m1.Host, 8000))
	assertNotFound(t, repo, domain.HTTPAddr(m2.Host))

	has, err := repo.HasHost(context.TODO(), m1.Host)
	assertNoErr(t, err)
	if got, want := has, true; got != want {
		t.Errorf("HasHost(%q) returned %t, want %t", m1.Host, got, want)
	}
	has, err = repo.HasHost(context.TODO(), m2.Host)
	assertNoErr(t, err)
	if got, want := has, false; got != want {
		t.Errorf("HasHost(%q) returned %t, want %t", m2.Host, got, want)
	}

	assertNoErr(t, repo.Create(context.TODO(), m2))
	assertFound(t, repo, domain.HTTPAddr(m2.Host), m2.Map(80))
	assertNoErr(t, repo.Create(context.TODO(), m3))
	assertFound(t, repo, domain.HTTPAddr(m3.Host), m3.Map(80))

	// Test Failed to create
	assertErr(t, repo.Create(context.TODO(), m4))
	assertFound(t, repo, domain.HTTPAddr(m3.Host), m3.Map(80))

	// Test List
	got, err := repo.List(context.TODO())
	assertNoErr(t, err)
	want := []*domain.Mapping{m1, m2, m3}
	diffOpts := []cmp.Option{
		cmpopts.SortSlices(func(m1, m2 *domain.Mapping) bool { return strings.Compare(m1.Host, m2.Host) < 0 }),
	}
	if diff := cmp.Diff(got, want, diffOpts...); diff != "" {
		t.Errorf("Returned list differs: (-got +want)\n%s", diff)
	}

	// Test Remove
	assertNoErr(t, repo.DeleteByHost(context.TODO(), m3.Host))
	assertFound(t, repo, domain.HTTPAddr(m1.Host), m1.Map(80))
	assertNotFound(t, repo, domain.HTTPAddr(m3.Host))

	// Test multiple values
	assertNoErr(t, repo.Create(context.TODO(), m4))
	assertFound(t, repo, domain.HTTPAddr(m4.Host), m4.Map(80))
	assertFound(t, repo, domain.NewAddr(m4.Host, 5432), m4.Map(5432))
}
