package ery_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/spf13/afero"
	"github.com/srvc/ery"
)

func TestNewConfig(t *testing.T) {
	yaml1 := `
<: &accounts-k8s
  name: accounts
  hostname: accounts.ery.local
  kubernetes:
    context: k8s.srvc.tools
    namespace: accounts
    labels:
      role: api
    ports:
      80: 8080

<: &admin-k8s
  name: admin
  hostname: admin.ery.local
  kubernetes:
    context: k8s.srvc.tools
    namespace: admin
    labels:
      role: web
    ports:
      80: 3000
`

	yaml2 := `
<: &blog-local
  name: blog
  hostname: blog.ery.local
  local:
    port_env:
      PORT: 80
    path: example.com/blog
    cmd: ["go", "run", "./cmd/server"]
`

	yaml3 := `
<: &admin-docker
  name: admin
  hostname: admin.ery.local
  docker:
    path: example.com/admin
    build:
      dockerfile: ./dev.dockerfile
    run:
      cmd: ["bin/rails", "s"]
      volumes:
      - .:/app
      - bundle:/usr/local/bundle
      - tmp:/app/tmp
      - log:/app/log
      port_envs:
      - PORT: 80
`

	yaml4 := `
root: /Users/ery/src

projects:
- name: blog
  apps:
  - *blog-local
  - *admin-k8s
  - *accounts-k8s

- name: admin
  apps:
  - *blog-local
  - *admin-docker
  - *accounts-k8s
`

	wantCfg := &ery.Config{
		Root: "/Users/ery/src",
		Projects: []*ery.Project{
			{
				Name: "blog",
				Apps: []*ery.App{
					{
						Name:     "blog",
						Hostname: "blog.ery.local",
						Local: &ery.LocalApp{
							PortEnv: map[string]ery.Port{"PORT": ery.Port(80)},
							Path:    "example.com/blog",
							Cmd:     []string{"go", "run", "./cmd/server"},
						},
					},
					{
						Name:     "admin",
						Hostname: "admin.ery.local",
						Kubernetes: &ery.KubernetesApp{
							Context:   "k8s.srvc.tools",
							Namespace: "admin",
							Labels:    map[string]string{"role": "web"},
							Ports:     map[ery.Port]ery.Port{80: 3000},
						},
					},
					{
						Name:     "accounts",
						Hostname: "accounts.ery.local",
						Kubernetes: &ery.KubernetesApp{
							Context:   "k8s.srvc.tools",
							Namespace: "accounts",
							Labels:    map[string]string{"role": "api"},
							Ports:     map[ery.Port]ery.Port{80: 8080},
						},
					},
				},
			},
			{
				Name: "admin",
				Apps: []*ery.App{
					{
						Name:     "blog",
						Hostname: "blog.ery.local",
						Local: &ery.LocalApp{
							PortEnv: map[string]ery.Port{"PORT": ery.Port(80)},
							Path:    "example.com/blog",
							Cmd:     []string{"go", "run", "./cmd/server"},
						},
					},
					{
						Name:     "admin",
						Hostname: "admin.ery.local",
						Docker: &ery.DockerApp{
							Ports: []ery.Port{80},
							Path:  "example.com/admin",
							Cmd:   []string{"bin/rails", "s"},
						},
					},
					{
						Name:     "accounts",
						Hostname: "accounts.ery.local",
						Kubernetes: &ery.KubernetesApp{
							Context:   "k8s.srvc.tools",
							Namespace: "accounts",
							Labels:    map[string]string{"role": "api"},
							Ports:     map[ery.Port]ery.Port{80: 8080},
						},
					},
				},
			},
		},
	}

	configDir := filepath.Join(os.Getenv("HOME"), ".config", "ery")

	cases := []struct {
		test  string
		setup func(t *testing.T, fs afero.Fs)
	}{
		{
			test: "with ery.yaml",
			setup: func(t *testing.T, fs afero.Fs) {
				afero.WriteFile(fs, filepath.Join(configDir, "ery.k8s.yaml"), []byte(yaml1), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.local.yaml"), []byte(yaml2), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.docker.yaml"), []byte(yaml3), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.yaml"), []byte(yaml4), 0644)
			},
		},
		{
			test: "without ery.yaml",
			setup: func(t *testing.T, fs afero.Fs) {
				afero.WriteFile(fs, filepath.Join(configDir, "ery.k8s.yaml"), []byte(yaml1), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.local.yaml"), []byte(yaml2), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.docker.yaml"), []byte(yaml3), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.projects.yaml"), []byte(yaml4), 0644)
			},
		},
		{
			test: "with ery.yml",
			setup: func(t *testing.T, fs afero.Fs) {
				afero.WriteFile(fs, filepath.Join(configDir, "ery.k8s.yml"), []byte(yaml1), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.local.yml"), []byte(yaml2), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.docker.yml"), []byte(yaml3), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.yml"), []byte(yaml4), 0644)
			},
		},
		{
			test: "without ery.yml",
			setup: func(t *testing.T, fs afero.Fs) {
				afero.WriteFile(fs, filepath.Join(configDir, "ery.k8s.yml"), []byte(yaml1), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.local.yml"), []byte(yaml2), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.docker.yml"), []byte(yaml3), 0644)
				afero.WriteFile(fs, filepath.Join(configDir, "ery.projects.yml"), []byte(yaml4), 0644)
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.test, func(t *testing.T) {
			baseFs := afero.NewMemMapFs()
			tc.setup(t, baseFs)

			fs, err := ery.NewUnionFs(baseFs)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			viper := ery.NewViper(fs)
			cfg, err := ery.NewConfig(viper)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if diff := cmp.Diff(wantCfg, cfg); diff != "" {
				t.Errorf("loaded config mismatch (-want +got):\n%s", diff)
			}
		})
	}
}
