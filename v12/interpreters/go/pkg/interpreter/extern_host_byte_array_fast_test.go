package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBuildExternFastInvokerForArrayU8ReturnKinds(t *testing.T) {
	toString := ast.Extern(
		ast.HostTargetGo,
		ast.Fn(
			"encode_like",
			[]*ast.FunctionParameter{ast.Param("value", ast.Gen(ast.Ty("Array"), ast.Ty("u8")))},
			nil,
			ast.Ty("String"),
			nil,
			nil,
			false,
			false,
		),
		`return ""`,
	)
	toBytes := ast.Extern(
		ast.HostTargetGo,
		ast.Fn(
			"decode_like",
			[]*ast.FunctionParameter{ast.Param("value", ast.Gen(ast.Ty("Array"), ast.Ty("u8")))},
			nil,
			ast.Gen(ast.Ty("Array"), ast.Ty("u8")),
			nil,
			nil,
			false,
			false,
		),
		`return []byte{}`,
	)

	stringInvoker := buildExternFastInvoker(toString, func(value []byte) string { return string(value) + "!" })
	if stringInvoker == nil {
		t.Fatalf("expected Array u8 -> String fast invoker")
	}
	gotString, err := stringInvoker(New(), []runtime.Value{runtime.ArrayStoreMonoValueFromU8Bytes([]byte("abc"))})
	if err != nil {
		t.Fatalf("invoke Array u8 -> String fast invoker: %v", err)
	}
	str, ok := gotString.(runtime.StringValue)
	if !ok || str.Val != "abc!" {
		t.Fatalf("unexpected string result %#v", gotString)
	}

	bytesInvoker := buildExternFastInvoker(toBytes, func(value []byte) []byte { return append(value[:0:0], value...) })
	if bytesInvoker == nil {
		t.Fatalf("expected Array u8 -> Array u8 fast invoker")
	}
	gotBytes, err := bytesInvoker(New(), []runtime.Value{runtime.ArrayStoreMonoValueFromU8Bytes([]byte{4, 5, 6})})
	if err != nil {
		t.Fatalf("invoke Array u8 -> Array u8 fast invoker: %v", err)
	}
	arr, ok := gotBytes.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array result, got %T", gotBytes)
	}
	for idx, want := range []uint8{4, 5, 6} {
		raw, ok, err := runtime.ArrayStoreMonoReadU8IfAvailable(arr.Handle, idx)
		if err != nil {
			t.Fatalf("read mono u8[%d]: %v", idx, err)
		}
		if !ok || raw != want {
			t.Fatalf("mono u8[%d] = %d/%v, want %d/true", idx, raw, ok, want)
		}
	}
}

func TestBuildExternFastInvokerForArrayU8BorrowOptIn(t *testing.T) {
	copied := ast.Extern(
		ast.HostTargetGo,
		ast.Fn(
			"encode_like",
			[]*ast.FunctionParameter{ast.Param("value", ast.Gen(ast.Ty("Array"), ast.Ty("u8")))},
			nil,
			ast.Ty("String"),
			nil,
			nil,
			false,
			false,
		),
		`return ""`,
	)
	borrowed := ast.Extern(
		ast.HostTargetGo,
		ast.Fn(
			"encode_like_borrowed",
			[]*ast.FunctionParameter{ast.Param("value", ast.Gen(ast.Ty("Array"), ast.Ty("u8")))},
			nil,
			ast.Ty("String"),
			nil,
			nil,
			false,
			false,
		),
		`value = able_borrowed_bytes(value)
return ""`,
	)

	source := []byte("abc")
	arr := runtime.ArrayStoreMonoValueFromOwnedU8Bytes(source)

	copiedInvoker := buildExternFastInvoker(copied, func(value []byte) string {
		if len(value) > 0 {
			value[0] = 'z'
		}
		return string(value)
	})
	if copiedInvoker == nil {
		t.Fatalf("expected copied Array u8 fast invoker")
	}
	if _, err := copiedInvoker(New(), []runtime.Value{arr}); err != nil {
		t.Fatalf("invoke copied Array u8 fast invoker: %v", err)
	}
	if source[0] != 'a' {
		t.Fatalf("non-opt-in Array u8 arg mutated source: %q", source)
	}

	borrowedInvoker := buildExternFastInvoker(borrowed, func(value []byte) string {
		if len(value) == 0 || &value[0] != &source[0] {
			t.Fatalf("expected borrowed Array u8 arg to reuse mono backing")
		}
		return string(value)
	})
	if borrowedInvoker == nil {
		t.Fatalf("expected borrowed Array u8 fast invoker")
	}
	got, err := borrowedInvoker(New(), []runtime.Value{arr})
	if err != nil {
		t.Fatalf("invoke borrowed Array u8 fast invoker: %v", err)
	}
	str, ok := got.(runtime.StringValue)
	if !ok || str.Val != "abc" {
		t.Fatalf("unexpected borrowed string result %#v", got)
	}
}

func TestBuildExternFastInvokerForArrayU8UnionSignature(t *testing.T) {
	interp := New()
	sig := ast.Fn(
		"decode_like",
		[]*ast.FunctionParameter{ast.Param("value", ast.Gen(ast.Ty("Array"), ast.Ty("u8")))},
		nil,
		&ast.UnionTypeExpression{
			Members: []ast.TypeExpression{
				ast.Ty("IOError"),
				ast.Gen(ast.Ty("Array"), ast.Ty("u8")),
			},
		},
		nil,
		nil,
		false,
		false,
	)
	def := ast.Extern(ast.HostTargetGo, sig, `return value`)
	invoker := buildExternFastInvoker(def, func(value []byte) interface{} { return value })
	if invoker == nil {
		t.Fatalf("expected fast invoker for Array u8 union signature")
	}
	got, err := invoker(interp, []runtime.Value{runtime.ArrayStoreMonoValueFromU8Bytes([]byte("abc"))})
	if err != nil {
		t.Fatalf("run Array u8 union invoker: %v", err)
	}
	arr, ok := got.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array result, got %T", got)
	}
	for idx, want := range []uint8{'a', 'b', 'c'} {
		raw, ok, err := runtime.ArrayStoreMonoReadU8IfAvailable(arr.Handle, idx)
		if err != nil {
			t.Fatalf("read union mono u8[%d]: %v", idx, err)
		}
		if !ok || raw != want {
			t.Fatalf("union mono u8[%d] = %d/%v, want %d/true", idx, raw, ok, want)
		}
	}
}

func TestBuildExternFastInvokerForBorrowedArrayU8UnionSignature(t *testing.T) {
	interp := New()
	sig := ast.Fn(
		"decode_like_borrowed",
		[]*ast.FunctionParameter{ast.Param("value", ast.Gen(ast.Ty("Array"), ast.Ty("u8")))},
		nil,
		&ast.UnionTypeExpression{
			Members: []ast.TypeExpression{
				ast.Ty("IOError"),
				ast.Gen(ast.Ty("Array"), ast.Ty("u8")),
			},
		},
		nil,
		nil,
		false,
		false,
	)
	def := ast.Extern(ast.HostTargetGo, sig, `value = able_borrowed_bytes(value)
return value`)
	invoker := buildExternFastInvoker(def, func(value []byte) interface{} { return value })
	if invoker == nil {
		t.Fatalf("expected fast invoker for borrowed Array u8 union signature")
	}
	got, err := invoker(interp, []runtime.Value{runtime.ArrayStoreMonoValueFromOwnedU8Bytes([]byte("abc"))})
	if err != nil {
		t.Fatalf("run borrowed Array u8 union invoker: %v", err)
	}
	arr, ok := got.(*runtime.ArrayValue)
	if !ok {
		t.Fatalf("expected array result, got %T", got)
	}
	for idx, want := range []uint8{'a', 'b', 'c'} {
		raw, ok, err := runtime.ArrayStoreMonoReadU8IfAvailable(arr.Handle, idx)
		if err != nil {
			t.Fatalf("read borrowed union mono u8[%d]: %v", idx, err)
		}
		if !ok || raw != want {
			t.Fatalf("borrowed union mono u8[%d] = %d/%v, want %d/true", idx, raw, ok, want)
		}
	}
}
