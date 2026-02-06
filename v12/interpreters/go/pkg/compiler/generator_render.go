package compiler

import (
	"bytes"
	"fmt"
	"go/format"
	"sort"
	"strings"

	"able/interpreter-go/pkg/ast"
)

func (g *generator) render() (map[string][]byte, error) {
	files := make(map[string][]byte)
	compiled, err := g.renderCompiled()
	if err != nil {
		return nil, err
	}
	files["compiled.go"] = compiled
	if g.opts.EmitMain {
		mainSrc, err := g.renderMain()
		if err != nil {
			return nil, err
		}
		files["main.go"] = mainSrc
	}
	return files, nil
}

func (g *generator) renderCompiled() ([]byte, error) {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", g.opts.PackageName)

	imports := g.importsForCompiled()
	if len(imports) > 0 {
		fmt.Fprintf(&buf, "import (\n")
		for _, imp := range imports {
			fmt.Fprintf(&buf, "\t%q\n", imp)
		}
		fmt.Fprintf(&buf, ")\n\n")
	}

	if g.hasFunctions() {
		fmt.Fprintf(&buf, "var __able_runtime *bridge.Runtime\n\n")
		if len(g.diagNodes) > 0 {
			for _, info := range g.diagNodes {
				initExpr := ""
				switch {
				case info.CallName != "":
					initExpr = fmt.Sprintf("&ast.FunctionCall{Callee: ast.NewIdentifier(%q)}", info.CallName)
				case info.CallMember != "":
					initExpr = fmt.Sprintf("&ast.FunctionCall{Callee: ast.NewMemberAccessExpression(ast.NewIdentifier(\"\"), ast.NewIdentifier(%q))}", info.CallMember)
				default:
					goType := info.GoType
					if strings.HasPrefix(goType, "*") {
						goType = "&" + strings.TrimPrefix(goType, "*")
					}
					initExpr = fmt.Sprintf("%s{}", goType)
				}
				fmt.Fprintf(&buf, "var %s = %s\n", info.Name, initExpr)
			}
			fmt.Fprintf(&buf, "\n")
		}
		if len(g.awaitExprs) > 0 {
			for _, name := range g.awaitExprs {
				fmt.Fprintf(&buf, "var %s = &ast.AwaitExpression{}\n", name)
			}
			fmt.Fprintf(&buf, "\n")
		}
		g.renderRuntimeHelpers(&buf)
	}

	g.renderStructs(&buf)
	if g.hasFunctions() {
		g.renderStructConverters(&buf)
		g.renderCompiledMethods(&buf)
		g.renderCompiledFunctions(&buf)
		g.renderMethodWrappers(&buf)
		g.renderWrappers(&buf)
		g.renderOverloadDispatchers(&buf)
		g.renderMethodThunks(&buf)
		g.renderRegister(&buf)
	}

	return formatSource(buf.Bytes())
}

func (g *generator) renderRuntimeHelpers(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "func __able_index(obj runtime.Value, idx runtime.Value) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.Index(__able_runtime, obj, idx)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_range(start runtime.Value, end runtime.Value, inclusive bool) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.Range(__able_runtime, start, end, inclusive)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_array_values(iterable runtime.Value) ([]runtime.Value, bool) {\n")
	fmt.Fprintf(buf, "\tarr, ok := iterable.(*runtime.ArrayValue)\n")
	fmt.Fprintf(buf, "\tif !ok {\n")
	fmt.Fprintf(buf, "\t\treturn nil, false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tvalues, err := bridge.ArrayElements(__able_runtime, arr)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\treturn values, true\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_resolve_iterator(iterable runtime.Value) *runtime.IteratorValue {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tit, err := bridge.ResolveIterator(__able_runtime, iterable)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif it == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"iterator is nil\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn it\n")
	fmt.Fprintf(buf, "}\n\n")
	if g.needsIterator {
		g.renderIteratorHelpers(buf)
	}
	fmt.Fprintf(buf, "func __able_index_set(obj runtime.Value, idx runtime.Value, value runtime.Value) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.IndexAssign(__able_runtime, obj, idx, value)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_member_set(obj runtime.Value, member runtime.Value, value runtime.Value) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.MemberAssign(__able_runtime, obj, member, value)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_member_get(obj runtime.Value, member runtime.Value) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.MemberGet(__able_runtime, obj, member)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_member_get_method(obj runtime.Value, member runtime.Value) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.MemberGetPreferMethods(__able_runtime, obj, member)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_call_value(fn runtime.Value, args []runtime.Value, call *ast.FunctionCall) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.CallValueWithNode(__able_runtime, fn, args, call)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_call_named(name string, args []runtime.Value, call *ast.FunctionCall) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.CallNamedWithNode(__able_runtime, name, args, call)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_global_get(name string, node ast.Node) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.Get(__able_runtime, name)\n")
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\tbridge.RaiseRuntimeErrorWithContext(__able_runtime, node, err)\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_global_set(name string, value runtime.Value, node ast.Node) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif err := bridge.Assign(__able_runtime, name, value); err != nil {\n")
	fmt.Fprintf(buf, "\t\tbridge.RaiseRuntimeErrorWithContext(__able_runtime, node, err)\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn value\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_push_call_frame(call *ast.FunctionCall) {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\treturn\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbridge.PushCallFrame(__able_runtime, call)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_pop_call_frame() {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\treturn\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbridge.PopCallFrame(__able_runtime)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_spawn(task func(*runtime.Environment) (runtime.Value, error)) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tfuture, err := bridge.Spawn(__able_runtime, task)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif future == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn future\n")
	fmt.Fprintf(buf, "}\n\n")
	if len(g.awaitExprs) > 0 {
		fmt.Fprintf(buf, "func __able_await(expr *ast.AwaitExpression, iterable runtime.Value) runtime.Value {\n")
		fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
		fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tval, err := bridge.Await(__able_runtime, expr, iterable)\n")
		fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
		fmt.Fprintf(buf, "\tif val == nil {\n")
		fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn val\n")
		fmt.Fprintf(buf, "}\n\n")
	}
	fmt.Fprintf(buf, "type __able_break struct {\n")
	fmt.Fprintf(buf, "\tvalue runtime.Value\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "type __able_continue_signal struct{}\n\n")
	fmt.Fprintf(buf, "type __able_return struct {\n")
	fmt.Fprintf(buf, "\tvalue any\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "type __able_break_label_signal struct {\n")
	fmt.Fprintf(buf, "\tlabel string\n")
	fmt.Fprintf(buf, "\tvalue runtime.Value\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "type __able_continue_label_signal struct {\n")
	fmt.Fprintf(buf, "\tlabel string\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_break_value(value runtime.Value) {\n")
	fmt.Fprintf(buf, "\tpanic(__able_break{value: value})\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_break_label(label string, value runtime.Value) {\n")
	fmt.Fprintf(buf, "\tpanic(__able_break_label_signal{label: label, value: value})\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_continue() {\n")
	fmt.Fprintf(buf, "\tpanic(__able_continue_signal{})\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_continue_label(label string) {\n")
	fmt.Fprintf(buf, "\tpanic(__able_continue_label_signal{label: label})\n")
	fmt.Fprintf(buf, "}\n\n")
	if len(g.diagNodes) > 0 {
		fmt.Fprintf(buf, "func __able_register_diag_nodes() {\n")
		fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
		fmt.Fprintf(buf, "\t\treturn\n")
		fmt.Fprintf(buf, "\t}\n")
		for _, info := range g.diagNodes {
			span := info.Span
			fmt.Fprintf(buf, "\tast.SetSpan(%s, ast.Span{Start: ast.Position{Line: %d, Column: %d}, End: ast.Position{Line: %d, Column: %d}})\n", info.Name, span.Start.Line, span.Start.Column, span.End.Line, span.End.Column)
			if info.Origin != "" {
				fmt.Fprintf(buf, "\tbridge.RegisterNodeOrigin(__able_runtime, %s, %q)\n", info.Name, info.Origin)
			}
		}
		fmt.Fprintf(buf, "}\n\n")
	}
	fmt.Fprintf(buf, "func __able_panic_on_error(err error) {\n")
	fmt.Fprintf(buf, "\tif err == nil {\n")
	fmt.Fprintf(buf, "\t\treturn\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tpanic(err)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_is_nil(val runtime.Value) bool {\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn true\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tswitch val.(type) {\n")
	fmt.Fprintf(buf, "\tcase runtime.NilValue, *runtime.NilValue:\n")
	fmt.Fprintf(buf, "\t\treturn true\n")
	fmt.Fprintf(buf, "\tdefault:\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_is_error(val runtime.Value) bool {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn bridge.IsError(__able_runtime, val)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_truthy(val runtime.Value) bool {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn bridge.IsTruthy(__able_runtime, val)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_error_to_struct(err runtime.ErrorValue) *runtime.StructInstanceValue {\n")
	fmt.Fprintf(buf, "\tfields := make(map[string]runtime.Value)\n")
	fmt.Fprintf(buf, "\tif err.Payload != nil {\n")
	fmt.Fprintf(buf, "\t\tfor k, v := range err.Payload {\n")
	fmt.Fprintf(buf, "\t\t\tfields[k] = v\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tfields[\"message\"] = runtime.StringValue{Val: err.Message}\n")
	fmt.Fprintf(buf, "\treturn &runtime.StructInstanceValue{Fields: fields}\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_struct_instance(value runtime.Value) *runtime.StructInstanceValue {\n")
	fmt.Fprintf(buf, "\tswitch v := value.(type) {\n")
	fmt.Fprintf(buf, "\tcase *runtime.StructInstanceValue:\n")
	fmt.Fprintf(buf, "\t\treturn v\n")
	fmt.Fprintf(buf, "\tcase runtime.ErrorValue:\n")
	fmt.Fprintf(buf, "\t\treturn __able_error_to_struct(v)\n")
	fmt.Fprintf(buf, "\tcase *runtime.ErrorValue:\n")
	fmt.Fprintf(buf, "\t\tif v == nil {\n")
	fmt.Fprintf(buf, "\t\t\treturn nil\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\treturn __able_error_to_struct(*v)\n")
	fmt.Fprintf(buf, "\tdefault:\n")
	fmt.Fprintf(buf, "\t\treturn nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_error_value(value runtime.Value) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn bridge.ErrorValue(__able_runtime, value)\n")
	fmt.Fprintf(buf, "}\n\n")
	if g.needsAst {
		fmt.Fprintf(buf, "func __able_cast(value runtime.Value, typeExpr ast.TypeExpression) runtime.Value {\n")
		fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
		fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tval, err := bridge.Cast(__able_runtime, typeExpr, value)\n")
		fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
		fmt.Fprintf(buf, "\tif val == nil {\n")
		fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn val\n")
		fmt.Fprintf(buf, "}\n\n")
		fmt.Fprintf(buf, "func __able_try_cast(value runtime.Value, typeExpr ast.TypeExpression) (runtime.Value, bool) {\n")
		fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
		fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tval, ok, err := bridge.MatchType(__able_runtime, typeExpr, value)\n")
		fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
		fmt.Fprintf(buf, "\tif !ok {\n")
		fmt.Fprintf(buf, "\t\treturn nil, false\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tif val == nil {\n")
		fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, true\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\treturn val, true\n")
		fmt.Fprintf(buf, "}\n\n")
	}
	fmt.Fprintf(buf, "func __able_stringify(val runtime.Value) string {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tstr, err := bridge.Stringify(__able_runtime, val)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\treturn str\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_raise_division_by_zero(node ast.Node) {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbridge.RaiseWithContext(__able_runtime, node, bridge.DivisionByZeroError(__able_runtime))\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_raise_overflow() {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbridge.Raise(bridge.OverflowError(__able_runtime, \"integer overflow\"))\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_raise_shift_out_of_range(shift int64, node ast.Node) {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tbridge.RaiseWithContext(__able_runtime, node, bridge.ShiftOutOfRangeError(__able_runtime, shift))\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_raise_return_type_mismatch(node ast.Node, expected string, actual string) {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\terr := fmt.Errorf(\"Return type mismatch: expected %%s, got %%s\", expected, actual)\n")
	fmt.Fprintf(buf, "\tbridge.RaiseRuntimeErrorWithContext(__able_runtime, node, err)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_signed_bounds(bits int) (int64, int64) {\n")
	fmt.Fprintf(buf, "\tif bits >= 64 {\n")
	fmt.Fprintf(buf, "\t\tmax := int64(^uint64(0) >> 1)\n")
	fmt.Fprintf(buf, "\t\treturn -max - 1, max\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tmax := int64(1<<(bits-1)) - 1\n")
	fmt.Fprintf(buf, "\tmin := -int64(1 << (bits - 1))\n")
	fmt.Fprintf(buf, "\treturn min, max\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_unsigned_max(bits int) uint64 {\n")
	fmt.Fprintf(buf, "\tif bits >= 64 {\n")
	fmt.Fprintf(buf, "\t\treturn ^uint64(0)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn (uint64(1) << uint(bits)) - 1\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_shift_left_signed(value int64, shift int64, bits int, node ast.Node) int64 {\n")
	fmt.Fprintf(buf, "\tif shift < 0 || shift >= int64(bits) {\n")
	fmt.Fprintf(buf, "\t\t__able_raise_shift_out_of_range(shift, node)\n")
	fmt.Fprintf(buf, "\t\treturn 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tmin, max := __able_signed_bounds(bits)\n")
	fmt.Fprintf(buf, "\tif shift > 0 {\n")
	fmt.Fprintf(buf, "\t\tif value > max>>shift || value < min>>shift {\n")
	fmt.Fprintf(buf, "\t\t\t__able_raise_overflow()\n")
	fmt.Fprintf(buf, "\t\t\treturn 0\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn value << shift\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_shift_right_signed(value int64, shift int64, bits int, node ast.Node) int64 {\n")
	fmt.Fprintf(buf, "\tif shift < 0 || shift >= int64(bits) {\n")
	fmt.Fprintf(buf, "\t\t__able_raise_shift_out_of_range(shift, node)\n")
	fmt.Fprintf(buf, "\t\treturn 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn value >> shift\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_shift_left_unsigned(value uint64, shift uint64, bits int, node ast.Node) uint64 {\n")
	fmt.Fprintf(buf, "\tif shift > uint64(^uint64(0)>>1) {\n")
	fmt.Fprintf(buf, "\t\t__able_raise_shift_out_of_range(0, node)\n")
	fmt.Fprintf(buf, "\t\treturn 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\ts := int64(shift)\n")
	fmt.Fprintf(buf, "\tif s < 0 || s >= int64(bits) {\n")
	fmt.Fprintf(buf, "\t\t__able_raise_shift_out_of_range(s, node)\n")
	fmt.Fprintf(buf, "\t\treturn 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tmax := __able_unsigned_max(bits)\n")
	fmt.Fprintf(buf, "\tif s > 0 && value > max>>s {\n")
	fmt.Fprintf(buf, "\t\t__able_raise_overflow()\n")
	fmt.Fprintf(buf, "\t\treturn 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn value << s\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_shift_right_unsigned(value uint64, shift uint64, bits int, node ast.Node) uint64 {\n")
	fmt.Fprintf(buf, "\tif shift > uint64(^uint64(0)>>1) {\n")
	fmt.Fprintf(buf, "\t\t__able_raise_shift_out_of_range(0, node)\n")
	fmt.Fprintf(buf, "\t\treturn 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\ts := int64(shift)\n")
	fmt.Fprintf(buf, "\tif s < 0 || s >= int64(bits) {\n")
	fmt.Fprintf(buf, "\t\t__able_raise_shift_out_of_range(s, node)\n")
	fmt.Fprintf(buf, "\t\treturn 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn value >> s\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_binary_op(op string, left runtime.Value, right runtime.Value) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.ApplyBinaryOperator(__able_runtime, op, left, right)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_unary_op(op string, operand runtime.Value) runtime.Value {\n")
	fmt.Fprintf(buf, "\tif __able_runtime == nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tval, err := bridge.ApplyUnaryOperator(__able_runtime, op, operand)\n")
	fmt.Fprintf(buf, "\t__able_panic_on_error(err)\n")
	fmt.Fprintf(buf, "\tif val == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn val\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_divmod_signed(a int64, b int64, node ast.Node) (int64, int64) {\n")
	fmt.Fprintf(buf, "\tif b == 0 {\n")
	fmt.Fprintf(buf, "\t\t__able_raise_division_by_zero(node)\n")
	fmt.Fprintf(buf, "\t\treturn 0, 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tq := a / b\n")
	fmt.Fprintf(buf, "\tr := a %% b\n")
	fmt.Fprintf(buf, "\tif r < 0 {\n")
	fmt.Fprintf(buf, "\t\tif b > 0 {\n")
	fmt.Fprintf(buf, "\t\t\tq -= 1\n")
	fmt.Fprintf(buf, "\t\t\tr += b\n")
	fmt.Fprintf(buf, "\t\t} else {\n")
	fmt.Fprintf(buf, "\t\t\tq += 1\n")
	fmt.Fprintf(buf, "\t\t\tr -= b\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn q, r\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_divmod_unsigned(a uint64, b uint64, node ast.Node) (uint64, uint64) {\n")
	fmt.Fprintf(buf, "\tif b == 0 {\n")
	fmt.Fprintf(buf, "\t\t__able_raise_division_by_zero(node)\n")
	fmt.Fprintf(buf, "\t\treturn 0, 0\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn a / b, a %% b\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) importsForCompiled() []string {
	importSet := map[string]struct{}{}
	needsRuntime := g.hasFunctions() || g.structUsesRuntimeValue()
	if g.hasFunctions() {
		importSet["fmt"] = struct{}{}
		importSet["able/interpreter-go/pkg/compiler/bridge"] = struct{}{}
		importSet["able/interpreter-go/pkg/ast"] = struct{}{}
		importSet["able/interpreter-go/pkg/interpreter"] = struct{}{}
	}
	if needsRuntime {
		importSet["able/interpreter-go/pkg/runtime"] = struct{}{}
	}
	if g.needsIterator {
		importSet["errors"] = struct{}{}
		importSet["sync"] = struct{}{}
	}
	imports := make([]string, 0, len(importSet))
	for imp := range importSet {
		imports = append(imports, imp)
	}
	sort.Strings(imports)
	return imports
}

func (g *generator) structUsesRuntimeValue() bool {
	for _, info := range g.structs {
		for _, field := range info.Fields {
			if field.GoType == "runtime.Value" {
				return true
			}
		}
	}
	return false
}

func (g *generator) renderStructs(buf *bytes.Buffer) {
	if len(g.structs) == 0 {
		return
	}
	for _, name := range g.sortedStructNames() {
		info := g.structs[name]
		if info == nil {
			continue
		}
		fmt.Fprintf(buf, "type %s struct {\n", info.GoName)
		for _, field := range info.Fields {
			fmt.Fprintf(buf, "\t%s %s\n", field.GoName, field.GoType)
		}
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderStructConverters(buf *bytes.Buffer) {
	if len(g.structs) == 0 {
		return
	}
	for _, name := range g.sortedStructNames() {
		info := g.structs[name]
		if info == nil {
			continue
		}
		g.renderStructFrom(buf, info)
		g.renderStructTo(buf, info)
	}
}

func (g *generator) renderStructFrom(buf *bytes.Buffer, info *structInfo) {
	fmt.Fprintf(buf, "func __able_struct_%s_from(value runtime.Value) (*%s, error) {\n", info.GoName, info.GoName)
	fmt.Fprintf(buf, "\tout := &%s{}\n", info.GoName)
	fmt.Fprintf(buf, "\tinst, ok := value.(*runtime.StructInstanceValue)\n")
	fmt.Fprintf(buf, "\tif !ok {\n")
	fmt.Fprintf(buf, "\t\treturn out, fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil || inst.Definition.Node.ID.Name != %q {\n", info.Name)
	fmt.Fprintf(buf, "\t\treturn out, fmt.Errorf(\"expected %s struct instance\")\n", info.Name)
	fmt.Fprintf(buf, "\t}\n")
	if len(info.Fields) == 0 {
		fmt.Fprintf(buf, "\treturn out, nil\n")
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	if info.Kind == ast.StructKindPositional {
		fmt.Fprintf(buf, "\tif len(inst.Positional) < %d {\n", len(info.Fields))
		fmt.Fprintf(buf, "\t\treturn out, fmt.Errorf(\"missing positional fields for %s\")\n", info.Name)
		fmt.Fprintf(buf, "\t}\n")
		for idx, field := range info.Fields {
			g.renderFieldFromPositional(buf, field, idx)
		}
	} else {
		fmt.Fprintf(buf, "\tif inst.Fields == nil {\n")
		fmt.Fprintf(buf, "\t\treturn out, fmt.Errorf(\"missing fields for %s\")\n", info.Name)
		fmt.Fprintf(buf, "\t}\n")
		for _, field := range info.Fields {
			g.renderFieldFromNamed(buf, field)
		}
	}
	fmt.Fprintf(buf, "\treturn out, nil\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderFieldFromNamed(buf *bytes.Buffer, field fieldInfo) {
	fmt.Fprintf(buf, "\t{\n")
	fmt.Fprintf(buf, "\t\tfieldValue, ok := inst.Fields[%q]\n", field.Name)
	fmt.Fprintf(buf, "\t\tif !ok {\n")
	fmt.Fprintf(buf, "\t\t\treturn out, fmt.Errorf(\"missing field %s\")\n", field.Name)
	fmt.Fprintf(buf, "\t\t}\n")
	g.renderValueConversion(buf, "\t\t", "fieldValue", field.GoType, "out."+field.GoName, "out")
	fmt.Fprintf(buf, "\t}\n")
}

func (g *generator) renderFieldFromPositional(buf *bytes.Buffer, field fieldInfo, idx int) {
	fmt.Fprintf(buf, "\t{\n")
	fmt.Fprintf(buf, "\t\tfieldValue := inst.Positional[%d]\n", idx)
	g.renderValueConversion(buf, "\t\t", "fieldValue", field.GoType, "out."+field.GoName, "out")
	fmt.Fprintf(buf, "\t}\n")
}

func (g *generator) renderStructTo(buf *bytes.Buffer, info *structInfo) {
	fmt.Fprintf(buf, "func __able_struct_%s_to(rt *bridge.Runtime, value *%s) (runtime.Value, error) {\n", info.GoName, info.GoName)
	fmt.Fprintf(buf, "\tif rt == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing runtime bridge\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing %s value\")\n", info.Name)
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tdef, err := rt.StructDefinition(%q)\n", info.Name)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t}\n")
	if info.Kind == ast.StructKindPositional {
		fmt.Fprintf(buf, "\tvalues := make([]runtime.Value, 0, %d)\n", len(info.Fields))
		for _, field := range info.Fields {
			g.renderValueToRuntime(buf, "value."+field.GoName, field.GoType, "values")
		}
		fmt.Fprintf(buf, "\treturn &runtime.StructInstanceValue{Definition: def, Positional: values}, nil\n")
	} else {
		fmt.Fprintf(buf, "\tfields := make(map[string]runtime.Value, %d)\n", len(info.Fields))
		for _, field := range info.Fields {
			g.renderValueToRuntimeNamed(buf, "value."+field.GoName, field.GoType, field.Name)
		}
		fmt.Fprintf(buf, "\treturn &runtime.StructInstanceValue{Definition: def, Fields: fields}, nil\n")
	}
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderCompiledFunctions(buf *bytes.Buffer) {
	for _, info := range g.sortedFunctionInfos() {
		if info == nil || !info.Compileable {
			continue
		}
		ctx := newCompileContext(info, g.functions, g.overloads)
		lines, retExpr, ok := g.compileBody(ctx, info)
		if !ok {
			continue
		}
		fmt.Fprintf(buf, "func __able_compiled_%s(", info.GoName)
		for i, param := range info.Params {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
		}
		resultName := "__able_result"
		fmt.Fprintf(buf, ") (%s %s) {\n", resultName, info.ReturnType)
		recoverValue := fmt.Sprintf("val, ok := ret.value.(%s); if !ok { panic(fmt.Errorf(\"compiler: return type mismatch\")) }; %s = val", info.ReturnType, resultName)
		if info.ReturnType == "runtime.Value" {
			recoverValue = fmt.Sprintf("if ret.value == nil { %s = runtime.NilValue{}; return }; val, ok := ret.value.(%s); if !ok { panic(fmt.Errorf(\"compiler: return type mismatch\")) }; %s = val", resultName, info.ReturnType, resultName)
		}
		fmt.Fprintf(buf, "\tdefer func() {\n")
		fmt.Fprintf(buf, "\t\tif recovered := recover(); recovered != nil {\n")
		fmt.Fprintf(buf, "\t\t\tif ret, ok := recovered.(__able_return); ok {\n")
		fmt.Fprintf(buf, "\t\t\t\t%s\n", recoverValue)
		fmt.Fprintf(buf, "\t\t\t\treturn\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tpanic(recovered)\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}()\n")
		for _, line := range lines {
			fmt.Fprintf(buf, "\t%s\n", line)
		}
		fmt.Fprintf(buf, "\treturn %s\n", retExpr)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderCompiledMethods(buf *bytes.Buffer) {
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil || !method.Info.Compileable {
			continue
		}
		info := method.Info
		ctx := newCompileContext(info, g.functions, g.overloads)
		lines, retExpr, ok := g.compileBody(ctx, info)
		if !ok {
			continue
		}
		fmt.Fprintf(buf, "func __able_compiled_%s(", info.GoName)
		for i, param := range info.Params {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "%s %s", param.GoName, param.GoType)
		}
		resultName := "__able_result"
		fmt.Fprintf(buf, ") (%s %s) {\n", resultName, info.ReturnType)
		recoverValue := fmt.Sprintf("val, ok := ret.value.(%s); if !ok { panic(fmt.Errorf(\"compiler: return type mismatch\")) }; %s = val", info.ReturnType, resultName)
		if info.ReturnType == "runtime.Value" {
			recoverValue = fmt.Sprintf("if ret.value == nil { %s = runtime.NilValue{}; return }; val, ok := ret.value.(%s); if !ok { panic(fmt.Errorf(\"compiler: return type mismatch\")) }; %s = val", resultName, info.ReturnType, resultName)
		}
		fmt.Fprintf(buf, "\tdefer func() {\n")
		fmt.Fprintf(buf, "\t\tif recovered := recover(); recovered != nil {\n")
		fmt.Fprintf(buf, "\t\t\tif ret, ok := recovered.(__able_return); ok {\n")
		fmt.Fprintf(buf, "\t\t\t\t%s\n", recoverValue)
		fmt.Fprintf(buf, "\t\t\t\treturn\n")
		fmt.Fprintf(buf, "\t\t\t}\n")
		fmt.Fprintf(buf, "\t\t\tpanic(recovered)\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}()\n")
		for _, line := range lines {
			fmt.Fprintf(buf, "\t%s\n", line)
		}
		fmt.Fprintf(buf, "\treturn %s\n", retExpr)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderWrappers(buf *bytes.Buffer) {
	for _, info := range g.sortedFunctionInfos() {
		if info == nil {
			continue
		}
		fmt.Fprintf(buf, "func __able_wrap_%s(rt *bridge.Runtime, ctx *runtime.NativeCallContext, args []runtime.Value) (result runtime.Value, err error) {\n", info.GoName)
		fmt.Fprintf(buf, "\tdefer func() {\n")
		fmt.Fprintf(buf, "\t\tif recovered := recover(); recovered != nil {\n")
		fmt.Fprintf(buf, "\t\t\tresult = nil\n")
		fmt.Fprintf(buf, "\t\t\terr = bridge.Recover(rt, ctx, recovered)\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}()\n")
		fmt.Fprintf(buf, "\tif rt != nil && ctx != nil && ctx.Env != nil {\n")
		fmt.Fprintf(buf, "\t\tprevEnv := rt.SwapEnv(ctx.Env)\n")
		fmt.Fprintf(buf, "\t\tdefer rt.SwapEnv(prevEnv)\n")
		fmt.Fprintf(buf, "\t}\n")
		if g.hasOptionalLastParam(info) && info.Arity > 0 {
			fmt.Fprintf(buf, "\tif len(args) == %d {\n", info.Arity-1)
			fmt.Fprintf(buf, "\t\targs = append(args, runtime.NilValue{})\n")
			fmt.Fprintf(buf, "\t}\n")
		}
		if info.Compileable {
			fmt.Fprintf(buf, "\tif len(args) != %d {\n", info.Arity)
			fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"arity mismatch calling %s: expected %d, got %%d\", len(args))\n", info.Name, info.Arity)
			fmt.Fprintf(buf, "\t}\n")
			for idx, param := range info.Params {
				argName := fmt.Sprintf("arg%d", idx)
				fmt.Fprintf(buf, "\t%sValue := args[%d]\n", argName, idx)
				g.renderArgConversion(buf, argName, param, info.Name)
			}
			fmt.Fprintf(buf, "\tcompiledResult := __able_compiled_%s(", info.GoName)
			for i, param := range info.Params {
				if i > 0 {
					fmt.Fprintf(buf, ", ")
				}
				fmt.Fprintf(buf, "%s", param.GoName)
			}
			fmt.Fprintf(buf, ")\n")
			g.renderReturnConversion(buf, "compiledResult", info.ReturnType, info.Definition.ReturnType, info.Name)
			fmt.Fprintf(buf, "}\n\n")
			continue
		}
		fmt.Fprintf(buf, "\treturn rt.CallOriginal(%q, args)\n", info.Name)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderMethodWrappers(buf *bytes.Buffer) {
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		info := method.Info
		fmt.Fprintf(buf, "func __able_wrap_%s(rt *bridge.Runtime, ctx *runtime.NativeCallContext, args []runtime.Value) (result runtime.Value, err error) {\n", info.GoName)
		fmt.Fprintf(buf, "\tdefer func() {\n")
		fmt.Fprintf(buf, "\t\tif recovered := recover(); recovered != nil {\n")
		fmt.Fprintf(buf, "\t\t\tresult = nil\n")
		fmt.Fprintf(buf, "\t\t\terr = bridge.Recover(rt, ctx, recovered)\n")
		fmt.Fprintf(buf, "\t\t}\n")
		fmt.Fprintf(buf, "\t}()\n")
		fmt.Fprintf(buf, "\tif rt != nil && ctx != nil && ctx.Env != nil {\n")
		fmt.Fprintf(buf, "\t\tprevEnv := rt.SwapEnv(ctx.Env)\n")
		fmt.Fprintf(buf, "\t\tdefer rt.SwapEnv(prevEnv)\n")
		fmt.Fprintf(buf, "\t}\n")
		if g.hasOptionalLastParam(info) && info.Arity > 0 {
			fmt.Fprintf(buf, "\tif len(args) == %d {\n", info.Arity-1)
			fmt.Fprintf(buf, "\t\targs = append(args, runtime.NilValue{})\n")
			fmt.Fprintf(buf, "\t}\n")
		}
		fmt.Fprintf(buf, "\tif len(args) != %d {\n", info.Arity)
		fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"arity mismatch calling %s: expected %d, got %%d\", len(args))\n", info.Name, info.Arity)
		fmt.Fprintf(buf, "\t}\n")
		for idx, param := range info.Params {
			argName := fmt.Sprintf("arg%d", idx)
			fmt.Fprintf(buf, "\t%sValue := args[%d]\n", argName, idx)
			g.renderArgConversion(buf, argName, param, info.Name)
		}
		fmt.Fprintf(buf, "\tcompiledResult := __able_compiled_%s(", info.GoName)
		for i, param := range info.Params {
			if i > 0 {
				fmt.Fprintf(buf, ", ")
			}
			fmt.Fprintf(buf, "%s", param.GoName)
		}
		fmt.Fprintf(buf, ")\n")
		g.renderReturnConversion(buf, "compiledResult", info.ReturnType, info.Definition.ReturnType, info.Name)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderMethodThunks(buf *bytes.Buffer) {
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		info := method.Info
		fmt.Fprintf(buf, "func __able_method_thunk_%s(env *runtime.Environment, args []runtime.Value) (runtime.Value, error) {\n", info.GoName)
		fmt.Fprintf(buf, "\tctx := &runtime.NativeCallContext{Env: env}\n")
		fmt.Fprintf(buf, "\treturn __able_wrap_%s(__able_runtime, ctx, args)\n", info.GoName)
		fmt.Fprintf(buf, "}\n\n")
	}
}

func (g *generator) renderRegister(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "func Register(interp *interpreter.Interpreter) (*bridge.Runtime, error) {\n")
	fmt.Fprintf(buf, "\treturn RegisterIn(interp, nil)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func RegisterIn(interp *interpreter.Interpreter, env *runtime.Environment) (*bridge.Runtime, error) {\n")
	fmt.Fprintf(buf, "\tif interp == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing interpreter\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif env == nil {\n")
	fmt.Fprintf(buf, "\t\tenv = interp.GlobalEnvironment()\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif env == nil {\n")
	fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"missing environment\")\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\trt := bridge.New(interp)\n")
	fmt.Fprintf(buf, "\t__able_runtime = rt\n")
	fmt.Fprintf(buf, "\trt.SetEnv(env)\n")
	if len(g.diagNodes) > 0 {
		fmt.Fprintf(buf, "\t__able_register_diag_nodes()\n")
	}
	for _, method := range g.sortedMethodInfos() {
		if method == nil || method.Info == nil {
			continue
		}
		if !g.registerableMethod(method) {
			continue
		}
		targetExpr, ok := g.renderTypeExpression(method.TargetType)
		if !ok {
			return
		}
		paramExprs, ok := g.renderMethodParamTypes(method)
		if !ok {
			return
		}
		fmt.Fprintf(buf, "\tif err := interp.RegisterCompiledMethodOverload(%q, %q, %t, %s, %s, __able_method_thunk_%s); err != nil {\n", method.TargetName, method.MethodName, method.ExpectsSelf, targetExpr, paramExprs, method.Info.GoName)
		fmt.Fprintf(buf, "\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t}\n")
	}
	for _, name := range g.sortedCallableNames() {
		if overload, ok := g.overloads[name]; ok && overload != nil {
			fmt.Fprintf(buf, "\tif original, err := env.Get(%q); err == nil {\n", name)
			fmt.Fprintf(buf, "\t\trt.RegisterOriginal(%q, original)\n", name)
			fmt.Fprintf(buf, "\t}\n")
			fmt.Fprintf(buf, "\t{\n")
			fmt.Fprintf(buf, "\t\toverloadFn := &runtime.NativeFunctionValue{Name: %q, Arity: -1}\n", name)
			fmt.Fprintf(buf, "\t\toverloadFn.Impl = func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
			fmt.Fprintf(buf, "\t\t\treturn %s(overloadFn, ctx, args, nil)\n", g.overloadWrapperName(name))
			fmt.Fprintf(buf, "\t\t}\n")
			fmt.Fprintf(buf, "\t\t%s = overloadFn\n", g.overloadValueName(name))
			fmt.Fprintf(buf, "\t\tenv.Define(%q, overloadFn)\n", name)
			fmt.Fprintf(buf, "\t}\n")
			continue
		}
		info := g.functions[name]
		if info == nil {
			continue
		}
		fmt.Fprintf(buf, "\tif original, err := env.Get(%q); err == nil {\n", info.Name)
		fmt.Fprintf(buf, "\t\trt.RegisterOriginal(%q, original)\n", info.Name)
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tenv.Define(%q, runtime.NativeFunctionValue{\n", info.Name)
		fmt.Fprintf(buf, "\t\tName: %q,\n", info.Name)
		fmt.Fprintf(buf, "\t\tArity: %d,\n", info.Arity)
		fmt.Fprintf(buf, "\t\tImpl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
		fmt.Fprintf(buf, "\t\t\treturn __able_wrap_%s(rt, ctx, args)\n", info.GoName)
		fmt.Fprintf(buf, "\t\t},\n")
		fmt.Fprintf(buf, "\t})\n")
	}
	fmt.Fprintf(buf, "\treturn rt, nil\n")
	fmt.Fprintf(buf, "}\n")
}

func (g *generator) renderMain() ([]byte, error) {
	if g.opts.PackageName != "main" {
		return nil, fmt.Errorf("compiler: EmitMain requires package name 'main'")
	}
	if g.opts.EntryPath == "" {
		return nil, fmt.Errorf("compiler: EmitMain requires EntryPath")
	}
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package main\n\n")
	fmt.Fprintf(&buf, "import (\n")
	fmt.Fprintf(&buf, "\t%q\n", "fmt")
	fmt.Fprintf(&buf, "\t%q\n", "os")
	fmt.Fprintf(&buf, "\t%q\n", "able/interpreter-go/pkg/driver")
	fmt.Fprintf(&buf, "\t%q\n", "able/interpreter-go/pkg/interpreter")
	fmt.Fprintf(&buf, ")\n\n")
	fmt.Fprintf(&buf, "func main() {\n")
	fmt.Fprintf(&buf, "\tentry := %q\n", g.opts.EntryPath)
	fmt.Fprintf(&buf, "\tloader, err := driver.NewLoader(nil)\n")
	fmt.Fprintf(&buf, "\tif err != nil {\n")
	fmt.Fprintf(&buf, "\t\tfmt.Fprintln(os.Stderr, err)\n")
	fmt.Fprintf(&buf, "\t\tos.Exit(1)\n")
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "\tdefer loader.Close()\n")
	fmt.Fprintf(&buf, "\tprogram, err := loader.Load(entry)\n")
	fmt.Fprintf(&buf, "\tif err != nil {\n")
	fmt.Fprintf(&buf, "\t\tfmt.Fprintln(os.Stderr, err)\n")
	fmt.Fprintf(&buf, "\t\tos.Exit(1)\n")
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "\tinterp := interpreter.New()\n")
	fmt.Fprintf(&buf, "\tinterp.SetArgs(os.Args[1:])\n")
	fmt.Fprintf(&buf, "\t_, entryEnv, _, err := interp.EvaluateProgram(program, interpreter.ProgramEvaluationOptions{})\n")
	fmt.Fprintf(&buf, "\tif err != nil {\n")
	fmt.Fprintf(&buf, "\t\tif code, ok := interpreter.ExitCodeFromError(err); ok {\n")
	fmt.Fprintf(&buf, "\t\t\tos.Exit(code)\n")
	fmt.Fprintf(&buf, "\t\t}\n")
	fmt.Fprintf(&buf, "\t\tfmt.Fprintln(os.Stderr, err)\n")
	fmt.Fprintf(&buf, "\t\tos.Exit(1)\n")
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "\tif _, err := RegisterIn(interp, entryEnv); err != nil {\n")
	fmt.Fprintf(&buf, "\t\tfmt.Fprintln(os.Stderr, err)\n")
	fmt.Fprintf(&buf, "\t\tos.Exit(1)\n")
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "\tif entryEnv == nil {\n")
	fmt.Fprintf(&buf, "\t\tentryEnv = interp.GlobalEnvironment()\n")
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "\tmainValue, err := entryEnv.Get(\"main\")\n")
	fmt.Fprintf(&buf, "\tif err != nil {\n")
	fmt.Fprintf(&buf, "\t\tfmt.Fprintln(os.Stderr, err)\n")
	fmt.Fprintf(&buf, "\t\tos.Exit(1)\n")
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "\tif _, err := interp.CallFunction(mainValue, nil); err != nil {\n")
	fmt.Fprintf(&buf, "\t\tif code, ok := interpreter.ExitCodeFromError(err); ok {\n")
	fmt.Fprintf(&buf, "\t\t\tos.Exit(code)\n")
	fmt.Fprintf(&buf, "\t\t}\n")
	fmt.Fprintf(&buf, "\t\tfmt.Fprintln(os.Stderr, err)\n")
	fmt.Fprintf(&buf, "\t\tos.Exit(1)\n")
	fmt.Fprintf(&buf, "\t}\n")
	fmt.Fprintf(&buf, "}\n")
	return formatSource(buf.Bytes())
}

func (g *generator) renderArgConversion(buf *bytes.Buffer, argName string, param paramInfo, funcName string) {
	goType := param.GoType
	target := param.GoName
	if g.typeCategory(goType) == "runtime" {
		if ifaceType, ok := g.interfaceTypeExpr(param.TypeExpr); ok {
			rendered, ok := g.renderTypeExpression(ifaceType)
			if ok {
				okName := fmt.Sprintf("%sOk", argName)
				expected := typeExpressionToString(ifaceType)
				fmt.Fprintf(buf, "\t%s, %s, err := bridge.MatchType(rt, %s, %sValue)\n", target, okName, rendered, argName)
				fmt.Fprintf(buf, "\tif err != nil {\n")
				fmt.Fprintf(buf, "\t\treturn nil, err\n")
				fmt.Fprintf(buf, "\t}\n")
				fmt.Fprintf(buf, "\tif !%s {\n", okName)
				fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"type mismatch calling %s: expected %s\")\n", funcName, expected)
				fmt.Fprintf(buf, "\t}\n")
				return
			}
		}
	}
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\t%s := %sValue\n", target, argName)
	case "bool":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsBool(%sValue)\n", target, argName)
		g.renderConvertErr(buf)
	case "string":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsString(%sValue)\n", target, argName)
		g.renderConvertErr(buf)
	case "rune":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsRune(%sValue)\n", target, argName)
		g.renderConvertErr(buf)
	case "float32":
		fmt.Fprintf(buf, "\t%sRaw, err := bridge.AsFloat(%sValue)\n", argName, argName)
		g.renderConvertErr(buf)
		fmt.Fprintf(buf, "\t%s := float32(%sRaw)\n", target, argName)
	case "float64":
		fmt.Fprintf(buf, "\t%s, err := bridge.AsFloat(%sValue)\n", target, argName)
		g.renderConvertErr(buf)
	case "int":
		fmt.Fprintf(buf, "\t%sRaw, err := bridge.AsInt(%sValue, bridge.NativeIntBits)\n", argName, argName)
		g.renderConvertErr(buf)
		fmt.Fprintf(buf, "\t%s := int(%sRaw)\n", target, argName)
	case "uint":
		fmt.Fprintf(buf, "\t%sRaw, err := bridge.AsUint(%sValue, bridge.NativeIntBits)\n", argName, argName)
		g.renderConvertErr(buf)
		fmt.Fprintf(buf, "\t%s := uint(%sRaw)\n", target, argName)
	case "int8", "int16", "int32", "int64":
		bits := g.intBits(goType)
		fmt.Fprintf(buf, "\t%sRaw, err := bridge.AsInt(%sValue, %d)\n", argName, argName, bits)
		g.renderConvertErr(buf)
		fmt.Fprintf(buf, "\t%s := %s(%sRaw)\n", target, goType, argName)
	case "uint8", "uint16", "uint32", "uint64":
		bits := g.intBits(goType)
		fmt.Fprintf(buf, "\t%sRaw, err := bridge.AsUint(%sValue, %d)\n", argName, argName, bits)
		g.renderConvertErr(buf)
		fmt.Fprintf(buf, "\t%s := %s(%sRaw)\n", target, goType, argName)
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "\t%s, err := __able_struct_%s_from(%sValue)\n", target, baseName, argName)
		g.renderConvertErr(buf)
	default:
		fmt.Fprintf(buf, "\t%s := %sValue\n", target, argName)
	}
}

func (g *generator) renderReturnConversion(buf *bytes.Buffer, resultName, goType string, returnType ast.TypeExpression, funcName string) {
	if g.typeCategory(goType) == "runtime" {
		if ifaceType, ok := g.interfaceTypeExpr(returnType); ok {
			rendered, ok := g.renderTypeExpression(ifaceType)
			if ok {
				okName := fmt.Sprintf("%sOk", resultName)
				expected := typeExpressionToString(ifaceType)
				fmt.Fprintf(buf, "\t%s, %s, err := bridge.MatchType(rt, %s, %s)\n", resultName, okName, rendered, resultName)
				fmt.Fprintf(buf, "\tif err != nil {\n")
				fmt.Fprintf(buf, "\t\treturn nil, err\n")
				fmt.Fprintf(buf, "\t}\n")
				fmt.Fprintf(buf, "\tif !%s {\n", okName)
				fmt.Fprintf(buf, "\t\treturn nil, fmt.Errorf(\"return type mismatch in %s: expected %s\")\n", funcName, expected)
				fmt.Fprintf(buf, "\t}\n")
				fmt.Fprintf(buf, "\treturn %s, nil\n", resultName)
				return
			}
		}
	}
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\treturn %s, nil\n", resultName)
	case "void":
		fmt.Fprintf(buf, "\t_ = %s\n", resultName)
		fmt.Fprintf(buf, "\treturn runtime.VoidValue{}, nil\n")
	case "bool":
		fmt.Fprintf(buf, "\treturn bridge.ToBool(%s), nil\n", resultName)
	case "string":
		fmt.Fprintf(buf, "\treturn bridge.ToString(%s), nil\n", resultName)
	case "rune":
		fmt.Fprintf(buf, "\treturn bridge.ToRune(%s), nil\n", resultName)
	case "float32":
		fmt.Fprintf(buf, "\treturn bridge.ToFloat32(%s), nil\n", resultName)
	case "float64":
		fmt.Fprintf(buf, "\treturn bridge.ToFloat64(%s), nil\n", resultName)
	case "int":
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\")), nil\n", resultName)
	case "uint":
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\")), nil\n", resultName)
	case "int8":
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\")), nil\n", resultName)
	case "int16":
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\")), nil\n", resultName)
	case "int32":
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\")), nil\n", resultName)
	case "int64":
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\")), nil\n", resultName)
	case "uint8":
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\")), nil\n", resultName)
	case "uint16":
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\")), nil\n", resultName)
	case "uint32":
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\")), nil\n", resultName)
	case "uint64":
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\")), nil\n", resultName)
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "\treturn __able_struct_%s_to(rt, %s)\n", baseName, resultName)
	default:
		fmt.Fprintf(buf, "\treturn %s, nil\n", resultName)
	}
}

func (g *generator) renderValueConversion(buf *bytes.Buffer, indent, valueVar, goType, assignTarget, returnExpr string) {
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "%s%s = %s\n", indent, assignTarget, valueVar)
	case "bool":
		fmt.Fprintf(buf, "%sconverted, err := bridge.AsBool(%s)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "string":
		fmt.Fprintf(buf, "%sconverted, err := bridge.AsString(%s)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "rune":
		fmt.Fprintf(buf, "%sconverted, err := bridge.AsRune(%s)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "float32":
		fmt.Fprintf(buf, "%sconvertedRaw, err := bridge.AsFloat(%s)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = float32(convertedRaw)\n", indent, assignTarget)
	case "float64":
		fmt.Fprintf(buf, "%sconverted, err := bridge.AsFloat(%s)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	case "int":
		fmt.Fprintf(buf, "%sconvertedRaw, err := bridge.AsInt(%s, bridge.NativeIntBits)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = int(convertedRaw)\n", indent, assignTarget)
	case "uint":
		fmt.Fprintf(buf, "%sconvertedRaw, err := bridge.AsUint(%s, bridge.NativeIntBits)\n", indent, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = uint(convertedRaw)\n", indent, assignTarget)
	case "int8", "int16", "int32", "int64":
		bits := g.intBits(goType)
		fmt.Fprintf(buf, "%sconvertedRaw, err := bridge.AsInt(%s, %d)\n", indent, valueVar, bits)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = %s(convertedRaw)\n", indent, assignTarget, goType)
	case "uint8", "uint16", "uint32", "uint64":
		bits := g.intBits(goType)
		fmt.Fprintf(buf, "%sconvertedRaw, err := bridge.AsUint(%s, %d)\n", indent, valueVar, bits)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = %s(convertedRaw)\n", indent, assignTarget, goType)
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "%sconverted, err := __able_struct_%s_from(%s)\n", indent, baseName, valueVar)
		g.renderConvertErrWith(buf, indent, returnExpr)
		fmt.Fprintf(buf, "%s%s = converted\n", indent, assignTarget)
	default:
		fmt.Fprintf(buf, "%sreturn %s, fmt.Errorf(\"unsupported field type\")\n", indent, returnExpr)
	}
}

func (g *generator) renderValueToRuntime(buf *bytes.Buffer, valueExpr, goType, targetSlice string) {
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\t%s = append(%s, %s)\n", targetSlice, targetSlice, valueExpr)
	case "bool":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToBool(%s))\n", targetSlice, targetSlice, valueExpr)
	case "string":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToString(%s))\n", targetSlice, targetSlice, valueExpr)
	case "rune":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToRune(%s))\n", targetSlice, targetSlice, valueExpr)
	case "float32":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToFloat32(%s))\n", targetSlice, targetSlice, valueExpr)
	case "float64":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToFloat64(%s))\n", targetSlice, targetSlice, valueExpr)
	case "int":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\")))\n", targetSlice, targetSlice, valueExpr)
	case "uint":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\")))\n", targetSlice, targetSlice, valueExpr)
	case "int8":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\")))\n", targetSlice, targetSlice, valueExpr)
	case "int16":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\")))\n", targetSlice, targetSlice, valueExpr)
	case "int32":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\")))\n", targetSlice, targetSlice, valueExpr)
	case "int64":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\")))\n", targetSlice, targetSlice, valueExpr)
	case "uint8":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\")))\n", targetSlice, targetSlice, valueExpr)
	case "uint16":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\")))\n", targetSlice, targetSlice, valueExpr)
	case "uint32":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\")))\n", targetSlice, targetSlice, valueExpr)
	case "uint64":
		fmt.Fprintf(buf, "\t%s = append(%s, bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\")))\n", targetSlice, targetSlice, valueExpr)
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "\tvalueField, err := __able_struct_%s_to(rt, %s)\n", baseName, valueExpr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\t%s = append(%s, valueField)\n", targetSlice, targetSlice)
	}
}

func (g *generator) renderValueToRuntimeNamed(buf *bytes.Buffer, valueExpr, goType, fieldName string) {
	switch g.typeCategory(goType) {
	case "runtime":
		fmt.Fprintf(buf, "\tfields[%q] = %s\n", fieldName, valueExpr)
	case "bool":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToBool(%s)\n", fieldName, valueExpr)
	case "string":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToString(%s)\n", fieldName, valueExpr)
	case "rune":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToRune(%s)\n", fieldName, valueExpr)
	case "float32":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToFloat32(%s)\n", fieldName, valueExpr)
	case "float64":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToFloat64(%s)\n", fieldName, valueExpr)
	case "int":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToInt(int64(%s), runtime.IntegerType(\"isize\"))\n", fieldName, valueExpr)
	case "uint":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToUint(uint64(%s), runtime.IntegerType(\"usize\"))\n", fieldName, valueExpr)
	case "int8":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToInt(int64(%s), runtime.IntegerType(\"i8\"))\n", fieldName, valueExpr)
	case "int16":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToInt(int64(%s), runtime.IntegerType(\"i16\"))\n", fieldName, valueExpr)
	case "int32":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToInt(int64(%s), runtime.IntegerType(\"i32\"))\n", fieldName, valueExpr)
	case "int64":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToInt(int64(%s), runtime.IntegerType(\"i64\"))\n", fieldName, valueExpr)
	case "uint8":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u8\"))\n", fieldName, valueExpr)
	case "uint16":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u16\"))\n", fieldName, valueExpr)
	case "uint32":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u32\"))\n", fieldName, valueExpr)
	case "uint64":
		fmt.Fprintf(buf, "\tfields[%q] = bridge.ToUint(uint64(%s), runtime.IntegerType(\"u64\"))\n", fieldName, valueExpr)
	case "struct":
		baseName, ok := g.structBaseName(goType)
		if !ok {
			baseName = strings.TrimPrefix(goType, "*")
		}
		fmt.Fprintf(buf, "\tvalueField, err := __able_struct_%s_to(rt, %s)\n", baseName, valueExpr)
		fmt.Fprintf(buf, "\tif err != nil {\n")
		fmt.Fprintf(buf, "\t\treturn nil, err\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tfields[%q] = valueField\n", fieldName)
	}
}

func (g *generator) renderConvertErr(buf *bytes.Buffer) {
	g.renderConvertErrWith(buf, "\t", "nil")
}

func (g *generator) renderConvertErrWith(buf *bytes.Buffer, indent string, returnExpr string) {
	fmt.Fprintf(buf, "%sif err != nil {\n", indent)
	fmt.Fprintf(buf, "%s\treturn %s, err\n", indent, returnExpr)
	fmt.Fprintf(buf, "%s}\n", indent)
}

func (g *generator) typeCategory(goType string) string {
	switch goType {
	case "runtime.Value":
		return "runtime"
	case "struct{}":
		return "void"
	case "bool":
		return "bool"
	case "string":
		return "string"
	case "rune":
		return "rune"
	case "float32":
		return "float32"
	case "float64":
		return "float64"
	case "int":
		return "int"
	case "uint":
		return "uint"
	case "int8", "int16", "int32", "int64":
		return "int" + goType[3:]
	case "uint8", "uint16", "uint32", "uint64":
		return "uint" + goType[4:]
	}
	for _, info := range g.structs {
		if info.GoName == goType {
			return "struct"
		}
		if strings.HasPrefix(goType, "*") && info.GoName == strings.TrimPrefix(goType, "*") {
			return "struct"
		}
	}
	return "unknown"
}

func (g *generator) sortedStructNames() []string {
	names := make([]string, 0, len(g.structs))
	for name := range g.structs {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func (g *generator) sortedFunctionNames() []string {
	names := make([]string, 0, len(g.functions))
	for name := range g.functions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func formatSource(src []byte) ([]byte, error) {
	formatted, err := format.Source(src)
	if err != nil {
		return src, err
	}
	return formatted, nil
}

func (g *generator) renderIteratorHelpers(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "type __able_generator_stop struct{}\n\n")
	fmt.Fprintf(buf, "func (__able_generator_stop) Error() string { return \"generator stopped\" }\n\n")
	fmt.Fprintf(buf, "func __able_is_generator_stop(err error) bool {\n")
	fmt.Fprintf(buf, "\tif err == nil {\n")
	fmt.Fprintf(buf, "\t\treturn false\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tvar stop __able_generator_stop\n")
	fmt.Fprintf(buf, "\treturn errors.As(err, &stop)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "type __able_generator_result struct {\n")
	fmt.Fprintf(buf, "\tvalue runtime.Value\n")
	fmt.Fprintf(buf, "\tdone bool\n")
	fmt.Fprintf(buf, "\terr error\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "type __able_generator struct {\n")
	fmt.Fprintf(buf, "\trequests chan struct{}\n")
	fmt.Fprintf(buf, "\tresults chan __able_generator_result\n")
	fmt.Fprintf(buf, "\tmu sync.Mutex\n")
	fmt.Fprintf(buf, "\tclosed bool\n")
	fmt.Fprintf(buf, "\tdone bool\n")
	fmt.Fprintf(buf, "\tcontrol runtime.Value\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_new_iterator(run func(gen runtime.Value) error) *runtime.IteratorValue {\n")
	fmt.Fprintf(buf, "\tgen := &__able_generator{requests: make(chan struct{}), results: make(chan __able_generator_result)}\n")
	fmt.Fprintf(buf, "\tgo gen.run(run)\n")
	fmt.Fprintf(buf, "\treturn runtime.NewIteratorValue(gen.next, gen.close)\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (g *__able_generator) run(run func(gen runtime.Value) error) {\n")
	fmt.Fprintf(buf, "\tdefer close(g.results)\n")
	fmt.Fprintf(buf, "\tif !g.awaitRequest() {\n\t\treturn\n\t}\n")
	fmt.Fprintf(buf, "\tvar runErr error\n")
	fmt.Fprintf(buf, "\tfunc() {\n")
	fmt.Fprintf(buf, "\t\tdefer func() {\n")
	fmt.Fprintf(buf, "\t\t\tif r := recover(); r != nil {\n")
	fmt.Fprintf(buf, "\t\t\t\tif err, ok := r.(error); ok {\n")
	fmt.Fprintf(buf, "\t\t\t\t\tif __able_is_generator_stop(err) {\n")
	fmt.Fprintf(buf, "\t\t\t\t\t\trunErr = __able_generator_stop{}\n")
	fmt.Fprintf(buf, "\t\t\t\t\t\treturn\n")
	fmt.Fprintf(buf, "\t\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t\t\trunErr = err\n")
	fmt.Fprintf(buf, "\t\t\t\t\treturn\n")
	fmt.Fprintf(buf, "\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t\trunErr = bridge.Recover(__able_runtime, nil, r)\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t}()\n")
	fmt.Fprintf(buf, "\t\trunErr = run(g.controllerValue())\n")
	fmt.Fprintf(buf, "\t}()\n")
	fmt.Fprintf(buf, "\tif runErr != nil {\n")
	fmt.Fprintf(buf, "\t\tif __able_is_generator_stop(runErr) {\n")
	fmt.Fprintf(buf, "\t\t\tg.results <- __able_generator_result{done: true}\n")
	fmt.Fprintf(buf, "\t\t\treturn\n")
	fmt.Fprintf(buf, "\t\t}\n")
	fmt.Fprintf(buf, "\t\tg.results <- __able_generator_result{err: runErr, done: true}\n")
	fmt.Fprintf(buf, "\t\treturn\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tg.results <- __able_generator_result{done: true}\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (g *__able_generator) next() (runtime.Value, bool, error) {\n")
	fmt.Fprintf(buf, "\tg.mu.Lock()\n")
	fmt.Fprintf(buf, "\tif g.closed || g.done {\n")
	fmt.Fprintf(buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(buf, "\t\treturn runtime.IteratorEnd, true, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tg.mu.Unlock()\n")
	fmt.Fprintf(buf, "\tg.requests <- struct{}{}\n")
	fmt.Fprintf(buf, "\tres, ok := <-g.results\n")
	fmt.Fprintf(buf, "\tg.mu.Lock()\n")
	fmt.Fprintf(buf, "\tif !ok {\n")
	fmt.Fprintf(buf, "\t\tg.done = true\n")
	fmt.Fprintf(buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(buf, "\t\treturn runtime.IteratorEnd, true, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif res.err != nil {\n")
	fmt.Fprintf(buf, "\t\tg.done = true\n")
	fmt.Fprintf(buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(buf, "\t\treturn nil, true, res.err\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tif res.done {\n")
	fmt.Fprintf(buf, "\t\tg.done = true\n")
	fmt.Fprintf(buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(buf, "\t\treturn runtime.IteratorEnd, true, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tg.mu.Unlock()\n")
	fmt.Fprintf(buf, "\tif res.value == nil {\n")
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}, false, nil\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn res.value, false, nil\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (g *__able_generator) close() {\n")
	fmt.Fprintf(buf, "\tg.mu.Lock()\n")
	fmt.Fprintf(buf, "\tif g.closed {\n")
	fmt.Fprintf(buf, "\t\tg.mu.Unlock()\n")
	fmt.Fprintf(buf, "\t\treturn\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tg.closed = true\n")
	fmt.Fprintf(buf, "\tclose(g.requests)\n")
	fmt.Fprintf(buf, "\tg.mu.Unlock()\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (g *__able_generator) awaitRequest() bool {\n")
	fmt.Fprintf(buf, "\t_, ok := <-g.requests\n")
	fmt.Fprintf(buf, "\treturn ok\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (g *__able_generator) emit(value runtime.Value) error {\n")
	fmt.Fprintf(buf, "\tg.results <- __able_generator_result{value: value}\n")
	fmt.Fprintf(buf, "\tif !g.awaitRequest() {\n")
	fmt.Fprintf(buf, "\t\treturn __able_generator_stop{}\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn nil\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func (g *__able_generator) controllerValue() runtime.Value {\n")
	fmt.Fprintf(buf, "\tg.mu.Lock()\n")
	fmt.Fprintf(buf, "\tdefer g.mu.Unlock()\n")
	fmt.Fprintf(buf, "\tif g.control != nil {\n")
	fmt.Fprintf(buf, "\t\treturn g.control\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tyieldFn := runtime.NativeFunctionValue{\n")
	fmt.Fprintf(buf, "\t\tName: \"__iterator_controller_yield\",\n")
	fmt.Fprintf(buf, "\t\tArity: 1,\n")
	fmt.Fprintf(buf, "\t\tImpl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
	fmt.Fprintf(buf, "\t\t\tvar value runtime.Value = runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t\t\tif len(args) > 0 {\n")
	fmt.Fprintf(buf, "\t\t\t\tif len(args) > 1 {\n")
	fmt.Fprintf(buf, "\t\t\t\t\treturn nil, fmt.Errorf(\"gen.yield expects at most one argument\")\n")
	fmt.Fprintf(buf, "\t\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\t\tvalue = args[0]\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\tif err := g.emit(value); err != nil {\n")
	fmt.Fprintf(buf, "\t\t\t\treturn nil, err\n")
	fmt.Fprintf(buf, "\t\t\t}\n")
	fmt.Fprintf(buf, "\t\t\treturn runtime.NilValue{}, nil\n")
	fmt.Fprintf(buf, "\t\t},\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tcloseFn := runtime.NativeFunctionValue{\n")
	fmt.Fprintf(buf, "\t\tName: \"__iterator_controller_close\",\n")
	fmt.Fprintf(buf, "\t\tArity: 0,\n")
	fmt.Fprintf(buf, "\t\tImpl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {\n")
	fmt.Fprintf(buf, "\t\t\tg.close()\n")
	fmt.Fprintf(buf, "\t\t\treturn runtime.NilValue{}, nil\n")
	fmt.Fprintf(buf, "\t\t},\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tstopFn := runtime.NativeFunctionValue{\n")
	fmt.Fprintf(buf, "\t\tName: \"__iterator_controller_stop\",\n")
	fmt.Fprintf(buf, "\t\tArity: 0,\n")
	fmt.Fprintf(buf, "\t\tImpl: func(_ *runtime.NativeCallContext, _ []runtime.Value) (runtime.Value, error) {\n")
	fmt.Fprintf(buf, "\t\t\tg.close()\n")
	fmt.Fprintf(buf, "\t\t\treturn runtime.NilValue{}, __able_generator_stop{}\n")
	fmt.Fprintf(buf, "\t\t},\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\tg.control = &runtime.StructInstanceValue{\n")
	fmt.Fprintf(buf, "\t\tFields: map[string]runtime.Value{\n")
	fmt.Fprintf(buf, "\t\t\t\"yield\": yieldFn,\n")
	fmt.Fprintf(buf, "\t\t\t\"close\": closeFn,\n")
	fmt.Fprintf(buf, "\t\t\t\"stop\":  stopFn,\n")
	fmt.Fprintf(buf, "\t\t},\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn g.control\n")
	fmt.Fprintf(buf, "}\n\n")
}
