package compiler

import (
	"fmt"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/typechecker"
)

func shadowedImportedNominalRescuePackageFiles() map[string]string {
	return map[string]string{
		"main.able": strings.Join([]string{
			"package demo",
			"",
			"import demo.remote.{Thing::RemoteThing}",
			"",
			"struct Thing { local: i32 }",
			"",
			"fn fail() -> RemoteThing {",
			"  raise(RemoteThing { remote: 1 })",
			"  RemoteThing { remote: 0 }",
			"}",
			"",
			"fn main() -> i32 {",
			"  do {",
			"    fail()",
			"    0",
			"  } rescue {",
			"    case thing => {",
			"      recovered := if true { thing } else { RemoteThing { remote: 2 } }",
			"      recovered.remote",
			"    }",
			"  }",
			"}",
			"",
		}, "\n"),
		"remote/module.able": strings.Join([]string{
			"struct Thing { remote: i32 }",
			"",
		}, "\n"),
	}
}

func shadowedImportedNominalFailureInferenceParts(t *testing.T) (*generator, *compileContext, *functionInfo, *ast.RaiseStatement) {
	t.Helper()
	program := loadShadowedInterfaceProgramFromFiles(t, shadowedImportedNominalRescuePackageFiles())
	checker := typechecker.NewProgramChecker()
	check, err := checker.Check(program)
	if err != nil {
		t.Fatalf("typecheck: %v", err)
	}
	gen := newGenerator(Options{PackageName: "main", EmitMain: true})
	gen.setTypecheckInference(check.Inferred)
	if err := gen.collect(program); err != nil {
		t.Fatalf("collect: %v", err)
	}
	report, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("dynamic features: %v", err)
	}
	gen.setDynamicFeatureReport(report)
	gen.resolveCompileabilityFixedPoint()

	var info *functionInfo
	for _, candidate := range gen.allFunctionInfos() {
		if candidate == nil || candidate.Definition == nil || candidate.Definition.ID == nil {
			continue
		}
		if candidate.Definition.ID.Name == "fail" {
			info = candidate
			break
		}
	}
	if info == nil || info.Definition == nil || info.Definition.Body == nil {
		names := make([]string, 0, len(gen.allFunctionInfos()))
		for _, candidate := range gen.allFunctionInfos() {
			if candidate == nil {
				continue
			}
			name := candidate.Name
			if candidate.Package != "" {
				name = fmt.Sprintf("%s::%s", candidate.Package, name)
			}
			if candidate.QualifiedName != "" {
				name += " [" + candidate.QualifiedName + "]"
			}
			names = append(names, name)
		}
		t.Fatalf("missing fail function info: packages=%v", names)
	}

	var raiseStmt *ast.RaiseStatement
	for _, stmt := range info.Definition.Body.Body {
		if raise, ok := stmt.(*ast.RaiseStatement); ok {
			raiseStmt = raise
			break
		}
	}
	if raiseStmt == nil || raiseStmt.Expression == nil {
		t.Fatalf("missing fail raise statement")
	}

	ctx := newCompileContext(gen, info, gen.functions[info.Package], gen.overloads[info.Package], info.Package, genericNameSet(info.Definition.GenericParams))
	return gen, ctx, info, raiseStmt
}

func shadowedImportedNominalRescueMainParts(t *testing.T) (*generator, *compileContext, *functionInfo, *ast.RescueExpression) {
	t.Helper()
	program := loadShadowedInterfaceProgramFromFiles(t, shadowedImportedNominalRescuePackageFiles())
	checker := typechecker.NewProgramChecker()
	check, err := checker.Check(program)
	if err != nil {
		t.Fatalf("typecheck: %v", err)
	}
	gen := newGenerator(Options{PackageName: "main", EmitMain: true})
	gen.setTypecheckInference(check.Inferred)
	if err := gen.collect(program); err != nil {
		t.Fatalf("collect: %v", err)
	}
	report, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("dynamic features: %v", err)
	}
	gen.setDynamicFeatureReport(report)
	gen.resolveCompileabilityFixedPoint()

	var info *functionInfo
	for _, candidate := range gen.allFunctionInfos() {
		if candidate == nil || candidate.Definition == nil || candidate.Definition.ID == nil {
			continue
		}
		if candidate.Definition.ID.Name == "main" && candidate.Package == "demo.demo" {
			info = candidate
			break
		}
	}
	if info == nil || info.Definition == nil || info.Definition.Body == nil {
		t.Fatalf("missing main function info")
	}
	var rescue *ast.RescueExpression
	for _, stmt := range info.Definition.Body.Body {
		if expr, ok := stmt.(*ast.RescueExpression); ok {
			rescue = expr
			break
		}
		if assign, ok := stmt.(*ast.AssignmentExpression); ok {
			if expr, ok := assign.Right.(*ast.RescueExpression); ok {
				rescue = expr
				break
			}
		}
	}
	if rescue == nil || rescue.MonitoredExpression == nil || len(rescue.Clauses) == 0 || rescue.Clauses[0] == nil {
		t.Fatalf("missing main rescue expression")
	}
	ctx := newCompileContext(gen, info, gen.functions[info.Package], gen.overloads[info.Package], info.Package, genericNameSet(info.Definition.GenericParams))
	return gen, ctx, info, rescue
}

func TestCompilerRaisedImportedShadowedNominalFailureTypePreservesRemotePackage(t *testing.T) {
	gen, ctx, _, raiseStmt := shadowedImportedNominalFailureInferenceParts(t)

	inferred := gen.failureTypeExprFromRaisedExpr(ctx, raiseStmt.Expression)
	if inferred == nil {
		t.Fatalf("expected raised imported nominal inference to succeed")
	}
	if gotPkg := gen.resolvedTypeExprPackage(ctx.packageName, inferred); gotPkg != "demo.remote" {
		t.Fatalf("expected raised imported nominal failure type to stay remote, got pkg=%q type=%q", gotPkg, normalizeTypeExprString(gen, ctx.packageName, inferred))
	}
}

func TestCompilerPropagatedImportedShadowedNominalFailureTypePreservesRemotePackage(t *testing.T) {
	gen, ctx, info, _ := shadowedImportedNominalFailureInferenceParts(t)

	inferred := gen.failureTypeExprFromFunctionInfoSeen(ctx, info, make(map[string]struct{}))
	if inferred == nil {
		t.Fatalf("expected propagated imported nominal inference to succeed")
	}
	if gotPkg := gen.resolvedTypeExprPackage(ctx.packageName, inferred); gotPkg != "demo.remote" {
		t.Fatalf("expected propagated imported nominal failure type to stay remote, got pkg=%q type=%q", gotPkg, normalizeTypeExprString(gen, ctx.packageName, inferred))
	}
}

func TestCompilerRescueBindingForImportedShadowedNominalFailureStaysNative(t *testing.T) {
	gen, ctx, _, rescue := shadowedImportedNominalRescueMainParts(t)

	rescueSubjectTypeExpr := gen.inferRescueSubjectTypeExpr(ctx, rescue.MonitoredExpression)
	if rescueSubjectTypeExpr == nil {
		t.Fatalf("expected rescue subject inference to succeed")
	}
	if gotPkg := gen.resolvedTypeExprPackage(ctx.packageName, rescueSubjectTypeExpr); gotPkg != "demo.remote" {
		t.Fatalf("expected rescue subject type to stay remote, got pkg=%q type=%q", gotPkg, normalizeTypeExprString(gen, ctx.packageName, rescueSubjectTypeExpr))
	}

	clauseCtx := ctx.child()
	clauseCtx.expectedTypeExpr = rescueSubjectTypeExpr
	_, _, bindLines, ok := gen.compileMatchPattern(clauseCtx, rescue.Clauses[0].Pattern, "subject", "runtime.Value")
	if !ok {
		t.Fatalf("expected rescue binding compilation to succeed: reason=%q", clauseCtx.reason)
	}
	if len(bindLines) == 0 {
		t.Fatalf("expected rescue binding lines to be emitted")
	}
	binding, ok := clauseCtx.lookup("thing")
	if !ok {
		t.Fatalf("expected rescue binding to be recorded")
	}
	if binding.GoType != "*Thing_a" {
		t.Fatalf("expected imported shadowed rescue binding to stay native, got goType=%q type=%q lines=%q", binding.GoType, normalizeTypeExprString(gen, ctx.packageName, binding.TypeExpr), strings.Join(bindLines, "\n"))
	}
	if gotPkg := gen.resolvedTypeExprPackage(ctx.packageName, binding.TypeExpr); gotPkg != "demo.remote" {
		t.Fatalf("expected imported shadowed rescue binding type to stay remote, got pkg=%q type=%q", gotPkg, normalizeTypeExprString(gen, ctx.packageName, binding.TypeExpr))
	}

	bodyBlock, ok := rescue.Clauses[0].Body.(*ast.BlockExpression)
	if !ok || bodyBlock == nil {
		t.Fatalf("expected rescue clause body block")
	}
	var recoveredAssign *ast.AssignmentExpression
	for _, stmt := range bodyBlock.Body {
		assign, ok := stmt.(*ast.AssignmentExpression)
		if !ok || assign == nil {
			continue
		}
		left, ok := assign.Left.(*ast.Identifier)
		if ok && left != nil && left.Name == "recovered" {
			recoveredAssign = assign
			break
		}
	}
	if recoveredAssign == nil {
		t.Fatalf("expected recovered assignment in rescue clause")
	}
	_, _, _, ok = gen.compileExprLines(clauseCtx, recoveredAssign, "")
	if !ok {
		t.Fatalf("expected recovered assignment to compile: reason=%q", clauseCtx.reason)
	}
	recovered, ok := clauseCtx.lookup("recovered")
	if !ok {
		t.Fatalf("expected recovered binding to be recorded")
	}
	if recovered.GoType != "*Thing_a" {
		t.Fatalf("expected recovered join binding to stay native, got goType=%q type=%q", recovered.GoType, normalizeTypeExprString(gen, ctx.packageName, recovered.TypeExpr))
	}
	if gotPkg := gen.resolvedTypeExprPackage(ctx.packageName, recovered.TypeExpr); gotPkg != "demo.remote" {
		t.Fatalf("expected recovered join binding type to stay remote, got pkg=%q type=%q", gotPkg, normalizeTypeExprString(gen, ctx.packageName, recovered.TypeExpr))
	}
}
