package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"able/interpreter-go/pkg/parser"
)

func main() {
	source, err := io.ReadAll(os.Stdin)
	if err != nil {
		exitErr("read source: %v", err)
	}

	p, err := parser.NewModuleParser()
	if err != nil {
		exitErr("init parser: %v", err)
	}
	defer p.Close()

	mod, err := p.ParseModule(source)
	if err != nil {
		exitErr("parse module: %v", err)
	}

	out, err := json.Marshal(mod)
	if err != nil {
		exitErr("encode module: %v", err)
	}

	if _, err := os.Stdout.Write(out); err != nil {
		exitErr("write output: %v", err)
	}
}

func exitErr(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
