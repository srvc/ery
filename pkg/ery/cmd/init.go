package cmd

import (
	"io/ioutil"
	"path/filepath"
	"strings"

	toml "github.com/pelletier/go-toml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/srvc/ery/pkg/app/command"
	"github.com/srvc/ery/pkg/ery"
	"go.uber.org/zap"
)

func newCmdInit(cfg *ery.Config) *cobra.Command {
	wd := cfg.WorkingDir

	var (
		proj = filepath.Base(wd)
		org  = filepath.Base(filepath.Dir(wd))
		tld  = cfg.TLD
	)

	hostname := strings.Join([]string{proj, org, tld}, ".")

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Genereate an ery's configuration file",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := command.Config{Hostname: hostname}
			data, err := toml.Marshal(cfg)
			if err != nil {
				return errors.WithStack(err)
			}

			err = ioutil.WriteFile(filepath.Join(wd, ".ery.toml"), data, 0644)
			if err != nil {
				return errors.WithStack(err)
			}

			zap.L().Info(".ery.toml is geneerated", zap.Any("config", cfg))

			return nil
		},
	}

	cmd.PersistentFlags().StringVar(&hostname, "host", hostname, "Hostname")

	return cmd
}
