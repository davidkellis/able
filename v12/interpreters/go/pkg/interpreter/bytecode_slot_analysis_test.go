package interpreter

import (
	"testing"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func TestAnalyzeFrameLayoutCachesParamSimpleTypes(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("String")),
			ast.Param("c", nil),
		},
		[]ast.Statement{
			ast.ID("a"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayout(interp, def)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if got, want := len(layout.paramSimpleTypes), 3; got != want {
		t.Fatalf("unexpected param simple type count: got=%d want=%d", got, want)
	}
	if got := layout.paramSimpleTypes[0]; got != "i32" {
		t.Fatalf("unexpected first param simple type: got=%q want=%q", got, "i32")
	}
	if got := layout.paramSimpleTypes[1]; got != "String" {
		t.Fatalf("unexpected second param simple type: got=%q want=%q", got, "String")
	}
	if got := layout.paramSimpleTypes[2]; got != "" {
		t.Fatalf("unexpected third param simple type: got=%q want empty", got)
	}
}

func TestAnalyzeFrameLayoutCachesParamCellKinds(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("String")),
			ast.Param("c", nil),
		},
		[]ast.Statement{
			ast.ID("a"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayout(interp, def)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if got, want := len(layout.paramKinds), 3; got != want {
		t.Fatalf("unexpected param kind count: got=%d want=%d", got, want)
	}
	if layout.paramKinds[0] != bytecodeCellKindI32 {
		t.Fatalf("expected first param to be i32 cell kind, got %d", layout.paramKinds[0])
	}
	if layout.paramKinds[1] != bytecodeCellKindValue || layout.paramKinds[2] != bytecodeCellKindValue {
		t.Fatalf("expected non-i32 params to stay boxed value kinds, got %d and %d", layout.paramKinds[1], layout.paramKinds[2])
	}
	if !layout.hasTypedSlots {
		t.Fatalf("expected layout to record typed slots")
	}
}

func TestAnalyzeFrameLayoutCachesParamSimpleChecks(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Ty("String")),
			ast.Param("c", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
		},
		[]ast.Statement{
			ast.ID("a"),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayout(interp, def)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if got, want := len(layout.paramSimpleChecks), 3; got != want {
		t.Fatalf("unexpected param simple check count: got=%d want=%d", got, want)
	}
	if got := layout.paramSimpleChecks[0]; got != bytecodeSimpleTypeCheckI32 {
		t.Fatalf("unexpected first param simple check: got=%d want=%d", got, bytecodeSimpleTypeCheckI32)
	}
	if got := layout.paramSimpleChecks[1]; got != bytecodeSimpleTypeCheckString {
		t.Fatalf("unexpected second param simple check: got=%d want=%d", got, bytecodeSimpleTypeCheckString)
	}
	if got := layout.paramSimpleChecks[2]; got != bytecodeSimpleTypeCheckUnknown {
		t.Fatalf("unexpected generic param simple check: got=%d want unknown", got)
	}
}

func TestAnalyzeFrameLayoutCachesParamCoercionMetadata(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Ty("i32")),
			ast.Param("b", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("c", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.ID("a"),
		},
		ast.Ty("i32"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayout(interp, def)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if got, want := len(layout.paramTypes), 3; got != want {
		t.Fatalf("unexpected param type count: got=%d want=%d", got, want)
	}
	if got, want := len(layout.paramNeedsCoercion), 3; got != want {
		t.Fatalf("unexpected coercion metadata count: got=%d want=%d", got, want)
	}
	if !layout.paramNeedsCoercion[0] {
		t.Fatalf("expected concrete primitive param to retain coercion check metadata")
	}
	if layout.paramNeedsCoercion[1] {
		t.Fatalf("expected Array param to skip runtime coercion metadata")
	}
	if layout.paramNeedsCoercion[2] {
		t.Fatalf("expected generic param to skip runtime coercion metadata")
	}
	if !layout.anyParamCoercion {
		t.Fatalf("expected layout to record at least one coercion-bearing parameter")
	}
	if layout.anyExplicitCoercion {
		t.Fatalf("expected only the first parameter to require coercion in this layout")
	}
}

func TestAnalyzeFrameLayoutCachesReturnGenericUse(t *testing.T) {
	interp := New()
	directGeneric := ast.Fn(
		"id",
		[]*ast.FunctionParameter{ast.Param("value", ast.Ty("T"))},
		[]ast.Statement{ast.ID("value")},
		ast.Ty("T"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	layout := analyzeFrameLayout(interp, directGeneric)
	if layout == nil {
		t.Fatalf("expected direct generic return frame layout")
	}
	if !layout.returnTypeUsesGenerics {
		t.Fatalf("expected direct generic return type to be cached")
	}

	nestedGeneric := ast.Fn(
		"maybe",
		nil,
		[]ast.Statement{ast.Nil()},
		ast.Nullable(ast.Ty("T")),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	layout = analyzeFrameLayout(interp, nestedGeneric)
	if layout == nil {
		t.Fatalf("expected nested generic return frame layout")
	}
	if !layout.returnTypeUsesGenerics {
		t.Fatalf("expected nested generic return type to be cached")
	}

	concrete := ast.Fn(
		"answer",
		nil,
		[]ast.Statement{ast.Int(42)},
		ast.Ty("i32"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)
	layout = analyzeFrameLayout(interp, concrete)
	if layout == nil {
		t.Fatalf("expected concrete return frame layout")
	}
	if layout.returnTypeUsesGenerics {
		t.Fatalf("expected concrete return type to avoid generic-use cache flag")
	}
}

func TestAnalyzeFrameLayoutCachesCanonicalReturnType(t *testing.T) {
	interp := New()
	returnType := ast.Result(ast.Ty("String"))
	def := ast.Fn(
		"read_text",
		nil,
		[]ast.Statement{ast.Str("ok")},
		returnType,
		nil,
		nil,
		false,
		false,
	)
	layout := analyzeFrameLayoutWithEnv(interp, def, interp.GlobalEnvironment())
	if layout == nil {
		t.Fatalf("expected return metadata frame layout")
	}
	if layout.returnTypeHasAlias {
		t.Fatalf("returnTypeHasAlias = true, want false")
	}
	if layout.returnCanonicalType == nil {
		t.Fatalf("expected cached canonical return type")
	}
	if !typeExpressionsEqual(layout.returnCanonicalType, returnType) {
		t.Fatalf("canonical return type = %s, want %s", typeExpressionToString(layout.returnCanonicalType), typeExpressionToString(returnType))
	}
}

func TestAnalyzeFrameLayoutCachesAliasExpandedCanonicalReturnType(t *testing.T) {
	interp := New()
	interp.RegisterTypeAlias(
		"AliasText",
		ast.NewTypeAliasDefinition(ast.ID("AliasText"), ast.Ty("String"), nil, nil, false),
	)
	returnType := ast.Result(ast.Ty("AliasText"))
	def := ast.Fn(
		"read_text",
		nil,
		[]ast.Statement{ast.Str("ok")},
		returnType,
		nil,
		nil,
		false,
		false,
	)
	layout := analyzeFrameLayoutWithEnv(interp, def, interp.GlobalEnvironment())
	if layout == nil {
		t.Fatalf("expected alias return metadata frame layout")
	}
	if !layout.returnTypeHasAlias {
		t.Fatalf("returnTypeHasAlias = false, want true")
	}
	want := ast.Result(ast.Ty("String"))
	if !typeExpressionsEqual(layout.returnCanonicalType, want) {
		t.Fatalf("canonical return type = %s, want %s", typeExpressionToString(layout.returnCanonicalType), typeExpressionToString(want))
	}
}

func TestAnalyzeFrameLayoutCachesMethodSetReturnGenericUse(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"next",
		[]*ast.FunctionParameter{ast.Param("self", ast.Ty("Self"))},
		[]ast.Statement{ast.Nil()},
		ast.Nullable(ast.Ty("T")),
		nil,
		nil,
		false,
		false,
	)
	methodSet := &runtime.MethodSet{
		TargetType: ast.Ty("Box"),
		GenericParams: []*ast.GenericParameter{
			ast.GenericParam("T", nil),
		},
	}

	layout := analyzeFrameLayoutWithEnvAndMethodSet(interp, def, nil, methodSet)
	if layout == nil {
		t.Fatalf("expected method-set generic return frame layout")
	}
	if !layout.returnTypeUsesGenerics {
		t.Fatalf("expected method-set generic return type to be cached")
	}
}

func TestAnalyzeFrameLayoutMarksControlFlowPreservationWithoutLoops(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{ast.Param("depth", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("depth"), ast.Int(0)),
				ast.Block(ast.Ret(ast.Int(1))),
			),
			ast.Bin("+", ast.ID("depth"), ast.Int(1)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayout(interp, def)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if !layout.preservesControlFlow {
		t.Fatalf("expected straight-line recursive layout to preserve ambient control stacks")
	}
}

func TestAnalyzeFrameLayoutMarksControlFlowMutationForLoopBodies(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{ast.Param("depth", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Loop(ast.Brk(nil, nil)),
		},
		ast.Ty("void"),
		nil,
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayout(interp, def)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if layout.preservesControlFlow {
		t.Fatalf("expected loop-bearing layout to require full control-stack restoration")
	}
}

func TestAnalyzeFrameLayoutCachesNoCoercionSummaryFlags(t *testing.T) {
	interp := New()
	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{
			ast.Param("a", ast.Gen(ast.Ty("Array"), ast.Ty("i32"))),
			ast.Param("b", ast.Ty("T")),
		},
		[]ast.Statement{
			ast.ID("a"),
		},
		ast.Ty("void"),
		[]*ast.GenericParameter{ast.GenericParam("T")},
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayout(interp, def)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if layout.paramNeedsCoercion[0] || layout.paramNeedsCoercion[1] {
		t.Fatalf("expected both parameters to skip runtime coercion metadata")
	}
	if layout.anyParamCoercion {
		t.Fatalf("expected layout to record no coercion-bearing parameters")
	}
	if layout.anyExplicitCoercion {
		t.Fatalf("expected layout to record no explicit coercion-bearing parameters")
	}
}

func TestInlineCoercionUnnecessaryAcceptsBoxedPrimitivePointers(t *testing.T) {
	intVal := runtime.IntegerValue{TypeSuffix: runtime.IntegerI32}
	floatVal := runtime.FloatValue{TypeSuffix: runtime.FloatF64}
	stringVal := runtime.StringValue{Val: "x"}
	boolVal := runtime.BoolValue{Val: true}

	if !inlineCoercionUnnecessary(ast.Ty("i32"), &intVal) {
		t.Fatalf("expected boxed integer pointer to match i32")
	}
	if !inlineCoercionUnnecessary(ast.Ty("f64"), &floatVal) {
		t.Fatalf("expected boxed float pointer to match f64")
	}
	if !inlineCoercionUnnecessary(ast.Ty("String"), &stringVal) {
		t.Fatalf("expected boxed string pointer to match String")
	}
	if !inlineCoercionUnnecessary(ast.Ty("Bool"), &boolVal) {
		t.Fatalf("expected boxed bool pointer to match Bool")
	}
}

func TestInlineCoercionUnnecessaryBySimpleCheck(t *testing.T) {
	intVal := runtime.NewSmallInt(7, runtime.IntegerI32)
	floatVal := runtime.FloatValue{TypeSuffix: runtime.FloatF64}
	stringVal := runtime.StringValue{Val: "x"}
	boolVal := runtime.BoolValue{Val: true}

	if !inlineCoercionUnnecessaryBySimpleCheck(bytecodeSimpleTypeCheckAnyInteger, intVal) {
		t.Fatalf("expected Int check to accept integer values")
	}
	if !inlineCoercionUnnecessaryBySimpleCheck(bytecodeSimpleTypeCheckI32, intVal) {
		t.Fatalf("expected i32 check to accept matching integer suffix")
	}
	if inlineCoercionUnnecessaryBySimpleCheck(bytecodeSimpleTypeCheckI64, intVal) {
		t.Fatalf("expected i64 check to reject mismatched integer suffix")
	}
	if !inlineCoercionUnnecessaryBySimpleCheck(bytecodeSimpleTypeCheckAnyFloat, &floatVal) {
		t.Fatalf("expected Float check to accept float pointers")
	}
	if !inlineCoercionUnnecessaryBySimpleCheck(bytecodeSimpleTypeCheckString, &stringVal) {
		t.Fatalf("expected String check to accept string pointers")
	}
	if !inlineCoercionUnnecessaryBySimpleCheck(bytecodeSimpleTypeCheckBool, &boolVal) {
		t.Fatalf("expected Bool check to accept bool pointers")
	}
}

func TestAnalyzeFrameLayoutWithEnvAllowsSimpleNamedStructLiteral(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Nullable(ast.Ty("Node")), "left"),
			ast.FieldDef(ast.Nullable(ast.Ty("Node")), "right"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, err := interp.evaluateStructDefinition(nodeDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}

	makeTree := ast.Fn(
		"make_tree",
		[]*ast.FunctionParameter{ast.Param("depth", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("depth"), ast.Int(0)),
				ast.Block(
					ast.Ret(ast.StructLit(
						[]*ast.StructFieldInitializer{
							ast.FieldInit(ast.Nil(), "left"),
							ast.FieldInit(ast.Nil(), "right"),
						},
						false,
						"Node",
						nil,
						nil,
					)),
				),
			),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Call("make_tree", ast.Bin("-", ast.ID("depth"), ast.Int(1))), "left"),
					ast.FieldInit(ast.Call("make_tree", ast.Bin("-", ast.ID("depth"), ast.Int(1))), "right"),
				},
				false,
				"Node",
				nil,
				nil,
			),
		},
		ast.Ty("Node"),
		nil,
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayoutWithEnv(interp, makeTree, env)
	if layout == nil {
		t.Fatalf("expected frame layout for simple named-struct literal function body")
	}
	if layout.selfCallOneArgFast == false {
		t.Fatalf("expected self-call one-arg fast metadata to stay enabled")
	}
	if layout.returnExactStructDef == nil || layout.returnExactStructDef.Node != nodeDef {
		t.Fatalf("expected return exact named-struct definition cache to capture Node")
	}
}

func TestAnalyzeFrameLayoutWithEnvCachesExactNamedStructParamDefinitions(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Nullable(ast.Ty("Node")), "left"),
			ast.FieldDef(ast.Nullable(ast.Ty("Node")), "right"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, err := interp.evaluateStructDefinition(nodeDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}

	def := ast.Fn(
		"check_tree",
		[]*ast.FunctionParameter{ast.Param("node", ast.Ty("Node"))},
		[]ast.Statement{
			ast.Int(1),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	layout := analyzeFrameLayoutWithEnv(interp, def, env)
	if layout == nil {
		t.Fatalf("expected frame layout")
	}
	if got, want := len(layout.paramExactStructDef), 1; got != want {
		t.Fatalf("unexpected exact struct param cache count: got=%d want=%d", got, want)
	}
	if layout.paramExactStructDef[0] == nil || layout.paramExactStructDef[0].Node != nodeDef {
		t.Fatalf("expected first param exact named-struct definition cache to capture Node")
	}
}

func TestEvaluateFunctionDefinition_MakeTreeStaysInlineEligible(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	nodeDef := ast.StructDef(
		"Node",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Nullable(ast.Ty("Node")), "left"),
			ast.FieldDef(ast.Nullable(ast.Ty("Node")), "right"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, err := interp.evaluateStructDefinition(nodeDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}

	makeTree := ast.Fn(
		"make_tree",
		[]*ast.FunctionParameter{ast.Param("depth", ast.Ty("i32"))},
		[]ast.Statement{
			ast.IfExpr(
				ast.Bin("<=", ast.ID("depth"), ast.Int(0)),
				ast.Block(
					ast.Ret(ast.StructLit(
						[]*ast.StructFieldInitializer{
							ast.FieldInit(ast.Nil(), "left"),
							ast.FieldInit(ast.Nil(), "right"),
						},
						false,
						"Node",
						nil,
						nil,
					)),
				),
			),
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.Call("make_tree", ast.Bin("-", ast.ID("depth"), ast.Int(1))), "left"),
					ast.FieldInit(ast.Call("make_tree", ast.Bin("-", ast.ID("depth"), ast.Int(1))), "right"),
				},
				false,
				"Node",
				nil,
				nil,
			),
		},
		ast.Ty("Node"),
		nil,
		nil,
		false,
		false,
	)
	if _, err := interp.evaluateFunctionDefinition(makeTree, env); err != nil {
		t.Fatalf("evaluateFunctionDefinition failed: %v", err)
	}

	raw, ok := env.Lookup("make_tree")
	if !ok {
		t.Fatalf("make_tree not defined")
	}
	fn, ok := raw.(*runtime.FunctionValue)
	if !ok || fn == nil {
		t.Fatalf("make_tree binding = %#v, want *runtime.FunctionValue", raw)
	}
	program, ok := fn.Bytecode.(*bytecodeProgram)
	if !ok || program == nil {
		t.Fatalf("make_tree bytecode = %#v, want *bytecodeProgram", fn.Bytecode)
	}
	if program.frameLayout == nil {
		t.Fatalf("expected make_tree bytecode to keep a frame layout")
	}

	callNode := ast.NewFunctionCall(ast.ID("make_tree"), []ast.Expression{ast.Int(1)}, nil, false)
	entry := bytecodeBuildCallNameCacheEntry(
		"make_tree",
		bytecodeResolvedIdentifierLookup{
			value:        fn,
			env:          env,
			envVersion:   env.RevisionWithHint(interp.envSingleThread),
			owner:        env,
			ownerVersion: env.RevisionWithHint(interp.envSingleThread),
		},
		fn,
		1,
		callNode,
	)
	if entry.dispatch != bytecodeCallNameDispatchInline || !entry.inlineDirect {
		t.Fatalf("expected make_tree call-name cache entry to stay direct-inline eligible, got dispatch=%v inlineDirect=%v", entry.dispatch, entry.inlineDirect)
	}
}

func TestAnalyzeFrameLayoutWithEnvRejectsStructLiteralPlusPlaceholder(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, err := interp.evaluateStructDefinition(boxDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}

	def := ast.Fn(
		"main",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("box"), ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(4), "value")},
				false,
				"Box",
				nil,
				nil,
			)),
			ast.Assign(ast.ID("plus"), ast.CallExpr(ast.Member(ast.ID("box"), "add"), ast.Placeholder())),
			ast.ID("plus"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	if layout := analyzeFrameLayoutWithEnv(interp, def, env); layout != nil {
		t.Fatalf("expected placeholder-bearing function to stay off slot layout when struct literals are present")
	}
}

func TestAnalyzeFrameLayoutWithEnvAllowsStructLiteralPlusDottedSlotMemberCall(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	sDef := ast.StructDef(
		"S",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("i32"), "n"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, err := interp.evaluateStructDefinition(sDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}

	def := ast.Fn(
		"call_get",
		nil,
		[]ast.Statement{
			ast.Assign(ast.ID("s"), ast.StructLit(
				[]*ast.StructFieldInitializer{ast.FieldInit(ast.Int(1), "n")},
				false,
				"S",
				nil,
				nil,
			)),
			ast.Call("s.get"),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	if layout := analyzeFrameLayoutWithEnv(interp, def, env); layout == nil {
		t.Fatalf("expected dotted identifier call on local head to use slot layout")
	}
	program, err := interp.lowerFunctionDefinitionBytecodeWithEnv(def, env)
	if err != nil {
		t.Fatalf("lowerFunctionDefinitionBytecodeWithEnv failed: %v", err)
	}
	if program == nil || program.frameLayout == nil {
		t.Fatalf("expected lowered function to keep slot layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpCallMember) {
		t.Fatalf("expected dotted local-head call to lower as CallMember")
	}
	if bytecodeProgramContainsOpcode(program, bytecodeOpCallName) {
		t.Fatalf("dotted local-head call should not lower as CallName")
	}
}

func TestAnalyzeFrameLayoutWithEnvAllowsSingletonStructLiteral(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	emptyDef := ast.StructDef("Empty", nil, ast.StructKindSingleton, nil, nil, false)
	boxDef := ast.StructDef(
		"Box",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("Empty"), "value"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	if _, err := interp.evaluateStructDefinition(emptyDef, env); err != nil {
		t.Fatalf("evaluate Empty definition failed: %v", err)
	}
	if _, err := interp.evaluateStructDefinition(boxDef, env); err != nil {
		t.Fatalf("evaluate Box definition failed: %v", err)
	}

	def := ast.Fn(
		"make_box",
		nil,
		[]ast.Statement{
			ast.StructLit(
				[]*ast.StructFieldInitializer{
					ast.FieldInit(ast.StructLit(nil, false, "Empty", nil, nil), "value"),
				},
				false,
				"Box",
				nil,
				nil,
			),
		},
		nil,
		nil,
		nil,
		false,
		false,
	)

	if layout := analyzeFrameLayoutWithEnv(interp, def, env); layout == nil {
		t.Fatalf("expected singleton struct literal to stay slot eligible")
	}
	program, err := interp.lowerFunctionDefinitionBytecodeWithEnv(def, env)
	if err != nil {
		t.Fatalf("lowerFunctionDefinitionBytecodeWithEnv failed: %v", err)
	}
	if program == nil || program.frameLayout == nil {
		t.Fatalf("expected lowered function to keep slot layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpStructLiteralNamedFast) {
		t.Fatalf("expected singleton/named struct literals to use fast literal opcode")
	}
}

func TestAnalyzeFrameLayoutWithEnvRejectsLocalSingletonIdentifierMatchAmbiguity(t *testing.T) {
	interp := NewBytecode()

	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.StructDef("Done", nil, ast.StructKindSingleton, nil, nil, false),
			ast.Match(
				ast.Nil(),
				ast.Mc(ast.ID("Done"), ast.Int(1)),
				ast.Mc(ast.Wc(), ast.Int(2)),
			),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	if layout := analyzeFrameLayoutWithEnv(interp, def, interp.GlobalEnvironment()); layout != nil {
		t.Fatalf("expected local singleton identifier match ambiguity to stay off slot layout")
	}
}

func TestAnalyzeFrameLayoutWithEnvRejectsGlobalSingletonIdentifierMatchAmbiguity(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	doneDef := ast.StructDef("Done", nil, ast.StructKindSingleton, nil, nil, false)
	if _, err := interp.evaluateStructDefinition(doneDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}

	def := ast.Fn(
		"f",
		nil,
		[]ast.Statement{
			ast.Match(
				ast.Nil(),
				ast.Mc(ast.ID("Done"), ast.Int(1)),
				ast.Mc(ast.Wc(), ast.Int(2)),
			),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	if layout := analyzeFrameLayoutWithEnv(interp, def, env); layout != nil {
		t.Fatalf("expected global singleton identifier match ambiguity to stay off slot layout")
	}
	program, err := interp.lowerFunctionDefinitionBytecodeWithEnv(def, env)
	if err != nil {
		t.Fatalf("lowerFunctionDefinitionBytecodeWithEnv failed: %v", err)
	}
	if program == nil {
		t.Fatalf("expected bytecode program")
	}
	if program.frameLayout != nil {
		t.Fatalf("expected lowered function to avoid slot layout")
	}
	if !bytecodeProgramContainsOpcode(program, bytecodeOpMatch) {
		t.Fatalf("expected unsupported identifier singleton match to use generic match opcode")
	}
}

func TestAnalyzeFrameLayoutWithEnvAllowsParamShadowedSingletonIdentifierMatch(t *testing.T) {
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	doneDef := ast.StructDef("Done", nil, ast.StructKindSingleton, nil, nil, false)
	if _, err := interp.evaluateStructDefinition(doneDef, env); err != nil {
		t.Fatalf("evaluateStructDefinition failed: %v", err)
	}

	def := ast.Fn(
		"f",
		[]*ast.FunctionParameter{ast.Param("Done", ast.Ty("i32"))},
		[]ast.Statement{
			ast.Match(
				ast.ID("Done"),
				ast.Mc(ast.ID("Done"), ast.Bin("+", ast.ID("Done"), ast.Int(1))),
			),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	if layout := analyzeFrameLayoutWithEnv(interp, def, env); layout == nil {
		t.Fatalf("expected param-shadowed singleton identifier match to keep slot layout")
	}
}

func TestInlineParamCoercionUnnecessaryUsesCachedSimpleCheck(t *testing.T) {
	layout := &bytecodeFrameLayout{
		paramSimpleTypes:  []string{"String"},
		paramSimpleChecks: []bytecodeSimpleTypeCheck{bytecodeSimpleTypeCheckI32},
	}
	value := runtime.NewSmallInt(7, runtime.IntegerI32)

	if !inlineParamCoercionUnnecessary(nil, layout, 0, ast.Ty("String"), value) {
		t.Fatalf("expected cached i32 simple check to accept matching integer value")
	}
}

func TestInlineCoerceValueBySimpleTypeIntegerWidening(t *testing.T) {
	value := runtime.NewSmallInt(7, runtime.IntegerI32)

	coerced, ok, err := inlineCoerceValueBySimpleType("i64", value)
	if err != nil {
		t.Fatalf("unexpected coercion error: %v", err)
	}
	if !ok {
		t.Fatalf("expected integer widening to be handled")
	}
	intVal, ok := coerced.(runtime.IntegerValue)
	if !ok {
		t.Fatalf("expected widened integer value, got %T", coerced)
	}
	if intVal.TypeSuffix != runtime.IntegerI64 {
		t.Fatalf("unexpected widened suffix: got=%s want=%s", intVal.TypeSuffix, runtime.IntegerI64)
	}
	if got, fits := intVal.ToInt64(); !fits || got != 7 {
		t.Fatalf("unexpected widened integer payload: got=%d fits=%v", got, fits)
	}
}

func TestInlineCoerceValueBySimpleTypeIntegerToFloat(t *testing.T) {
	value := runtime.NewSmallInt(7, runtime.IntegerI32)

	coerced, ok, err := inlineCoerceValueBySimpleType("f64", value)
	if err != nil {
		t.Fatalf("unexpected coercion error: %v", err)
	}
	if !ok {
		t.Fatalf("expected integer-to-float coercion to be handled")
	}
	floatVal, ok := coerced.(runtime.FloatValue)
	if !ok {
		t.Fatalf("expected float value, got %T", coerced)
	}
	if floatVal.TypeSuffix != runtime.FloatF64 || floatVal.Val != 7 {
		t.Fatalf("unexpected float coercion result: got=%#v", floatVal)
	}
}
