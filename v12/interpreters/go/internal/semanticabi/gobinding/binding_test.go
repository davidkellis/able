package gobinding_test

import (
	"math/big"
	"testing"

	"able/interpreter-go/internal/semanticabi/gobinding"
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

func TestArrayAliasAndCycleRoundTrip(t *testing.T) {
	shared := &runtime.ArrayValue{Elements: []runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)}}
	root := &runtime.ArrayValue{}
	root.Elements = []runtime.Value{shared, shared, root}
	decoded, _, err := gobinding.RoundTrip(root)
	if err != nil {
		t.Fatal(err)
	}
	got := decoded[0].(*runtime.ArrayValue)
	if got == root {
		t.Fatal("round trip retained original root pointer")
	}
	left, right := got.Elements[0].(*runtime.ArrayValue), got.Elements[1].(*runtime.ArrayValue)
	if left != right {
		t.Fatal("shared alias split into two arrays")
	}
	if got.Elements[2].(*runtime.ArrayValue) != got {
		t.Fatal("array self-cycle was not restored")
	}
	left.Elements[0] = runtime.NewSmallInt(9, runtime.IntegerI32)
	if value, _ := right.Elements[0].(runtime.IntegerValue).ToInt64(); value != 9 {
		t.Fatalf("alias mutation = %d, want 9", value)
	}
}

func TestRecursiveClosureRoundTripCallsInBothInterpreters(t *testing.T) {
	for _, engine := range []*interpreter.Interpreter{interpreter.New(), interpreter.NewBytecode()} {
		module := ast.Mod([]ast.Statement{
			ast.Assign(ast.ID("captured"), ast.Int(41)),
			ast.Lam(nil, ast.ID("captured")),
		}, nil, nil)
		value, _, err := engine.EvaluateModule(module)
		if err != nil {
			t.Fatal(err)
		}
		fn := value.(*runtime.FunctionValue)
		fn.Closure.DefineWithoutMerge("self", fn)
		decoded, _, err := gobinding.RoundTrip(fn)
		if err != nil {
			t.Fatal(err)
		}
		got := decoded[0].(*runtime.FunctionValue)
		self, ok := got.Closure.LookupInCurrentScope("self")
		if !ok || self != got {
			t.Fatal("recursive closure binding did not point to restored function")
		}
		result, err := engine.CallFunction(got, nil)
		if err != nil {
			t.Fatal(err)
		}
		integer, ok := result.(runtime.IntegerValue)
		if !ok {
			t.Fatalf("call result = %T", result)
		}
		number, _ := integer.ToInt64()
		if number != 41 {
			t.Fatalf("call result = %d, want 41", number)
		}
	}
}

func TestNominalErrorInterfaceAndFutureRoundTrip(t *testing.T) {
	definition := &runtime.StructDefinitionValue{Node: ast.StructDef("Box", []*ast.StructFieldDefinition{ast.FieldDef(ast.Ty("i32"), "value")}, ast.StructKindNamed, nil, nil, false), NamedFieldIndices: map[string]int{"value": 0}}
	instance := &runtime.StructInstanceValue{Definition: definition, Fields: map[string]runtime.Value{"value": runtime.NewSmallInt(5, runtime.IntegerI32)}}
	ifaceDef := &runtime.InterfaceDefinitionValue{Node: ast.Iface("Show", nil, nil, nil, nil, nil, false), QualifiedName: "example.Show"}
	iface := &runtime.InterfaceValue{Interface: ifaceDef, Underlying: instance, Methods: map[string]runtime.Value{}, SharedMethods: map[string]runtime.Value{}}
	errValue := &runtime.ErrorValue{TypeName: ast.ID("Wrapped"), Payload: map[string]runtime.Value{"value": iface}, Message: "wrapped"}
	pkg := &runtime.PackageValue{Name: "example", IdentityKey: "example", Public: map[string]runtime.Value{"box": instance, "failure": errValue}}
	future := runtime.NewFuture()
	future.MarkStarted()
	future.Resolve(instance)
	decoded, snapshot, err := gobinding.RoundTrip(pkg, errValue, future)
	if err != nil {
		t.Fatal(err)
	}
	gotPackage := decoded[0].(*runtime.PackageValue)
	gotInstance := gotPackage.Public["box"].(*runtime.StructInstanceValue)
	gotError := decoded[1].(*runtime.ErrorValue)
	gotInterface := gotError.Payload["value"].(*runtime.InterfaceValue)
	if gotInterface.Underlying != gotInstance {
		t.Fatal("nominal identity split between package and Error/Interface graph")
	}
	gotFuture := decoded[2].(*runtime.FutureValue)
	result, _, status := gotFuture.Snapshot()
	if status != runtime.FutureResolved || result != gotInstance {
		t.Fatalf("future = (%T,%v), want shared instance/resolved", result, status)
	}
	collection, err := snapshot.Heap.Collect()
	if err != nil || collection.Reachable == 0 {
		t.Fatalf("host/root collection = %+v: %v", collection, err)
	}
}

func TestBothInterpretersProduceAliasGraphsThatRoundTrip(t *testing.T) {
	module := ast.Mod([]ast.Statement{
		ast.Assign(ast.ID("shared"), ast.Arr(ast.Int(1), ast.Int(2))),
		ast.Assign(ast.ID("root"), ast.Arr(ast.ID("shared"), ast.ID("shared"))),
		ast.ID("root"),
	}, nil, nil)
	for _, engine := range []*interpreter.Interpreter{interpreter.New(), interpreter.NewBytecode()} {
		value, _, err := engine.EvaluateModule(module)
		if err != nil {
			t.Fatal(err)
		}
		decoded, _, err := gobinding.RoundTrip(value)
		if err != nil {
			t.Fatal(err)
		}
		root := decoded[0].(*runtime.ArrayValue)
		if root.Elements[0].(*runtime.ArrayValue) != root.Elements[1].(*runtime.ArrayValue) {
			t.Fatal("interpreter-produced alias split")
		}
	}
}

func TestWideIntegerHasherAndIteratorRoundTrip(t *testing.T) {
	wide := new(big.Int).Lsh(big.NewInt(1), 100)
	wide.Neg(wide)
	hasher := runtime.NewHasherValueFromState(0xfeedbeef)
	retained := &runtime.ArrayValue{Elements: []runtime.Value{runtime.NewSmallInt(9, runtime.IntegerI32)}}
	position, finalized := 1, false
	iterator := runtime.NewIteratorValueFromHostDriver(runtime.IteratorHostDriver{
		Next: func() (runtime.Value, bool, error) {
			if position >= 3 {
				return runtime.IteratorEnd, true, nil
			}
			position++
			return runtime.NewSmallInt(int64(position), runtime.IntegerI32), false, nil
		},
		Finalize: func() { finalized = true },
		Retained: []runtime.Value{retained},
	}, false)
	decoded, snapshot, err := gobinding.RoundTrip(runtime.NewBigIntValue(wide, runtime.IntegerI128), hasher, iterator, retained)
	if err != nil {
		t.Fatal(err)
	}
	gotWide := decoded[0].(runtime.IntegerValue)
	if gotWide.TypeSuffix != runtime.IntegerI128 || gotWide.BigInt().Cmp(wide) != 0 {
		t.Fatalf("wide integer = %s %s", gotWide.BigInt(), gotWide.TypeSuffix)
	}
	if decoded[1].(*runtime.HasherValue).SemanticState() != hasher.SemanticState() {
		t.Fatal("Hasher semantic state changed")
	}
	gotIterator := decoded[2].(*runtime.IteratorValue)
	driver, closed := gotIterator.HostDriverSnapshot()
	if closed || len(driver.Retained) != 1 || driver.Retained[0] != decoded[3] {
		t.Fatal("iterator retained root identity changed")
	}
	next, done, err := gotIterator.Next()
	if err != nil || done {
		t.Fatalf("iterator next = (%v,%v,%v)", next, done, err)
	}
	number, _ := next.(runtime.IntegerValue).ToInt64()
	if number != 2 {
		t.Fatalf("iterator resumed at %d, want 2", number)
	}
	gotIterator.Close()
	if !finalized {
		t.Fatal("iterator finalizer was not retained")
	}
	collection, err := snapshot.Heap.Collect()
	if err != nil || collection.Reachable == 0 {
		t.Fatalf("indirect/host root collection = %+v: %v", collection, err)
	}
}

func TestMappedKindInventoryRoundTrips(t *testing.T) {
	fn := &runtime.FunctionValue{Declaration: ast.Lam(nil, ast.Nil())}
	structDef := &runtime.StructDefinitionValue{Node: ast.StructDef("Empty", nil, ast.StructKindNamed, nil, nil, false), NamedFieldIndices: map[string]int{}}
	interfaceDef := &runtime.InterfaceDefinitionValue{Node: ast.Iface("EmptyInterface", nil, nil, nil, nil, nil, false)}
	native := runtime.NativeFunctionValue{Name: "identity", Arity: 1}
	values := []runtime.Value{
		runtime.BoolValue{Val: true}, runtime.CharValue{Val: 'a'}, runtime.NilValue{}, runtime.VoidValue{},
		runtime.FloatValue{Val: 1.5, TypeSuffix: runtime.FloatF64}, runtime.IteratorEnd,
		runtime.StringValue{Val: "able"}, &runtime.ArrayValue{}, &runtime.HashMapValue{}, fn,
		runtime.NewHasherValueFromState(7),
		&runtime.FunctionOverloadValue{Overloads: []*runtime.FunctionValue{fn}}, structDef,
		&runtime.TypeRefValue{TypeName: "i32"}, &runtime.StructInstanceValue{Definition: structDef, Fields: map[string]runtime.Value{}},
		interfaceDef, &runtime.InterfaceValue{Interface: interfaceDef, Underlying: runtime.NilValue{}, Methods: map[string]runtime.Value{}, SharedMethods: map[string]runtime.Value{}},
		&runtime.UnionDefinitionValue{Node: ast.UnionDef("Maybe", []ast.TypeExpression{ast.Ty("nil")}, nil, nil, false)},
		&runtime.PackageValue{Name: "p", Public: map[string]runtime.Value{}}, &runtime.DynPackageValue{Name: "p"}, &runtime.DynRefValue{Package: "p", Name: "x"},
		&runtime.ErrorValue{TypeName: ast.ID("Failure"), Payload: map[string]runtime.Value{}},
		&runtime.BoundMethodValue{Receiver: runtime.NilValue{}, Method: native},
		&runtime.ImplementationNamespaceValue{Methods: map[string]runtime.Value{}},
		runtime.NewIteratorValue(nil, nil),
		&runtime.PartialFunctionValue{Target: fn},
	}
	decoded, _, err := gobinding.RoundTrip(values...)
	if err != nil {
		t.Fatal(err)
	}
	if len(decoded) != len(values) {
		t.Fatalf("decoded kinds = %d, want %d", len(decoded), len(values))
	}
	for index := range values {
		if decoded[index].Kind() != values[index].Kind() {
			t.Fatalf("kind[%d] = %s, want %s", index, decoded[index].Kind(), values[index].Kind())
		}
	}

	future := runtime.NewFuture()
	future.Resolve(runtime.NilValue{})
	hosts := []runtime.Value{native, &runtime.HostHandleValue{HandleType: "test"}, runtime.NativeBoundMethodValue{Receiver: runtime.NilValue{}, Method: native}, future}
	decoded, _, err = gobinding.RoundTrip(hosts...)
	if err != nil {
		t.Fatal(err)
	}
	for index := range hosts {
		if decoded[index].Kind() != hosts[index].Kind() {
			t.Fatalf("host kind[%d] = %s, want %s", index, decoded[index].Kind(), hosts[index].Kind())
		}
	}
}
