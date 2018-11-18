package command

import (
	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/spf13/viper"
)

type Config struct {
	Hostname string `toml:"hostname"`
}

func loadConfig(fs afero.Fs, wd string, filename string) (cfg *Config, err error) {
	v := viper.New()
	v.SetFs(fs)
	v.AddConfigPath(wd)
	v.SetConfigName(filename)

	err = errors.WithStack(v.ReadInConfig())
	if err != nil {
		return
	}

	err = errors.WithStack(v.Unmarshal(&cfg))

	return
}
