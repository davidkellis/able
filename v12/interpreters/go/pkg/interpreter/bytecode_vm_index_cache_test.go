package interpreter

import (
	"math/big"
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_DirectArrayIndexFastPath(t *testing.T) {
	cases := []struct {
		name  string
		value runtime.Value
		want  int
	}{
		{name: "small_value", value: runtime.NewSmallInt(7, runtime.IntegerI32), want: 7},
		{name: "small_pointer", value: func() runtime.Value {
			v := runtime.NewSmallInt(11, runtime.IntegerI32)
			return &v
		}(), want: 11},
		{name: "boxed_big_value", value: runtime.NewBigIntValue(big.NewInt(19), runtime.IntegerI32), want: 19},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, handled, err := bytecodeDirectArrayIndex(tc.value)
			if err != nil {
				t.Fatalf("direct array index returned error: %v", err)
			}
			if !handled {
				t.Fatalf("expected direct array index to handle %T", tc.value)
			}
			if got != tc.want {
				t.Fatalf("unexpected direct array index: got=%d want=%d", got, tc.want)
			}
		})
	}
}

func TestBytecodeVM_DirectArrayIndexSetSyncsSharedAliases(t *testing.T) {
	interp := NewBytecode()
	vm := newBytecodeVM(interp, interp.global)
	first := interp.newArrayValue([]runtime.Value{
		runtime.NewSmallInt(1, runtime.IntegerI32),
	}, 1)
	_, err := interp.ensureArrayState(first, 0)
	if err != nil {
		t.Fatalf("ensure first array state: %v", err)
	}
	second, err := interp.arrayValueFromHandle(first.Handle, 0, 0)
	if err != nil {
		t.Fatalf("arrayValueFromHandle: %v", err)
	}
	if !first.TrackedAliases || !second.TrackedAliases {
		t.Fatalf("expected both aliases to be marked shared before direct set")
	}

	written := runtime.StringValue{Val: "x"}
	got, handled, err := vm.resolveDirectArrayIndexSet(first, runtime.NewSmallInt(0, runtime.IntegerI32), written, ast.AssignmentAssign, "", false)
	if err != nil {
		t.Fatalf("direct array index set returned error: %v", err)
	}
	if !handled {
		t.Fatalf("expected direct array index set to handle tracked array write")
	}
	if !valuesEqual(got, written) {
		t.Fatalf("unexpected direct array index set result: got=%#v want=%#v", got, written)
	}
	if observed, ok := second.Elements[0].(runtime.StringValue); !ok || observed.Val != "x" {
		t.Fatalf("expected shared alias to observe direct bytecode set, got %#v", second.Elements[0])
	}
}

func TestBytecodeVM_IndexMethodCacheTracksArrayElementType(t *testing.T) {
	indexIface := ast.Iface(
		"Index",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"get",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("idx", ast.Ty("i32")),
				},
				ast.Ty("i32"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)

	getI32 := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Int(11),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	getString := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("String"))),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Int(22),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		indexIface,
		ast.Impl("Index", ast.Gen(ast.Ty("Array"), ast.Ty("i32")), []*ast.FunctionDefinition{getI32}, nil, nil, nil, nil, false),
		ast.Impl("Index", ast.Gen(ast.Ty("Array"), ast.Ty("String")), []*ast.FunctionDefinition{getString}, nil, nil, nil, nil, false),
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(1), ast.Int(2))),
		ast.Assign(ast.ID("first"), ast.Index(ast.ID("arr"), ast.Int(1))),
		ast.AssignOp(ast.AssignmentAssign, ast.Index(ast.ID("arr"), ast.Int(0)), ast.Str("x")),
		ast.Assign(ast.ID("second"), ast.Index(ast.ID("arr"), ast.Int(1))),
		ast.ID("second"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode index cache array element-type dispatch mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 22 {
		t.Fatalf("expected second index lookup to use Array String impl and return 22, got %#v", got)
	}
}

func TestBytecodeVM_IndexSetCompoundCacheInvalidatesWhenImplAppears(t *testing.T) {
	indexIface := ast.Iface(
		"Index",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"get",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("idx", ast.Ty("i32")),
				},
				ast.Ty("i32"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)
	indexMutIface := ast.Iface(
		"IndexMut",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"set",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("idx", ast.Ty("i32")),
					ast.Param("value", ast.Ty("i32")),
				},
				ast.Ty("i32"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)

	bump := ast.Fn(
		"bump",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("delta", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAdd, ast.Index(ast.ID("arr"), ast.Int(0)), ast.ID("delta")),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	getI32 := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ID("marker"), ast.Bin("+", ast.ID("marker"), ast.Int(10))),
			ast.Int(7),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	setI32 := ast.Fn(
		"set",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("idx", ast.Ty("i32")),
			ast.Param("value", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.AssignOp(ast.AssignmentAssign, ast.ID("marker"), ast.Bin("+", ast.ID("marker"), ast.ID("value"))),
			ast.Int(0),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		indexIface,
		indexMutIface,
		bump,
		ast.Assign(ast.ID("marker"), ast.Int(0)),
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(1))),
		ast.Call("bump", ast.ID("arr"), ast.Int(2)),
		ast.Impl("Index", ast.Gen(ast.Ty("Array"), ast.Ty("i32")), []*ast.FunctionDefinition{getI32}, nil, nil, nil, nil, false),
		ast.Impl("IndexMut", ast.Gen(ast.Ty("Array"), ast.Ty("i32")), []*ast.FunctionDefinition{setI32}, nil, nil, nil, nil, false),
		ast.Call("bump", ast.ID("arr"), ast.Int(5)),
		ast.ID("marker"),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode compound index cache invalidation mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 22 {
		t.Fatalf("expected marker 22 after impl-backed compound assignment, got %#v", got)
	}
}

func TestBytecodeVM_IndexGetFastPathInvalidatesWhenImplAppears(t *testing.T) {
	indexIface := ast.Iface(
		"Index",
		[]*ast.FunctionSignature{
			ast.FnSig(
				"get",
				[]*ast.FunctionParameter{
					ast.Param("self", ast.Ty("Self")),
					ast.Param("idx", ast.Ty("i32")),
				},
				ast.Ty("i32"),
				nil,
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		nil,
		false,
	)

	read := ast.Fn(
		"read",
		[]*ast.FunctionParameter{
			ast.Param("arr", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
		},
		[]ast.Statement{
			ast.Index(ast.ID("arr"), ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	getI32 := ast.Fn(
		"get",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("idx", ast.Ty("i32")),
		},
		[]ast.Statement{
			ast.Int(99),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		indexIface,
		read,
		ast.Assign(ast.ID("arr"), ast.Arr(ast.Int(7))),
		ast.Assign(ast.ID("before"), ast.Call("read", ast.ID("arr"))),
		ast.Impl("Index", ast.Gen(ast.Ty("Array"), ast.Ty("i32")), []*ast.FunctionDefinition{getI32}, nil, nil, nil, nil, false),
		ast.Assign(ast.ID("after"), ast.Call("read", ast.ID("arr"))),
		ast.Bin("+", ast.ID("before"), ast.ID("after")),
	}, nil, nil)

	want := mustEvalModule(t, New(), module)
	got := runBytecodeModule(t, module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode index get fast-path invalidation mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 106 {
		t.Fatalf("expected before+after marker 106 after impl-backed read, got %#v", got)
	}
}
