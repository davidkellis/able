package typechecker

import (
	"able/interpreter-go/pkg/ast"
	"strings"
	"testing"
)

func TestBlockExpressionScopesBindings(t *testing.T) {
	checker := New()
	block := ast.Block(
		ast.Assign(ast.ID("inner"), ast.Int(5)),
		ast.ID("inner"),
	)
	module := ast.NewModule([]ast.Statement{block}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[block]
	if !ok {
		t.Fatalf("expected block inference entry")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected block to return i32, got %q", typeName(typ))
	}
	if _, exists := checker.global.Lookup("inner"); exists {
		t.Fatalf("expected inner binding to remain scoped to block")
	}
}
func TestIfExpressionMergesBranchTypes(t *testing.T) {
	checker := New()
	ifExpr := ast.IfExpr(
		ast.Bool(true),
		ast.Block(ast.Int(1)),
		ast.ElseIf(ast.Block(ast.Int(2)), ast.Bool(false)),
	)
	ifExpr.ElseBody = ast.Block(ast.Int(3))
	assign := ast.Assign(ast.ID("value"), ifExpr)
	module := ast.NewModule([]ast.Statement{assign}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[ifExpr]
	if !ok {
		t.Fatalf("expected inference for if expression")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected if expression to have type i32, got %q", typeName(typ))
	}
}
func TestIfExpressionConditionMustBeBool(t *testing.T) {
	checker := New()
	ifExpr := ast.IfExpr(
		ast.Int(1),
		ast.Block(ast.Int(2)),
	)
	module := ast.NewModule([]ast.Statement{ifExpr}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostics for non-bool condition")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "condition must be bool") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected condition bool diagnostic, got %v", diags)
	}
}
func TestReturnOutsideFunctionProducesDiagnostic(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{ast.Ret(nil)}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for return outside function")
	}
	if want := "outside function"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected diagnostic mentioning %q, got %v", want, diags)
	}
}
func TestRescueExpressionMergesTypes(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Str("ok"))
	rescue := ast.Rescue(
		ast.ID("value"),
		ast.Mc(ast.Wc(), ast.Str("fallback")),
	)
	module := ast.NewModule([]ast.Statement{assign, rescue}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[rescue]
	if !ok {
		t.Fatalf("expected inference entry for rescue expression")
	}
	if typeName(typ) != "String" {
		t.Fatalf("expected rescue expression to infer String, got %q", typeName(typ))
	}
}
func TestRescueGuardMustBeBool(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Str("ok"))
	rescue := ast.Rescue(
		ast.ID("value"),
		ast.Mc(ast.Wc(), ast.Str("fallback"), ast.Int(1)),
	)
	module := ast.NewModule([]ast.Statement{assign, rescue}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "rescue guard must evaluate to bool") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected rescue guard diagnostic, got %v", diags)
	}
}

func TestFunctionBodyReportsLiteralOverflowForAnnotatedReturn(t *testing.T) {
	checker := New()
	_, literalType := checker.checkExpression(nil, ast.Int(512))
	expectedType := checker.resolveTypeReference(ast.Ty("u8"))
	if msg, ok := literalMismatchMessage(literalType, expectedType); !ok || !strings.Contains(msg, "literal 512") {
		t.Fatalf("expected literal mismatch helper to flag overflow, got ok=%v msg=%q", ok, msg)
	}
	ret := ast.Ret(ast.Int(512))
	fn := ast.Fn(
		"make_byte",
		nil,
		[]ast.Statement{ret},
		ast.Ty("u8"),
		nil,
		nil,
		false,
		false,
	)
	module := ast.NewModule([]ast.Statement{fn}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if inferred, ok := checker.infer[ret.Argument]; ok {
		if msg, ok := literalMismatchMessage(inferred, expectedType); !ok || !strings.Contains(msg, "literal 512") {
			t.Fatalf("expected inferred return argument to carry literal info, got %v", inferred)
		}
	} else {
		t.Fatalf("expected inference entry for return argument")
	}
	if len(diags) != 1 {
		t.Fatalf("expected diagnostic for literal overflow, got %v", diags)
	}
	if !strings.Contains(diags[0].Message, "literal 512 does not fit in u8") {
		t.Fatalf("expected literal overflow message, got %q", diags[0].Message)
	}
}
func TestOrElseExpressionMergesTypes(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Str("ok"))
	orElse := ast.OrElseBlock(
		ast.ID("value"),
		ast.Block(ast.Str("fallback")),
		nil,
	)
	module := ast.NewModule([]ast.Statement{assign, orElse}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[orElse]
	if !ok {
		t.Fatalf("expected inference entry for or-else expression")
	}
	if typeName(typ) != "String" {
		t.Fatalf("expected or-else expression to infer String, got %q", typeName(typ))
	}
}
func TestOrElseProducesUnionWhenTypesDiffer(t *testing.T) {
	checker := New()
	orElse := ast.OrElseBlock(
		ast.Int(1),
		ast.Block(ast.Str("fallback")),
		nil,
	)
	module := ast.NewModule([]ast.Statement{orElse}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[orElse]
	if !ok {
		t.Fatalf("expected inference entry for or-else expression")
	}
	if typeName(typ) != "i32 | String" {
		t.Fatalf("expected union literal for mismatched or-else types, got %q", typeName(typ))
	}
}
func TestOrElseBindsErrorInHandler(t *testing.T) {
	checker := New()
	errorStruct := ast.StructDef(
		"Error",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("String"), "message"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	assign := ast.Assign(ast.ID("value"), ast.Str("ok"))
	orElse := ast.OrElseBlock(
		ast.ID("value"),
		ast.Block(ast.Member(ast.ID("err"), "message")),
		"err",
	)
	module := ast.NewModule([]ast.Statement{errorStruct, assign, orElse}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for or-else handler, got %v", diags)
	}
}
func TestEnsureExpressionReturnsTryType(t *testing.T) {
	checker := New()
	assign := ast.Assign(ast.ID("value"), ast.Int(1))
	ensure := ast.Ensure(
		ast.ID("value"),
		ast.Assign(ast.ID("cleanup"), ast.Int(0)),
	)
	module := ast.NewModule([]ast.Statement{assign, ensure}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[ensure]
	if !ok {
		t.Fatalf("expected inference entry for ensure expression")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected ensure expression to infer i32, got %q", typeName(typ))
	}
}
func TestRaiseAllowsAnyValue(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{ast.Raise(ast.Int(1))}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for raise with non-error value, got %v", diags)
	}
}
func TestRaiseAcceptsErrorStruct(t *testing.T) {
	checker := New()
	errorStruct := ast.StructDef(
		"Error",
		[]*ast.StructFieldDefinition{
			ast.FieldDef(ast.Ty("String"), "message"),
		},
		ast.StructKindNamed,
		nil,
		nil,
		false,
	)
	raise := ast.Raise(
		ast.StructLit(
			[]*ast.StructFieldInitializer{
				ast.FieldInit(ast.Str("boom"), "message"),
			},
			false,
			"Error",
			nil,
			nil,
		),
	)
	module := ast.NewModule([]ast.Statement{errorStruct, raise}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for valid raise, got %v", diags)
	}
	if len(checker.obligations) != 0 {
		t.Fatalf("expected no obligations for raise check, got %v", checker.obligations)
	}
}
func TestBreakRequiresLoop(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{ast.Brk(nil, nil)}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for break outside loop")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "break statement must appear inside a loop") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing break diagnostic: %v", diags)
	}
}
func TestContinueRequiresLoop(t *testing.T) {
	checker := New()
	module := ast.NewModule([]ast.Statement{ast.Cont(nil)}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for continue outside loop")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "continue statement must appear inside a loop") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("missing continue diagnostic: %v", diags)
	}
}
func TestWhileConditionMustBeBool(t *testing.T) {
	checker := New()
	loop := ast.Wloop(
		ast.Int(1),
		ast.Block(),
	)
	module := ast.NewModule([]ast.Statement{loop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) == 0 {
		t.Fatalf("expected diagnostic for while condition")
	}
	if want := "while condition must be bool"; !strings.Contains(diags[0].Message, want) {
		t.Fatalf("expected message %q, got %q", want, diags[0].Message)
	}
}
func TestForLoopIterableMustBeArrayRangeOrIterator(t *testing.T) {
	checker := New()
	loop := ast.ForIn(ast.ID("value"), ast.Int(5))
	module := ast.NewModule([]ast.Statement{loop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "for-loop iterable must be array, range, String, or iterator") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected for-loop iterable diagnostic, got %v", diags)
	}
}
func TestForLoopRangeElementType(t *testing.T) {
	checker := New()
	pattern := ast.ID("value")
	bodyExpr := ast.ID("value")
	loop := ast.ForLoopPattern(
		pattern,
		ast.Range(ast.Int(0), ast.Int(3), false),
		ast.Block(bodyExpr),
	)
	module := ast.NewModule([]ast.Statement{loop}, nil, nil)
	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics, got %v", diags)
	}
	typ, ok := checker.infer[bodyExpr]
	if !ok {
		t.Fatalf("expected inference for range loop body identifier")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected loop value to have type i32, got %q", typeName(typ))
	}
}

func TestForLoopIterableSupportsStdlibCollections(t *testing.T) {
	checker := New()
	listStruct := buildGenericStructDefinition("List", []string{"T"})
	linkedListStruct := buildGenericStructDefinition("LinkedList", []string{"T"})
	lazySeqStruct := buildGenericStructDefinition("LazySeq", []string{"T"})
	vectorStruct := buildGenericStructDefinition("Vector", []string{"T"})
	hashSetStruct := buildGenericStructDefinition("HashSet", []string{"T"})
	dequeStruct := buildGenericStructDefinition("Deque", []string{"T"})
	queueStruct := buildGenericStructDefinition("Queue", []string{"T"})
	channelStruct := buildGenericStructDefinition("Channel", []string{"T"})
	bitSetStruct := buildGenericStructDefinition("BitSet", nil)

	listValue := ast.ID("value")
	linkedValue := ast.ID("linkedValue")
	lazyValue := ast.ID("lazyValue")
	listLoop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("value")), ast.Ty("String")),
		ast.ID("items"),
		ast.Block(listValue),
	)
	linkedListLoop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("linkedValue")), ast.Ty("i32")),
		ast.ID("linkedItems"),
		ast.Block(linkedValue),
	)
	lazySeqLoop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("lazyValue")), ast.Ty("String")),
		ast.ID("lazyItems"),
		ast.Block(lazyValue),
	)

	vectorValue := ast.ID("item")
	vectorLoop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("item")), ast.Ty("i32")),
		ast.ID("values"),
		ast.Block(vectorValue),
	)

	setValue := ast.ID("entry")
	setLoop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("entry")), ast.Ty("String")),
		ast.ID("entries"),
		ast.Block(setValue),
	)

	dequeValue := ast.ID("dequeValue")
	dequeLoop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("dequeValue")), ast.Ty("String")),
		ast.ID("dequeItems"),
		ast.Block(dequeValue),
	)

	queueValue := ast.ID("queueValue")
	queueLoop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("queueValue")), ast.Ty("i32")),
		ast.ID("queueItems"),
		ast.Block(queueValue),
	)
	channelValue := ast.ID("channelValue")
	channelLoop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("channelValue")), ast.Ty("String")),
		ast.ID("channelItems"),
		ast.Block(channelValue),
	)
	bitValue := ast.ID("bit")
	bitSetLoop := ast.ForLoopPattern(
		ast.TypedP(ast.PatternFrom(ast.ID("bit")), ast.Ty("i32")),
		ast.ID("bitset"),
		ast.Block(bitValue),
	)

	listFn := ast.Fn(
		"consumeList",
		[]*ast.FunctionParameter{
			ast.Param("items", ast.Gen(ast.Ty("List"), ast.Ty("String"))),
		},
		[]ast.Statement{
			listLoop,
			ast.Ret(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	linkedListFn := ast.Fn(
		"consumeLinkedList",
		[]*ast.FunctionParameter{
			ast.Param("linkedItems", ast.Gen(ast.Ty("LinkedList"), ast.Ty("i32"))),
		},
		[]ast.Statement{
			linkedListLoop,
			ast.Ret(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	lazySeqFn := ast.Fn(
		"consumeLazySeq",
		[]*ast.FunctionParameter{
			ast.Param("lazyItems", ast.Gen(ast.Ty("LazySeq"), ast.Ty("String"))),
		},
		[]ast.Statement{
			lazySeqLoop,
			ast.Ret(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	vectorFn := ast.Fn(
		"consumeVector",
		[]*ast.FunctionParameter{
			ast.Param("values", ast.Gen(ast.Ty("Vector"), ast.Ty("i32"))),
		},
		[]ast.Statement{
			vectorLoop,
			ast.Ret(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	setFn := ast.Fn(
		"consumeHashSet",
		[]*ast.FunctionParameter{
			ast.Param("entries", ast.Gen(ast.Ty("HashSet"), ast.Ty("String"))),
		},
		[]ast.Statement{
			setLoop,
			ast.Ret(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	dequeFn := ast.Fn(
		"consumeDeque",
		[]*ast.FunctionParameter{
			ast.Param("dequeItems", ast.Gen(ast.Ty("Deque"), ast.Ty("String"))),
		},
		[]ast.Statement{
			dequeLoop,
			ast.Ret(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	queueFn := ast.Fn(
		"consumeQueue",
		[]*ast.FunctionParameter{
			ast.Param("queueItems", ast.Gen(ast.Ty("Queue"), ast.Ty("i32"))),
		},
		[]ast.Statement{
			queueLoop,
			ast.Ret(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	channelFn := ast.Fn(
		"consumeChannel",
		[]*ast.FunctionParameter{
			ast.Param("channelItems", ast.Gen(ast.Ty("Channel"), ast.Ty("String"))),
		},
		[]ast.Statement{
			channelLoop,
			ast.Ret(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)
	bitSetFn := ast.Fn(
		"consumeBitSet",
		[]*ast.FunctionParameter{
			ast.Param("bitset", ast.Ty("BitSet")),
		},
		[]ast.Statement{
			bitSetLoop,
			ast.Ret(ast.Int(0)),
		},
		ast.Ty("i32"),
		nil,
		nil,
		false,
		false,
	)

	module := ast.NewModule(
		[]ast.Statement{
			listStruct,
			linkedListStruct,
			lazySeqStruct,
			vectorStruct,
			hashSetStruct,
			dequeStruct,
			queueStruct,
			channelStruct,
			bitSetStruct,
			listFn,
			linkedListFn,
			lazySeqFn,
			vectorFn,
			setFn,
			dequeFn,
			queueFn,
			channelFn,
			bitSetFn,
		},
		nil,
		nil,
	)

	diags, err := checker.CheckModule(module)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(diags) != 0 {
		t.Fatalf("expected no diagnostics for stdlib collection iterables, got %v", diags)
	}

	typ, ok := checker.infer[listValue]
	if !ok {
		t.Fatalf("expected inference for list loop value")
	}
	if typeName(typ) != "String" {
		t.Fatalf("expected list loop value to infer String, got %q", typeName(typ))
	}

	typ, ok = checker.infer[linkedValue]
	if !ok {
		t.Fatalf("expected inference for linked list loop value")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected linked list loop value to infer i32, got %q", typeName(typ))
	}

	typ, ok = checker.infer[lazyValue]
	if !ok {
		t.Fatalf("expected inference for lazy seq loop value")
	}
	if typeName(typ) != "String" {
		t.Fatalf("expected lazy seq loop value to infer String, got %q", typeName(typ))
	}

	typ, ok = checker.infer[vectorValue]
	if !ok {
		t.Fatalf("expected inference for vector loop value")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected vector loop value to infer i32, got %q", typeName(typ))
	}

	typ, ok = checker.infer[setValue]
	if !ok {
		t.Fatalf("expected inference for hash set loop value")
	}
	if typeName(typ) != "String" {
		t.Fatalf("expected hash set loop value to infer String, got %q", typeName(typ))
	}

	typ, ok = checker.infer[dequeValue]
	if !ok {
		t.Fatalf("expected inference for deque loop value")
	}
	if typeName(typ) != "String" {
		t.Fatalf("expected deque loop value to infer String, got %q", typeName(typ))
	}

	typ, ok = checker.infer[queueValue]
	if !ok {
		t.Fatalf("expected inference for queue loop value")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected queue loop value to infer i32, got %q", typeName(typ))
	}

	typ, ok = checker.infer[channelValue]
	if !ok {
		t.Fatalf("expected inference for channel loop value")
	}
	if typeName(typ) != "String" {
		t.Fatalf("expected channel loop value to infer String, got %q", typeName(typ))
	}

	typ, ok = checker.infer[bitValue]
	if !ok {
		t.Fatalf("expected inference for bit set loop value")
	}
	if typeName(typ) != "i32" {
		t.Fatalf("expected bit set loop value to infer i32, got %q", typeName(typ))
	}
}
