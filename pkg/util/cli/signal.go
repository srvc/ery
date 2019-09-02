package cliutil

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"
)

func CobraRunE(f func(ctx context.Context, cmd *cobra.Command, args []string) error, opts ...Option) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
		opts = append([]Option{WithStdout(cmd.OutOrStdout())}, opts...)
		return Run(context.Background(), func(ctx context.Context) error { return f(ctx, cmd, args) }, opts...)
	}
}

func Run(ctx context.Context, f func(context.Context) error, opts ...Option) error {
	cfg := DefaultConfig.clone()
	cfg.apply(opts)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	sdCh := make(chan error)
	defer close(sdCh)

	go func() { sdCh <- f(ctx) }()

	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, cfg.Signals...)
	defer close(sigCh)
	defer signal.Stop(sigCh)

	var sigCnt int

	for {
		select {
		case <-sigCh:
			sigCnt++
			switch sigCnt {
			case 1:
				fmt.Fprintln(cfg.Stdout, "Received signal, gracefully stopping...")
				cancel()
			case 3:
				return errors.New("Aborting")
			}
		case err := <-sdCh:
			fmt.Fprintln(cfg.Stdout, "Shutdown successfully")
			return err
		}
	}
}

var DefaultConfig = &Config{
	Signals: []os.Signal{os.Interrupt, syscall.SIGINT, syscall.SIGTERM},
	Stdout:  ioutil.Discard,
}

type Config struct {
	Signals []os.Signal
	Stdout  io.Writer
}

func (c *Config) clone() *Config { cc := *c; return &cc }

func (c *Config) apply(opts []Option) {
	for _, f := range opts {
		f(c)
	}
}

type Option func(*Config)

func WithSignals(sigs ...os.Signal) Option {
	return func(c *Config) { c.Signals = append(c.Signals, sigs...) }
}

func WithStdout(w io.Writer) Option {
	return func(c *Config) { c.Stdout = w }
}
