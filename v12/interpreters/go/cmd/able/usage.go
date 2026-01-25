package main

import (
	"fmt"
	"os"
)

func modeCommandLabel(mode executionMode) string {
	switch mode {
	case modeCheck:
		return "able check"
	default:
		return "able run"
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "Usage:")
	fmt.Fprintln(os.Stderr, "  able run [target]")
	fmt.Fprintln(os.Stderr, "  able run <file.able>")
	fmt.Fprintln(os.Stderr, "  able <file.able>")
	fmt.Fprintln(os.Stderr, "  able check [target]")
	fmt.Fprintln(os.Stderr, "  able check <file.able>")
	fmt.Fprintln(os.Stderr, "  able test [paths]")
	fmt.Fprintln(os.Stderr, "  able repl")
	fmt.Fprintln(os.Stderr, "  able deps install")
	fmt.Fprintln(os.Stderr, "  able deps update [dependency ...]")
}
