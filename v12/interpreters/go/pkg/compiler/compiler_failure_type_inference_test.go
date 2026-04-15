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
