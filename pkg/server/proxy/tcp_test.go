package proxy_test

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/srvc/ery/pkg/ery/domain"
	"github.com/srvc/ery/pkg/server/proxy"
	netutil "github.com/srvc/ery/pkg/util/net"
)

func TestTCPServer(t *testing.T) {
	svr := httptest.NewUnstartedServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodGet && r.URL.Path == "/ping" {
				w.Write([]byte("pong"))
			} else {
				w.WriteHeader(http.StatusNotFound)
				w.Write([]byte("not found"))
			}
		}),
	)
	defer svr.Close()

	port, err := netutil.GetFreePort()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	proxy := proxy.NewTCPServer(
		&domain.Addr{IP: "127.0.0.1", Port: domain.Port(port)},
		&domain.Addr{IP: "127.0.0.1", Port: domain.Port(svr.Listener.Addr().(*net.TCPAddr).Port)},
	)

	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := proxy.Serve(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	}()

	svr.Start()

	time.Sleep(10 * time.Millisecond) // wait for starting servers

	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/ping", port))
	if err != nil {
		t.Fatalf("failed to request: %v", err)
	}

	if got, want := resp.StatusCode, http.StatusOK; got != want {
		t.Errorf("Response status code was %d, want %d", got, want)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()

	if got, want := string(data), "pong"; got != want {
		t.Errorf("Response body was %q, want %q", got, want)
	}
}
