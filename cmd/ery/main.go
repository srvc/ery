package main

import (
	"fmt"
	"os"

	"github.com/izumin5210/clig/pkg/clib"
	"github.com/srvc/ery/cmd/ery/cmd"
)

func main() {
	defer clib.Close()

	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run() error {
	return cmd.New(clib.Stdio()).Execute()
}
