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
	// ExperimentalMonoArrays and ExperimentalMonoArraysSet are retained for
	// source compatibility only. Static Array lowering is permanently enabled;
	// these fields no longer permit callers to select runtime-store carriers.
	ExperimentalMonoArrays    bool
	ExperimentalMonoArraysSet bool
	// ExperimentalExecutionContext force-enables the generated-call context ABI
	// for diagnostic comparison. Programs that contain an Able await expression
	// select the scheduler-required context ABI automatically.
	ExperimentalExecutionContext bool
	// EmitDynamicBoundaryTelemetry adds debug-only counters to generated code
	// for classified dynamic/host/runtime-service crossings. It must remain
	// opt-in so normal compiled binaries have no telemetry branch or allocation.
	EmitDynamicBoundaryTelemetry bool
	// EmitCallPathTelemetry adds debug-only counters for residual generated
	// call helpers. It is selection instrumentation only: normal generated
	// binaries must not contain its counters, branches, or atomic operations.
	EmitCallPathTelemetry bool
	// EmitTypedBoundaryTelemetry adds main-only debug counters for conversions
	// between native generated values and the shared runtime representation.
	// It is selection instrumentation and must remain absent from normal builds.
	EmitTypedBoundaryTelemetry bool
	// CollectNominalEffects computes conservative typed callable effects for
	// diagnostics. The result does not select carriers or alter generated Go.
	CollectNominalEffects bool
	// CollectNominalOwnership computes fail-closed interprocedural ownership
	// transfer proofs for diagnostics. Report collection remains independent
	// from the default generated ownership execution path.
	CollectNominalOwnership bool
	// ExperimentalNominalOwnership is retained for source compatibility.
	// Proven caller-owned nominal-result lowering is enabled by default.
	ExperimentalNominalOwnership bool
	// DisableNominalOwnership disables caller-owned nominal-result lowering for
	// diagnostic baselines. Ordinary compilation must leave this false.
	DisableNominalOwnership bool
}

type Result struct {
	Files            map[string][]byte
	Warnings         []string
	Fallbacks        []FallbackInfo
	NominalEffects   *NominalEffectReport
	NominalOwnership *NominalOwnershipReport
}

type Compiler struct {
	opts Options
}

func New(opts Options) *Compiler {
	if opts.PackageName == "" {
		opts.PackageName = "ablecompiled"
	}
	if !opts.ExperimentalMonoArraysSet {
		opts.ExperimentalMonoArrays = true
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
		switch diag.Diagnostic.Code {
		case typechecker.DiagnosticCodeStaticOnlyInterfaceMethod:
			return nil, fmt.Errorf("compiler: static-only interface method call rejected: %s", message)
		case typechecker.DiagnosticCodeInvariantTypeArgument:
			return nil, fmt.Errorf("compiler: invariant type argument mismatch rejected: %s", message)
		case typechecker.DiagnosticCodeCallableSignatureMismatch:
			return nil, fmt.Errorf("compiler: callable signature mismatch rejected: %s", message)
		}
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
	gen.resolveCallerOwnedResults()
	var nominalOwnership *NominalOwnershipReport
	if !c.opts.DisableNominalOwnership {
		executionReport := gen.prepareNominalOwnershipExecution()
		if c.opts.CollectNominalOwnership {
			nominalOwnership = executionReport
		}
	}
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
	var nominalEffects *NominalEffectReport
	if c.opts.CollectNominalEffects {
		nominalEffects = gen.resolveNominalParameterEffects()
	}
	if c.opts.CollectNominalOwnership && nominalOwnership == nil {
		nominalOwnership = gen.resolveNominalOwnership()
	}
	gen.warnings = append(warnings, gen.warnings...)
	return &Result{
		Files:            files,
		Warnings:         gen.warnings,
		Fallbacks:        fallbacks,
		NominalEffects:   nominalEffects,
		NominalOwnership: nominalOwnership,
	}, nil
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
