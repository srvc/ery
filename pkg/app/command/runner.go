package command

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/pkg/errors"
	"github.com/spf13/afero"
	"github.com/srvc/ery/pkg/domain"
	"go.uber.org/zap"
)

type Runner interface {
	Run(ctx context.Context, name string, args []string) error
}

func NewRunner(
	fs afero.Fs,
	mappingRepo domain.MappingRepository,
	defaultPort domain.Port,
	workingDir string,
	outW, errW io.Writer,
	inR io.Reader,
) Runner {
	return &runnerImpl{
		fs:          fs,
		mappingRepo: mappingRepo,
		workingDir:  workingDir,
		defaultPort: defaultPort,
		outW:        outW,
		errW:        errW,
		inR:         inR,
		log:         zap.L().Named("command"),
	}
}

type runnerImpl struct {
	fs          afero.Fs
	mappingRepo domain.MappingRepository
	defaultPort domain.Port
	workingDir  string
	outW, errW  io.Writer
	inR         io.Reader

	log *zap.Logger

	cfg  *Config
	port domain.Port
}

func (r *runnerImpl) Run(ctx context.Context, name string, args []string) error {
	err := r.setup(ctx)
	if err != nil {
		return errors.WithStack(err)
	}

	defer r.cleanup(context.TODO())

	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdin = r.inR
	cmd.Stdout = r.outW
	cmd.Stderr = r.errW
	cmd.Env = append(os.Environ(), fmt.Sprintf("PORT=%d", r.port))
	r.log.Debug("execute command",
		zap.String("name", name),
		zap.Strings("args", args),
		zap.String("host", r.cfg.Hostname),
		zap.Uint16("port", uint16(r.port)),
	)

	return errors.WithStack(cmd.Run())
}

func (r *runnerImpl) setup(ctx context.Context) error {
	var err error

	r.cfg, err = loadConfig(r.fs, r.workingDir, ".ery")
	if err != nil {
		return errors.WithStack(err)
	}

	addr := domain.Addr{
		Host: r.cfg.Hostname,
		Port: r.defaultPort, // TODO: should be configurable
	}
	rAddr, err := r.mappingRepo.Create(ctx, addr, 0)
	if err != nil {
		return errors.WithStack(err)
	}
	r.port = rAddr.Port

	return nil
}

func (r *runnerImpl) cleanup(ctx context.Context) (err error) {
	err = errors.WithStack(r.mappingRepo.DeleteByHost(ctx, r.cfg.Hostname))
	if err != nil {
		r.log.Warn(
			"deleting mappings returned error",
			zap.String("host", r.cfg.Hostname),
			zap.Uint16("port", uint16(r.port)),
			zap.Error(err),
		)
	}
	return
}
