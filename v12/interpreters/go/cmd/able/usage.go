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
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] run [target]")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] run <file.able>")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] <file.able>")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] check [target]")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] check <file.able>")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] test [paths]")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] repl")
	fmt.Fprintln(os.Stderr, "  able deps install")
	fmt.Fprintln(os.Stderr, "  able deps update [dependency ...]")
}
