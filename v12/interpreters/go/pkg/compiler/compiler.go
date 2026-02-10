package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/typechecker"
)

type Options struct {
	PackageName string
	EmitMain    bool
	EntryPath   string
}

type Result struct {
	Files     map[string][]byte
	Warnings  []string
	Fallbacks []FallbackInfo
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
	checker := typechecker.NewProgramChecker()
	check, err := checker.Check(program)
	if err != nil {
		return nil, err
	}
	var warnings []string
	for _, diag := range check.Diagnostics {
		message := typechecker.DescribeModuleDiagnostic(diag)
		warnings = append(warnings, message)
	}
	gen := newGenerator(c.opts)
	if err := gen.collect(program); err != nil {
		return nil, err
	}
	if report, err := DetectDynamicFeatures(program); err != nil {
		return nil, err
	} else {
		appendDynamicFeatureWarnings(gen, report)
	}
	files, err := gen.render()
	if err != nil {
		return nil, err
	}
	gen.warnings = append(warnings, gen.warnings...)
	return &Result{Files: files, Warnings: gen.warnings, Fallbacks: gen.collectFallbacks()}, nil
}

func (r *Result) Write(dir string) error {
	if r == nil {
		return fmt.Errorf("compiler: nil result")
	}
	return writeFiles(dir, r.Files)
}
