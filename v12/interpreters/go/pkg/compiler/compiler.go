package compiler

import (
	"fmt"

	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/typechecker"
)

type Options struct {
	PackageName              string
	EmitMain                 bool
	EntryPath                string
	RequireNoFallbacks       bool
	RequireStaticNoFallbacks bool
	ExperimentalMonoArrays   bool
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
	gen.setTypecheckInference(check.Inferred)
	if err := gen.collect(program); err != nil {
		return nil, err
	}
	dynamicReport, err := DetectDynamicFeatures(program)
	if err != nil {
		return nil, err
	}
	gen.setDynamicFeatureReport(dynamicReport)
	// collect() resolves compileability before dynamic usage is known; rerun so
	// dynamic modules are allowed to keep explicit boundary call sites compiled.
	gen.resolveCompileabilityFixedPoint()
	gen.discardRedundantImplFallbackSpecializations()
	appendDynamicFeatureWarnings(gen, dynamicReport)
	fallbacks := gen.collectFallbacks()
	if err := c.validateFallbackPolicy(fallbacks, dynamicReport); err != nil {
		return nil, err
	}
	files, err := gen.render()
	if err != nil {
		return nil, err
	}
	gen.discardRedundantImplFallbackSpecializations()
	fallbacks = gen.collectFallbacks()
	if err := c.validateFallbackPolicy(fallbacks, dynamicReport); err != nil {
		return nil, err
	}
	gen.warnings = append(warnings, gen.warnings...)
	return &Result{Files: files, Warnings: gen.warnings, Fallbacks: fallbacks}, nil
}

func (c *Compiler) validateFallbackPolicy(fallbacks []FallbackInfo, dynamicReport *DynamicFeatureReport) error {
	if len(fallbacks) == 0 {
		return nil
	}
	first := fallbacks[0]
	name := first.Name
	if name == "" {
		name = "<unknown>"
	}
	reason := first.Reason
	if reason == "" {
		reason = "unspecified fallback reason"
	}
	if c.opts.RequireNoFallbacks {
		return fmt.Errorf("compiler: fallback not allowed (count=%d, first=%s: %s)", len(fallbacks), name, reason)
	}
	if c.opts.RequireStaticNoFallbacks && (dynamicReport == nil || !dynamicReport.UsesDynamic()) {
		return fmt.Errorf("compiler: static fallback not allowed (count=%d, first=%s: %s)", len(fallbacks), name, reason)
	}
	return nil
}

func (r *Result) Write(dir string) error {
	if r == nil {
		return fmt.Errorf("compiler: nil result")
	}
	return writeFiles(dir, r.Files)
}
