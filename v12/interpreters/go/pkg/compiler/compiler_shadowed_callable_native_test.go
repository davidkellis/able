package compiler

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/typechecker"
)

func loadShadowedCallableJoinProgram(t *testing.T) *driver.Program {
	t.Helper()
	root := t.TempDir()
	if err := os.WriteFile(filepath.Join(root, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(root, "remote"), 0o755); err != nil {
		t.Fatalf("mkdir remote: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "main.able"), []byte(strings.Join([]string{
		"package demo",
		"",
		"import demo.remote.{Thing::RemoteThing}",
		"",
		"struct Thing { local: i32 }",
		"",
		"fn main() -> i32 {",
		"  mixed := if true {",
		"    fn() -> RemoteThing { RemoteThing { remote: 1 } }",
		"  } else {",
		"    fn() -> Thing { Thing { local: 2 } }",
		"  }",
		"  mixed match {",
		"    case build: (() -> RemoteThing) => build().remote,",
		"    case build: (() -> Thing) => build().local",
		"  }",
		"}",
		"",
	}, "\n")), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "remote", "module.able"), []byte("struct Thing { remote: i32 }\n"), 0o600); err != nil {
		t.Fatalf("write remote/module.able: %v", err)
	}

	loader, err := driver.NewLoader(nil)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })

	program, err := loader.Load(filepath.Join(root, "main.able"))
	if err != nil {
		t.Fatalf("load program: %v", err)
	}
	return program
}

func shadowedCallableJoinMainParts(t *testing.T, program *driver.Program) (*generator, *compileContext, *ast.AssignmentExpression, *ast.MatchExpression, string, *ast.TypedPattern, ast.Expression) {
	t.Helper()
	mainPkg := ""
	if program != nil && program.Entry != nil {
		mainPkg = program.Entry.Package
	}
	if mainPkg == "" {
		mainPkg = "demo"
	}
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
	dynamicReport, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("dynamic features: %v", err)
	}
	gen.setDynamicFeatureReport(dynamicReport)
	gen.resolveCompileabilityFixedPoint()

	var assign *ast.AssignmentExpression
	var matchExpr *ast.MatchExpression
	for _, mod := range program.Modules {
		for _, stmt := range mod.AST.Body {
			fn, ok := stmt.(*ast.FunctionDefinition)
			if !ok || fn == nil || fn.ID == nil || fn.ID.Name != "main" || fn.Body == nil {
				continue
			}
			for _, bodyStmt := range fn.Body.Body {
				switch typed := bodyStmt.(type) {
				case *ast.AssignmentExpression:
					assign = typed
				case *ast.MatchExpression:
					matchExpr = typed
				}
			}
		}
	}
	if assign == nil || assign.Right == nil || matchExpr == nil || len(matchExpr.Clauses) == 0 {
		t.Fatalf("could not locate mixed assignment and match expression in main")
	}
	pattern, ok := matchExpr.Clauses[0].Pattern.(*ast.TypedPattern)
	if !ok || pattern == nil {
		t.Fatalf("expected first clause pattern to be a typed pattern, got %T", matchExpr.Clauses[0].Pattern)
	}
	otherPattern, ok := matchExpr.Clauses[1].Pattern.(*ast.TypedPattern)
	if !ok || otherPattern == nil || otherPattern.TypeAnnotation == nil {
		t.Fatalf("expected second clause pattern to be a typed pattern, got %T", matchExpr.Clauses[1].Pattern)
	}

	temp := 0
	ctx := &compileContext{
		locals:      make(map[string]paramInfo),
		packageName: mainPkg,
		temps:       &temp,
		returnType:  "int32",
	}
	if _, ok := assign.Right.(*ast.IfExpression); !ok {
		t.Fatalf("expected mixed assignment right-hand side to be an if expression")
	}
	leftTypeExpr := gen.lowerNormalizedTypeExpr(ctx, pattern.TypeAnnotation)
	rightTypeExpr := gen.lowerNormalizedTypeExpr(ctx, otherPattern.TypeAnnotation)
	leftType, ok := gen.joinCarrierTypeFromTypeExpr(ctx, leftTypeExpr)
	if !ok || leftType == "" {
		t.Fatalf("recover left callable carrier from %q", typeExpressionToString(leftTypeExpr))
	}
	rightType, ok := gen.joinCarrierTypeFromTypeExpr(ctx, rightTypeExpr)
	if !ok || rightType == "" {
		t.Fatalf("recover right callable carrier from %q", typeExpressionToString(rightTypeExpr))
	}
	if leftType == "__able_fn_void_to_runtime_Value" || rightType == "__able_fn_void_to_runtime_Value" {
		t.Fatalf("expected imported/local callable pattern carriers to stay native, got left=%q right=%q", leftType, rightType)
	}
	subjectType, ok := gen.joinResultType(ctx, leftType, rightType)
	if !ok || subjectType == "" {
		t.Fatalf("join callable carrier types %q and %q", leftType, rightType)
	}
	subjectTypeExpr := ast.UnionT(leftTypeExpr, rightTypeExpr)
	ctx.setLocalBinding("mixed", paramInfo{
		Name:     "mixed",
		GoName:   "mixed",
		GoType:   subjectType,
		TypeExpr: subjectTypeExpr,
	})
	ctx.matchSubjectTypeExpr = subjectTypeExpr
	return gen, ctx, assign, matchExpr, subjectType, pattern, matchExpr.Clauses[0].Body
}

func TestCompilerShadowedCallableTypedPatternBindingRecordsLocal(t *testing.T) {
	program := loadShadowedCallableJoinProgram(t)
	gen, ctx, _, _, subjectType, pattern, _ := shadowedCallableJoinMainParts(t, program)

	lines, ok := gen.compileNativeUnionTypedPatternBindings(ctx, "mixed", subjectType, pattern)
	if !ok {
		union := gen.nativeUnionInfoForGoType(subjectType)
		t.Fatalf("compile native union typed-pattern bindings for subjectType=%q union=%v pattern=%q: %s", subjectType, union != nil, typeExpressionToString(pattern.TypeAnnotation), ctx.reason)
	}
	if len(lines) == 0 {
		t.Fatalf("expected native union typed-pattern bindings to emit code")
	}

	binding, ok := ctx.lookup("build")
	if !ok {
		t.Fatalf("expected typed-pattern binding to record local callable binding")
	}
	if binding.GoType == "" || binding.GoType == "runtime.Value" || binding.GoType == "any" {
		t.Fatalf("expected typed-pattern binding to stay on a native callable carrier, got %q", binding.GoType)
	}
	if _, ok := binding.TypeExpr.(*ast.FunctionTypeExpression); !ok {
		t.Fatalf("expected typed-pattern binding to keep a function type expr, got %T", binding.TypeExpr)
	}
}

func TestCompilerShadowedCallableTypedPatternBodyCompilesWithRecordedBinding(t *testing.T) {
	program := loadShadowedCallableJoinProgram(t)
	gen, ctx, _, _, subjectType, pattern, body := shadowedCallableJoinMainParts(t, program)

	if _, ok := gen.compileNativeUnionTypedPatternBindings(ctx, "mixed", subjectType, pattern); !ok {
		union := gen.nativeUnionInfoForGoType(subjectType)
		t.Fatalf("compile native union typed-pattern bindings for subjectType=%q union=%v pattern=%q: %s", subjectType, union != nil, typeExpressionToString(pattern.TypeAnnotation), ctx.reason)
	}
	bodyLines, bodyExpr, bodyType, ok := gen.compileTailExpression(ctx, "", body)
	if !ok {
		t.Fatalf("compile typed-pattern clause body: %s", ctx.reason)
	}
	joined := strings.Join(append(bodyLines, bodyExpr), "\n")
	if bodyType != "int32" {
		t.Fatalf("expected typed-pattern clause body to stay native i32, got %q:\n%s", bodyType, joined)
	}
	if strings.Contains(joined, "__able_call_value(") {
		t.Fatalf("expected typed-pattern clause body to avoid dynamic callable dispatch:\n%s", joined)
	}
}

func TestCompilerShadowedCallableJoinAssignmentKeepsNativeUnionCarrier(t *testing.T) {
	program := loadShadowedCallableJoinProgram(t)
	gen, ctx, assign, _, subjectType, _, _ := shadowedCallableJoinMainParts(t, program)
	delete(ctx.locals, "mixed")
	ctx.matchSubjectTypeExpr = nil

	if _, _, valueType, ok := gen.compileAssignment(ctx, assign); !ok {
		t.Fatalf("compile mixed assignment: %s", ctx.reason)
	} else if valueType != subjectType {
		t.Fatalf("expected mixed assignment valueType=%q, got %q", subjectType, valueType)
	}

	mixed, ok := ctx.lookup("mixed")
	if !ok {
		t.Fatalf("expected mixed assignment to record local binding")
	}
	if mixed.GoType != subjectType {
		t.Fatalf("expected mixed binding GoType=%q, got %q", subjectType, mixed.GoType)
	}
	if gen.nativeUnionInfoForGoType(mixed.GoType) == nil {
		t.Fatalf("expected mixed binding GoType=%q to stay on a native union carrier", mixed.GoType)
	}
}

func TestCompilerShadowedCallableJoinMatchExpressionCompilesNatively(t *testing.T) {
	program := loadShadowedCallableJoinProgram(t)
	gen, ctx, assign, matchExpr, subjectType, _, _ := shadowedCallableJoinMainParts(t, program)
	delete(ctx.locals, "mixed")
	ctx.matchSubjectTypeExpr = nil

	if _, _, valueType, ok := gen.compileAssignment(ctx, assign); !ok {
		t.Fatalf("compile mixed assignment: %s", ctx.reason)
	} else if valueType != subjectType {
		t.Fatalf("expected mixed assignment valueType=%q, got %q", subjectType, valueType)
	}

	lines, expr, goType, ok := gen.compileMatchExpression(ctx, matchExpr, "")
	if !ok {
		t.Fatalf("compile mixed callable match expression: %s", ctx.reason)
	}
	joined := strings.Join(append(lines, expr), "\n")
	if goType != "int32" {
		t.Fatalf("expected mixed callable match expression type=int32, got %q:\n%s", goType, joined)
	}
	for _, fragment := range []string{"__able_call_value(", "__able_try_cast(", "bridge.MatchType("} {
		if strings.Contains(joined, fragment) {
			t.Fatalf("expected mixed callable match expression to avoid %q:\n%s", fragment, joined)
		}
	}
}
