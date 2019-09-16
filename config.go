package ery

import (
	"bytes"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/afero"
	"github.com/spf13/viper"

	api_pb "github.com/srvc/ery/api"
)

type Config struct {
	Root     string
	Projects []*Project
}

type Project struct {
	Name string
	Apps []*App
}

type App struct {
	Name     string
	Hostname string

	Local      *LocalApp
	Docker     *DockerApp
	Kubernetes *KubernetesApp
}

func (a *App) Pb() *api_pb.App {
	pb := &api_pb.App{
		Name:     a.Name,
		Hostname: a.Hostname,
	}

	switch {
	case a.Local != nil:
		pb.Type = api_pb.App_TYPE_LOCAL
		for name, port := range a.Local.PortEnv {
			pb.Ports = append(pb.Ports, &api_pb.App_Port{
				Network:     api_pb.App_Port_TCP, // TODO
				ExposedPort: uint32(port),
				Env:         name,
			})
		}
	case a.Docker != nil:
		pb.Type = api_pb.App_TYPE_DOCKER
		for name, port := range a.Docker.Run.PortEnv {
			pb.Ports = append(pb.Ports, &api_pb.App_Port{
				Network:     api_pb.App_Port_TCP, // TODO
				ExposedPort: uint32(port),
				Env:         name,
			})
		}
	case a.Kubernetes != nil:
		pb.Type = api_pb.App_TYPE_KUBERNETES
		// TODO: not yet supported
	}

	return pb
}

type LocalApp struct {
	PortEnv map[string]Port `mapstructure:"port_env"`
	Cmd     []string
	Path    string
}

type DockerApp struct {
	Path  string
	Build struct {
		Dockerfile string
		Target     string
	}
	Run struct {
		Cmd     [][]string
		Volumes []string
		PortEnv map[string]Port `mapstructure:"port_env"`
	}
}

type KubernetesApp struct {
	Context   string
	Namespace string
	Labels    map[string]string
	Ports     map[Port]Port
}

func (c *Config) FindProject(name string) *Project {
	for _, p := range c.Projects {
		if p.Name == name {
			return p
		}
	}
	return nil
}

var configDir = filepath.Join(os.Getenv("HOME"), ".config", "ery")

func NewViper(fs afero.Fs) *viper.Viper {
	v := viper.New()
	v.SetFs(fs)
	v.AddConfigPath(configDir)
	v.SetConfigName("ery")
	return v
}

func NewFs() afero.Fs {
	return afero.NewOsFs()
}

type UnionFS struct {
	afero.Fs
}

func (fs *UnionFS) MergeConfigFiles() error {
	var matches []string
	for _, ext := range []string{"yaml", "yml"} {
		resp, err := afero.Glob(fs, filepath.Join(configDir, "ery.*."+ext))
		if err != nil {
			return err
		}
		matches = append(matches, resp...)
	}

	if len(matches) == 0 {
		return nil
	}

	path := filepath.Join(configDir, "ery.yml")
	if ok, _ := afero.Exists(fs, path); !ok {
		path = filepath.Join(configDir, "ery.yaml")
	}
	out, err := fs.OpenFile(path, os.O_RDWR, 0644)
	buf := new(bytes.Buffer)
	switch {
	case err == nil:
		_, err = io.Copy(buf, out)
		if err != nil {
			return err
		}
		err = out.Truncate(0)
		if err != nil {
			return err
		}
		_, err = out.Seek(0, 0)
		if err != nil {
			return err
		}
	case os.IsNotExist(err):
		out, err = fs.Create(path)
		if err != nil {
			return err
		}
	default:
		return err
	}
	defer out.Close()

	for _, match := range matches {
		in, err := fs.Open(match)
		if err != nil {
			return err
		}
		defer in.Close()
		_, err = io.Copy(out, in)
		if err != nil {
			return err
		}
	}

	_, err = io.Copy(out, buf)
	if err != nil {
		return err
	}

	return nil
}

func NewUnionFs(baseFs afero.Fs) (*UnionFS, error) {
	fs := afero.NewCopyOnWriteFs(
		afero.NewReadOnlyFs(baseFs),
		afero.NewMemMapFs(),
	)

	ufs := &UnionFS{Fs: fs}
	err := ufs.MergeConfigFiles()
	if err != nil {
		return nil, err
	}

	return ufs, nil
}

func NewConfig(v *viper.Viper) (*Config, error) {
	err := v.ReadInConfig()
	if err != nil {
		return nil, err
	}
	var cfg Config
	err = v.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}
