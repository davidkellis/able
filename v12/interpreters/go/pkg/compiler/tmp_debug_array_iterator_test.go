package compiler

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/driver"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/typechecker"
)

func TestTmpDebugArrayIteratorShapes(t *testing.T) {
	moduleRoot, workDir := compilerTestWorkDir(t, "ablec-debug-array-iterator-shapes")
	entryPath := filepath.Join(workDir, "main.able")
	source := strings.Join([]string{
		"package demo",
		"",
		"import able.collections.linked_list.{LinkedList}",
		"",
		"fn main() -> i64 {",
		"  values: LinkedList i32 = LinkedList.new()",
		"  values.push_back(1)",
		"  values.push_back(2)",
		"  values.push_back(3)",
		"  iter := values.lazy().map<i64>({ value => (value as i64) * 3_i64 }).filter({ value => value >= 6_i64 })",
		"  iter.collect<Array i64>().reduce<i64>(0_i64, { acc, value => acc + value })",
		"}",
		"",
	}, "\n")
	if err := os.WriteFile(filepath.Join(workDir, "package.yml"), []byte("name: demo\n"), 0o600); err != nil {
		t.Fatalf("write package.yml: %v", err)
	}
	if err := os.WriteFile(entryPath, []byte(source), 0o600); err != nil {
		t.Fatalf("write main.able: %v", err)
	}
	searchPaths, err := buildExecSearchPaths(entryPath, workDir, interpreter.FixtureManifest{})
	if err != nil {
		t.Fatalf("build search paths: %v", err)
	}
	loader, err := driver.NewLoader(searchPaths)
	if err != nil {
		t.Fatalf("loader init: %v", err)
	}
	t.Cleanup(func() { loader.Close() })
	program, err := loader.Load(entryPath)
	if err != nil {
		t.Fatalf("load program: %v", err)
	}
	checker := typechecker.NewProgramChecker()
	check, err := checker.Check(program)
	if err != nil {
		t.Fatalf("typecheck: %v", err)
	}
	gen := newGenerator(Options{PackageName: "main"})
	gen.setTypecheckInference(check.Inferred)
	if err := gen.collect(program); err != nil {
		t.Fatalf("collect: %v", err)
	}
	dynamicReport, err := DetectDynamicFeatures(program)
	if err != nil {
		t.Fatalf("detect dynamic features: %v", err)
	}
	gen.setDynamicFeatureReport(dynamicReport)
	gen.resolveCompileabilityFixedPoint()
	preRenderFallbacks := gen.collectFallbacks()
	files, err := gen.render()
	if err != nil {
		t.Fatalf("render: %v", err)
	}
	resultFallbacks := gen.collectFallbacks()
	_ = moduleRoot
	compiled := string(files["compiled.go"])
	var parts []string
	parts = append(parts, "pre-render fallbacks:")
	for _, fb := range preRenderFallbacks {
		parts = append(parts, fb.Name+": "+fb.Reason)
	}
	parts = append(parts, "post-render fallbacks:")
	for _, fb := range resultFallbacks {
		parts = append(parts, fb.Name+": "+fb.Reason)
	}
	parts = append(parts, "matching function infos:")
	for _, info := range gen.sortedFunctionInfos() {
		if info == nil {
			continue
		}
		if !strings.Contains(info.Name, "impl Iterator for ArrayIterator<T>.next") &&
			!strings.Contains(info.Name, "impl Iterator for ArrayIterator<T>.filter_map") &&
			!strings.Contains(info.Name, "impl Iterable for Array<T>.iterator") {
			continue
		}
		line := info.Name + " | go=" + info.GoName + " | ret=" + info.ReturnType + " | compileable=" + fmt.Sprintf("%t", info.Compileable) + " | reason=" + info.Reason + " | bindings=" + formatDebugBindings(info.TypeBindings)
		if impl := gen.implMethodByInfo[info]; impl != nil {
			line += " | implTarget=" + debugTypeExprString(impl.TargetType)
			line += " | specTarget=" + debugTypeExprString(gen.specializedImplTargetType(impl, info.TypeBindings))
			line += " | rawParam0=" + debugTypeExprString(gen.functionParamTypeExpr(info, 0))
			line += " | implTargetKind=" + fmt.Sprintf("%T", impl.TargetType)
			if generic, ok := impl.TargetType.(*ast.GenericTypeExpression); ok && generic != nil {
				line += " | implTargetBase=" + debugTypeExprString(generic.Base)
			}
			line += " | implIface=" + impl.InterfaceName
			if info.Definition != nil && len(info.Definition.Params) > 0 && info.Definition.Params[0] != nil && info.Definition.Params[0].ParamType != nil {
				line += " | defSelf=" + debugTypeExprString(info.Definition.Params[0].ParamType)
			}
			if info.Definition != nil && info.Definition.ReturnType != nil {
				line += " | defRet=" + debugTypeExprString(info.Definition.ReturnType)
			}
			if len(info.Params) > 0 && info.Params[0].TypeExpr != nil {
				line += " | param0=" + debugTypeExprString(info.Params[0].TypeExpr)
			}
			contextBindings := gen.compileContextTypeBindings(info)
			concreteTarget := gen.specializedImplTargetType(impl, contextBindings)
			if concreteTarget == nil {
				concreteTarget = impl.TargetType
			}
			ifaceBindings := gen.implTypeBindings(impl.InterfaceName, impl.InterfaceGenerics, impl.InterfaceArgs, concreteTarget)
			selfTarget := gen.implSelfTargetType(info.Package, concreteTarget, ifaceBindings)
			line += " | concreteTarget=" + debugTypeExprString(concreteTarget)
			line += " | ctxBindings=" + formatDebugBindings(contextBindings)
			line += " | ifaceBindings=" + formatDebugBindings(ifaceBindings)
			line += " | selfTarget=" + debugTypeExprString(selfTarget)
		}
		parts = append(parts, line)
	}
	for _, marker := range []string{
		"func __able_compiled_fn_main",
		"func __able_compiled_impl_Iterator_next_",
		"func __able_compiled_impl_Iterable_iterator_",
		"func __able_compiled_iface_Iterator_filter_default",
		"func __able_compiled_iface_Enumerable__map_default",
		"func __able_compiled_impl_Enumerable_iterator_",
	} {
		if idx := strings.Index(compiled, marker); idx >= 0 {
			end := idx + 1800
			if end > len(compiled) {
				end = len(compiled)
			}
			parts = append(parts, compiled[idx:end])
		}
	}
	t.Fatalf("%s", strings.Join(parts, "\n\n---\n\n"))
}

func formatDebugBindings(bindings map[string]ast.TypeExpression) string {
	if len(bindings) == 0 {
		return "{}"
	}
	keys := make([]string, 0, len(bindings))
	for name := range bindings {
		keys = append(keys, name)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, name := range keys {
		parts = append(parts, name+"="+debugTypeExprString(bindings[name]))
	}
	return "{" + strings.Join(parts, ", ") + "}"
}

func debugTypeExprString(expr ast.TypeExpression) string {
	if expr == nil {
		return "<nil>"
	}
	switch typed := expr.(type) {
	case *ast.SimpleTypeExpression:
		if typed.Name == nil {
			return "<simple:nil>"
		}
		return typed.Name.Name
	case *ast.GenericTypeExpression:
		args := make([]string, 0, len(typed.Arguments))
		for _, arg := range typed.Arguments {
			args = append(args, debugTypeExprString(arg))
		}
		return debugTypeExprString(typed.Base) + "<" + strings.Join(args, ", ") + ">"
	case *ast.FunctionTypeExpression:
		params := make([]string, 0, len(typed.ParamTypes))
		for _, param := range typed.ParamTypes {
			params = append(params, debugTypeExprString(param))
		}
		return "fn(" + strings.Join(params, ", ") + ") -> " + debugTypeExprString(typed.ReturnType)
	case *ast.NullableTypeExpression:
		return "?" + debugTypeExprString(typed.InnerType)
	case *ast.ResultTypeExpression:
		return "!" + debugTypeExprString(typed.InnerType)
	case *ast.UnionTypeExpression:
		members := make([]string, 0, len(typed.Members))
		for _, member := range typed.Members {
			members = append(members, debugTypeExprString(member))
		}
		return strings.Join(members, " | ")
	case *ast.WildcardTypeExpression:
		return "_"
	default:
		return fmt.Sprintf("%T", expr)
	}
}
