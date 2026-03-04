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
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] run [--with-tests] [target]")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] run [--with-tests] <file.able>")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] <file.able>")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] check [--with-tests] [target]")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] check [--with-tests] <file.able>")
	fmt.Fprintln(os.Stderr, "  able build [--with-tests] [target]")
	fmt.Fprintln(os.Stderr, "  able build [--with-tests] <file.able>")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] test [--compiled] [paths]")
	fmt.Fprintln(os.Stderr, "  able [--exec-mode=treewalker|bytecode] repl")
	fmt.Fprintln(os.Stderr, "  able deps install")
	fmt.Fprintln(os.Stderr, "  able deps update [dependency ...]")
	fmt.Fprintln(os.Stderr, "  able override add <git-url> <local-path>")
	fmt.Fprintln(os.Stderr, "  able override remove <git-url>")
	fmt.Fprintln(os.Stderr, "  able override list")
	fmt.Fprintln(os.Stderr, "  able setup")
}
