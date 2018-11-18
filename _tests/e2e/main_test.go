package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/srvc/ery/pkg/domain"
	"github.com/srvc/ery/pkg/util/netutil"
)

func TestEry(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	wd := filepath.Dir(file)

	ctx := context.Background()

	ery := newEry(t, filepath.Join(wd, "..", "..", "bin", "ery"))

	eryCmd := ery.Command(ctx, "start", "-v")
	checkErr(t, eryCmd.Start())
	defer func() {
		eryCmd.Process.Signal(os.Interrupt)
		eryCmd.Wait()
	}()

	webDir := filepath.Join(wd, "web")
	tmpDir := filepath.Join(wd, "tmp")
	localIP := netutil.LocalIP()

	httpImageName := "srvc_ery_e2e_testing_http"
	checkErr(t, exec.CommandContext(ctx, "docker", "build", "-t="+httpImageName, webDir).Run())

	createWorkspace := func(t *testing.T, name string) (dir string, closer func()) {
		t.Helper()
		dir = filepath.Join(tmpDir, name)
		closer = func() {
			err := os.RemoveAll(dir)
			if err != nil {
				t.Logf("failed to remove %s: %v", dir, err)
			}
		}
		checkErr(t, os.MkdirAll(dir, 0755))
		checkErr(t, ioutil.WriteFile(filepath.Join(dir, ".ery.toml"), []byte(`hostname = "`+name+`.services.local"`+"\n"), 0644))
		return
	}

	startServerOnLocal := func(t *testing.T, name string) func() {
		t.Helper()
		dir, removeWorkspace := createWorkspace(t, name)
		cmd := ery.Command(ctx, "go", "run", webDir)
		cmd.Dir = dir
		checkErr(t, cmd.Start())
		return func() {
			cmd.Process.Signal(os.Interrupt)
			err := cmd.Wait()
			if err != nil {
				t.Logf("failed to stop server: %v", err)
			}
			removeWorkspace()
		}
	}

	startServerOnDocker := func(t *testing.T, name string) func() {
		t.Helper()
		port := getFreePort(t)
		containerName := httpImageName + "__" + name
		cmd := exec.CommandContext(ctx, "docker",
			"run",
			"--rm",
			"--detach",
			fmt.Sprintf("--name=%s", containerName),
			fmt.Sprintf("--env=PORT=%d", ery.proxyPort),
			fmt.Sprintf("--env=DNS_PORT=%d", ery.dnsPort),
			fmt.Sprintf("-p=%d:%d", port, ery.proxyPort),
			fmt.Sprintf("--dns=%s", localIP),
			fmt.Sprintf("--label=tools.srvc.ery.hostname=%s.services.local", name),
			httpImageName,
			"go", "run", ".",
		)
		checkCmd(t, cmd)
		return func() {
			checkCmd(t, exec.Command("docker", "stop", containerName))
		}
	}

	getURL := func(name string) string {
		return fmt.Sprintf("http://%s.services.local:%d", name, ery.proxyPort)
	}

	time.Sleep(5 * time.Second)

	checkCmd(t, ery.Command(ctx, "ps"))

	t.Run("local http server", func(t *testing.T) {
		defer startServerOnLocal(t, "local")()

		time.Sleep(5 * time.Second)

		cli := ery.HTTPClient()

		resp, err := cli.Get(getURL("local") + "/ping")
		checkErr(t, err)
		data, err := ioutil.ReadAll(resp.Body)
		checkErr(t, err)

		if got, want := resp.StatusCode, 200; got != want {
			t.Errorf("status is %d, want %d", got, want)
		}

		if got, want := string(data), "pong"; got != want {
			t.Errorf("returned %q, want %q", got, want)
		}
	})

	t.Run("docker http server", func(t *testing.T) {
		defer startServerOnDocker(t, "docker")()

		time.Sleep(5 * time.Second)

		cli := ery.HTTPClient()

		resp, err := cli.Get(getURL("docker") + "/ping")
		checkErr(t, err)
		data, err := ioutil.ReadAll(resp.Body)
		checkErr(t, err)

		if got, want := resp.StatusCode, 200; got != want {
			t.Errorf("status is %d, want %d", got, want)
		}

		if got, want := string(data), "pong"; got != want {
			t.Errorf("returned %q, want %q", got, want)
		}
	})

	t.Run("docker to docker", func(t *testing.T) {
		defer startServerOnDocker(t, "docker1")()
		defer startServerOnDocker(t, "docker2")()

		time.Sleep(5 * time.Second)

		cli := ery.HTTPClient()

		v := url.Values{}
		v.Add("url", getURL("docker2")+"/ping")
		resp, err := cli.Get(getURL("docker1") + "/delegate?" + v.Encode())
		checkErr(t, err)
		data, err := ioutil.ReadAll(resp.Body)
		checkErr(t, err)

		if got, want := resp.StatusCode, 200; got != want {
			t.Errorf("status is %d, want %d", got, want)
		}

		if got, want := string(data), "pong"; got != want {
			t.Errorf("returned %q, want %q", got, want)
		}
	})

	t.Run("docker to local", func(t *testing.T) {
		defer startServerOnDocker(t, "docker")()
		defer startServerOnLocal(t, "local")()

		time.Sleep(5 * time.Second)

		cli := ery.HTTPClient()

		v := url.Values{}
		v.Add("url", getURL("local")+"/ping")
		resp, err := cli.Get(getURL("docker") + "/delegate?" + v.Encode())
		checkErr(t, err)
		data, err := ioutil.ReadAll(resp.Body)
		checkErr(t, err)

		if got, want := resp.StatusCode, 200; got != want {
			t.Errorf("status is %d, want %d", got, want)
		}

		if got, want := string(data), "pong"; got != want {
			t.Errorf("returned %q, want %q", got, want)
		}
	})

	time.Sleep(5 * time.Second)
}

func checkCmd(t *testing.T, cmd *exec.Cmd) {
	t.Helper()
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Log(string(out))
		t.Errorf("unexpected error: %v", err)
	}
}

func checkErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func getFreePort(t *testing.T) domain.Port {
	t.Helper()
	port, err := netutil.GetFreePort()
	checkErr(t, err)
	return port
}

type ery struct {
	bin                string
	dnsPort, proxyPort domain.Port
}

func newEry(t *testing.T, bin string) *ery {
	t.Helper()

	return &ery{
		bin:       bin,
		dnsPort:   getFreePort(t),
		proxyPort: getFreePort(t),
	}
}

func (e *ery) Command(ctx context.Context, args ...string) *exec.Cmd {
	args = append([]string{fmt.Sprintf("--dns-port=%d", e.dnsPort), fmt.Sprintf("--proxy-port=%d", e.proxyPort), "-v"}, args...)
	return exec.CommandContext(ctx, e.bin, args...)
}

func (e *ery) HTTPClient() *http.Client {
	dialerFunc := func(ctx context.Context, network, address string) (net.Conn, error) {
		d := net.Dialer{}
		return d.DialContext(ctx, "udp", fmt.Sprintf(":%d", e.dnsPort))
	}

	resolver := &net.Resolver{PreferGo: true, Dial: dialerFunc}
	dialer := net.Dialer{Resolver: resolver}

	transport := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		Dial:                  dialer.Dial,
		DialContext:           dialer.DialContext,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{Transport: transport}
}
