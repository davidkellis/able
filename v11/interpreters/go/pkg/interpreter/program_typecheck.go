package interpreter

import (
	"able/interpreter10-go/pkg/driver"
	"able/interpreter10-go/pkg/typechecker"
)

// ModuleDiagnostic is re-exported for backwards compatibility.
type ModuleDiagnostic = typechecker.ModuleDiagnostic

// PackageSummary exposes the public API metadata collected during program checks.
type PackageSummary = typechecker.PackageSummary

// ProgramCheckResult aggregates diagnostics and package summaries for a program check.
type ProgramCheckResult = typechecker.CheckResult

// DescribeModuleDiagnostic proxies to the typechecker helper for formatting.
func DescribeModuleDiagnostic(diag ModuleDiagnostic) string {
	return typechecker.DescribeModuleDiagnostic(diag)
}

// TypecheckProgram applies the Able typechecker across all modules in the provided program
// and returns diagnostics plus per-package export summaries.
func TypecheckProgram(program *driver.Program) (ProgramCheckResult, error) {
	pc := typechecker.NewProgramChecker()
	return pc.Check(program)
}
