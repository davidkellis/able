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

	execMode, remaining, err := parseExecMode(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	if len(remaining) == 0 {
		printUsage()
		return 1
	}

	switch remaining[0] {
	case "--help", "-h":
		printUsage()
		return 0
	case "--version", "-V", "version":
		fmt.Fprintln(os.Stdout, cliToolVersion)
		return 0
	case "run":
		return runEntry(remaining[1:], execMode)
	case "repl":
		return runRepl(remaining[1:], execMode)
	case "check":
		return runCheck(remaining[1:], execMode)
	case "test":
		return runTest(remaining[1:], execMode)
	case "deps":
		return runDeps(remaining[1:])
	default:
		return runEntry(remaining, execMode)
	}
}
