package main

import (
	"os"

	"github.com/b-jonathan/taco/internal/cli"
	"github.com/b-jonathan/taco/internal/logx"
)

func main() {
	if err := logx.Init(); err != nil {
		logx.Errorf("%v", err)
		os.Exit(1)
	}
	if err := cli.Execute(); err != nil {
		logx.Errorf("%v", err)
		os.Exit(1)
	}
}
