package interpreter

import (
	"reflect"
	"strings"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestExternReflectStringSliceResult(t *testing.T) {
	got, ok := externReflectStringSliceResult(reflect.ValueOf([]string{"a", "b"}))
	if !ok {
		t.Fatalf("expected fast string-slice conversion")
	}
	arr, ok := got.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array result, got %T", got)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected two elements, got %d", len(arr.Elements))
	}
	first, ok := arr.Elements[0].(runtime.StringValue)
	if !ok || first.Val != "a" {
		t.Fatalf("unexpected first element %#v", arr.Elements[0])
	}
	second, ok := arr.Elements[1].(runtime.StringValue)
	if !ok || second.Val != "b" {
		t.Fatalf("unexpected second element %#v", arr.Elements[1])
	}
}

func TestExternStringSliceCacheClonesCachedTemplate(t *testing.T) {
	var cache externStringSliceCache
	source := []string{"alpha", "beta"}

	first := cache.result(source)
	firstArr, ok := first.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected first array result, got %T", first)
	}
	firstArr.Elements[0] = runtime.StringValue{Val: "changed"}

	second := cache.result(source)
	secondArr, ok := second.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected second array result, got %T", second)
	}
	secondFirst, ok := secondArr.Elements[0].(runtime.StringValue)
	if !ok || secondFirst.Val != "alpha" {
		t.Fatalf("expected cached template clone to preserve alpha, got %#v", secondArr.Elements[0])
	}

	source[0] = "gamma"
	third := cache.result(source)
	thirdArr, ok := third.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected third array result, got %T", third)
	}
	thirdFirst, ok := thirdArr.Elements[0].(runtime.StringValue)
	if !ok || thirdFirst.Val != "gamma" {
		t.Fatalf("expected cache invalidation on source change, got %#v", thirdArr.Elements[0])
	}
}

func TestExternUnionPreferredMemberForStringSlice(t *testing.T) {
	union := &ast.UnionTypeExpression{
		Members: []ast.TypeExpression{
			ast.Ty("IOError"),
			ast.Gen(ast.Ty("Array"), ast.Ty("String")),
		},
	}

	member := externUnionPreferredMemberForHostValue(union, reflect.ValueOf([]string{"x"}))
	if member == nil {
		t.Fatalf("expected preferred union member")
	}
	if !externIsArrayStringType(member) {
		t.Fatalf("expected Array String member, got %T", member)
	}
}

func TestFromHostValueUnionArrayStringFastPath(t *testing.T) {
	interp := New()
	union := &ast.UnionTypeExpression{
		Members: []ast.TypeExpression{
			ast.Ty("IOError"),
			ast.Gen(ast.Ty("Array"), ast.Ty("String")),
		},
	}

	got, err := interp.fromHostValue(union, reflect.ValueOf([]string{"ceiling", "science"}))
	if err != nil {
		t.Fatalf("fromHostValue: %v", err)
	}
	arr, ok := got.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array result, got %T", got)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected two elements, got %d", len(arr.Elements))
	}
}

func TestExternModuleBuildsFastInvokerForHotI32Signature(t *testing.T) {
	interp := New()
	lenSig := ast.Fn(
		"len_like",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("String"))},
		nil,
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	lenExtern := ast.Extern(ast.HostTargetGo, lenSig, `return int32(len(value))`)
	mod := ast.Mod([]ast.Statement{lenExtern}, nil, ast.Pkg([]interface{}{"host"}, false))

	if _, _, err := interp.EvaluateModule(mod); err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	pkg := interp.externHostPackages["host"]
	if pkg == nil {
		t.Fatalf("expected extern host package")
	}
	state := pkg.targets[ast.HostTargetGo]
	if state == nil {
		t.Fatalf("expected go extern target state")
	}
	module, err := interp.ensureExternHostModule("host", ast.HostTargetGo, state, pkg)
	if err != nil {
		t.Fatalf("ensure extern host module: %v", err)
	}

	invoker, err := module.lookupInvoker(lenExtern)
	if err != nil {
		t.Fatalf("lookup len invoker: %v", err)
	}
	if invoker == nil {
		t.Fatalf("expected fast invoker for len_like")
	}
	got, err := invoker(interp, []runtime.Value{runtime.StringValue{Val: "abc"}})
	if err != nil {
		t.Fatalf("run len invoker: %v", err)
	}
	intVal, ok := got.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected integer result, got %T", got)
	}
	if !intVal.IsSmall() || intVal.Int64Fast() != 3 || intVal.TypeSuffix != runtime.IntegerI32 {
		t.Fatalf("unexpected integer result %#v", intVal)
	}
}

func TestExternModuleBuildsFastInvokerForUnionArrayStringSignature(t *testing.T) {
	interp := New()
	linesSig := ast.Fn(
		"lines_like",
		[]*ast.FunctionParameter{ast.Param("path", ast.Ty("String"))},
		nil,
		&ast.UnionTypeExpression{
			Members: []ast.TypeExpression{
				ast.Ty("IOError"),
				ast.Gen(ast.Ty("Array"), ast.Ty("String")),
			},
		},
		nil,
		nil,
		false,
		false,
	)
	linesExtern := ast.Extern(ast.HostTargetGo, linesSig, `return []string{path, path}`)
	mod := ast.Mod([]ast.Statement{linesExtern}, nil, ast.Pkg([]interface{}{"host"}, false))

	if _, _, err := interp.EvaluateModule(mod); err != nil {
		t.Fatalf("evaluate module: %v", err)
	}

	pkg := interp.externHostPackages["host"]
	if pkg == nil {
		t.Fatalf("expected extern host package")
	}
	state := pkg.targets[ast.HostTargetGo]
	if state == nil {
		t.Fatalf("expected go extern target state")
	}
	module, err := interp.ensureExternHostModule("host", ast.HostTargetGo, state, pkg)
	if err != nil {
		t.Fatalf("ensure extern host module: %v", err)
	}

	invoker, err := module.lookupInvoker(linesExtern)
	if err != nil {
		t.Fatalf("lookup union lines invoker: %v", err)
	}
	if invoker == nil {
		t.Fatalf("expected fast invoker for lines_like")
	}
	got, err := invoker(interp, []runtime.Value{runtime.StringValue{Val: "wordlist.txt"}})
	if err != nil {
		t.Fatalf("run union lines invoker: %v", err)
	}
	arr, ok := got.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array result, got %T", got)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("expected two elements, got %d", len(arr.Elements))
	}
	first, ok := arr.Elements[0].(runtime.StringValue)
	if !ok || first.Val != "wordlist.txt" {
		t.Fatalf("unexpected first element %#v", arr.Elements[0])
	}
}

func TestBuildExternFastInvoker_StringReplaceSkipsHostCallWhenNeedleMissing(t *testing.T) {
	def := ast.Extern(
		ast.HostTargetGo,
		ast.Fn(
			"string_replace_fast",
			[]*ast.FunctionParameter{
				ast.Param("haystack", ast.Ty("String")),
				ast.Param("old", ast.Ty("String")),
				ast.Param("replacement", ast.Ty("String")),
			},
			nil,
			ast.Ty("String"),
			nil,
			nil,
			false,
			false,
		),
		`return strings.ReplaceAll(haystack, old, replacement)`,
	)

	calls := 0
	invoker := buildExternFastInvoker(def, func(haystack string, old string, replacement string) string {
		calls++
		return strings.ReplaceAll(haystack, old, replacement)
	})
	if invoker == nil {
		t.Fatalf("expected fast invoker")
	}

	got, err := invoker(New(), []runtime.Value{
		runtime.StringValue{Val: "science"},
		runtime.StringValue{Val: "zzz"},
		runtime.StringValue{Val: ""},
	})
	if err != nil {
		t.Fatalf("invoke fast string_replace_fast: %v", err)
	}
	if calls != 0 {
		t.Fatalf("expected missing-needle fast path to skip host call, got %d calls", calls)
	}
	str, ok := got.(runtime.StringValue)
	if !ok || str.Val != "science" {
		t.Fatalf("unexpected fast replace result %#v", got)
	}
}

func TestBuildExternFastInvoker_OtherThreeStringExternStillCallsHost(t *testing.T) {
	def := ast.Extern(
		ast.HostTargetGo,
		ast.Fn(
			"other_transform",
			[]*ast.FunctionParameter{
				ast.Param("left", ast.Ty("String")),
				ast.Param("middle", ast.Ty("String")),
				ast.Param("right", ast.Ty("String")),
			},
			nil,
			ast.Ty("String"),
			nil,
			nil,
			false,
			false,
		),
		`return left + right`,
	)

	calls := 0
	invoker := buildExternFastInvoker(def, func(left string, middle string, right string) string {
		calls++
		return left + right
	})
	if invoker == nil {
		t.Fatalf("expected fast invoker")
	}

	got, err := invoker(New(), []runtime.Value{
		runtime.StringValue{Val: "science"},
		runtime.StringValue{Val: "zzz"},
		runtime.StringValue{Val: "!"},
	})
	if err != nil {
		t.Fatalf("invoke generic 3-string fast invoker: %v", err)
	}
	if calls != 1 {
		t.Fatalf("expected non-replace extern to call host implementation once, got %d", calls)
	}
	str, ok := got.(runtime.StringValue)
	if !ok || str.Val != "science!" {
		t.Fatalf("unexpected generic 3-string result %#v", got)
	}
}
