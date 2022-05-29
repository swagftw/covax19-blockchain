package main

import (
	"github.com/swagftw/covax19-blockchain/pkg/cli"
	"os"
)

func main() {
	defer os.Exit(0)

	cmd := cli.CommandLine{}
	cmd.Run()
}
