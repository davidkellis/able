package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestBytecodeVM_MemberMethodCacheTracksStructDefinition(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	structS := ast.StructDef(
		"S",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	structT := ast.StructDef(
		"T",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	sPing := ast.Fn(
		"ping",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
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
	tPing := ast.Fn(
		"ping",
		[]*ast.FunctionParameter{
			ast.Param("self", ast.Ty("Self")),
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

	callPing := ast.Fn(
		"call_ping",
		[]*ast.FunctionParameter{
			ast.Param("value", nil),
		},
		[]ast.Statement{
			ast.CallExpr(ast.Member(ast.ID("value"), "ping")),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		structS,
		structT,
		ast.Methods(ast.Ty("S"), []*ast.FunctionDefinition{sPing}, nil, nil),
		ast.Methods(ast.Ty("T"), []*ast.FunctionDefinition{tPing}, nil, nil),
		callPing,
		ast.Assign(
			ast.ID("s"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(1), "n"),
			}, false, "S", nil, nil),
		),
		ast.Assign(
			ast.ID("t"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(2), "n"),
			}, false, "T", nil, nil),
		),
		ast.Call("call_ping", ast.ID("s")),
		ast.Call("call_ping", ast.ID("t")),
		ast.Call("call_ping", ast.ID("t")),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode member cache receiver-definition mismatch: got=%#v want=%#v", got, want)
	}
	if intVal, ok := got.(runtime.IntegerValue); !ok || intVal.BigInt().Int64() != 22 {
		t.Fatalf("expected second call_ping to use T.ping and return 22, got %#v", got)
	}

	stats := interp.BytecodeStats()
	if stats.MemberMethodCacheMiss < 2 {
		t.Fatalf("expected receiver-definition changes to force cache misses, got misses=%d", stats.MemberMethodCacheMiss)
	}
	if stats.MemberMethodCacheHits == 0 {
		t.Fatalf("expected member method cache hit after repeating the T receiver")
	}
}

func TestBytecodeVM_NonMethodMemberAccessSkipsMemberMethodCacheCounters(t *testing.T) {
	t.Setenv("ABLE_BYTECODE_STATS", "1")

	structDef := ast.StructDef(
		"S",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)

	readField := ast.Fn(
		"read_field",
		[]*ast.FunctionParameter{
			ast.Param("s", ast.Ty("S")),
		},
		[]ast.Statement{
			ast.Member(ast.ID("s"), "n"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.Mod([]ast.Statement{
		structDef,
		readField,
		ast.Assign(
			ast.ID("s"),
			ast.StructLit([]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Int(7), "n"),
			}, false, "S", nil, nil),
		),
		ast.Call("read_field", ast.ID("s")),
		ast.Call("read_field", ast.ID("s")),
	}, nil, nil)

	interp := NewBytecode()
	got := runBytecodeModuleWithInterpreter(t, interp, module)
	want := mustEvalModule(t, New(), module)
	if !valuesEqual(got, want) {
		t.Fatalf("bytecode non-method member access mismatch: got=%#v want=%#v", got, want)
	}

	stats := interp.BytecodeStats()
	if stats.MemberMethodCacheHits != 0 || stats.MemberMethodCacheMiss != 0 {
		t.Fatalf("expected non-method member access to skip member-method cache counters, got hits=%d misses=%d", stats.MemberMethodCacheHits, stats.MemberMethodCacheMiss)
	}
}
