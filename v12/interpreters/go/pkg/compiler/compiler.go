package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/driver"
)

type Options struct {
	PackageName string
	EmitMain    bool
	EntryPath   string
}

type Result struct {
	Files    map[string][]byte
	Warnings []string
}

type Compiler struct {
	opts Options
}

func New(opts Options) *Compiler {
	if opts.PackageName == "" {
		opts.PackageName = "ablecompiled"
	}
	return &Compiler{opts: opts}
}

func (c *Compiler) Compile(program *driver.Program) (*Result, error) {
	if program == nil || program.Entry == nil || program.Entry.AST == nil {
		return nil, fmt.Errorf("compiler: missing entry program")
	}
	gen := newGenerator(c.opts)
	if err := gen.collect(program); err != nil {
		return nil, err
	}
	files, err := gen.render()
	if err != nil {
		return nil, err
	}
	return &Result{Files: files, Warnings: gen.warnings}, nil
}

func (r *Result) Write(dir string) error {
	if r == nil {
		return fmt.Errorf("compiler: nil result")
	}
	return writeFiles(dir, r.Files)
}
