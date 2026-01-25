package main

import (
	"errors"
	"fmt"
	"os"
)

const cliToolVersion = "able-cli 0.0.0-dev"

var errManifestNotFound = errors.New("package.yml not found")

type executionMode int

const (
	modeRun executionMode = iota
	modeCheck
)

func main() {
	os.Exit(run(os.Args[1:]))
}

func run(args []string) int {
	if len(args) == 0 {
		printUsage()
		return 1
	}

	switch args[0] {
	case "--help", "-h":
		printUsage()
		return 0
	case "--version", "-V", "version":
		fmt.Fprintln(os.Stdout, cliToolVersion)
		return 0
	case "run":
		return runEntry(args[1:])
	case "repl":
		return runRepl(args[1:])
	case "check":
		return runCheck(args[1:])
	case "test":
		return runTest(args[1:])
	case "deps":
		return runDeps(args[1:])
	default:
		return runEntry(args)
	}
}
