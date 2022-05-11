package main

import (
	"github.com/swagftw/covax19-blockchain/cli"
	"os"
)

func main() {
	defer os.Exit(0)
	commandLine := cli.CommandLine{}
	commandLine.Run()
}
