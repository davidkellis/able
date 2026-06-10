package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"math/big"
	"os"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/internal/semanticabi/gobinding"
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/interpreter"
	"able/interpreter-go/pkg/runtime"
)

type report struct {
	Schema                    string    `json:"schema"`
	Decision                  string    `json:"decision"`
	ManifestIdentity          string    `json:"manifest_identity"`
	RuntimeKinds              int       `json:"runtime_kinds"`
	ExactKinds                int       `json:"exact_kinds"`
	ConditionalKinds          int       `json:"conditional_kinds"`
	OpaqueHostKinds           int       `json:"opaque_host_kinds"`
	BlockedKinds              int       `json:"blocked_kinds"`
	PackageLayoutReconciled   bool      `json:"package_layout_reconciled"`
	TwoPhaseGraphConstruction bool      `json:"two_phase_graph_construction"`
	OwnedWideScalars          bool      `json:"owned_wide_scalars"`
	InspectableHasherState    bool      `json:"inspectable_hasher_state"`
	ExplicitIteratorDriver    bool      `json:"explicit_iterator_host_driver"`
	Vectors                   []vector  `json:"vectors"`
	Blockers                  []blocker `json:"blockers"`
	ProductionMigration       bool      `json:"production_migration_admitted"`
	NextLane                  string    `json:"next_lane"`
	Exclusions                []string  `json:"exclusions"`
}

type vector struct {
	Name   string `json:"name"`
	Detail string `json:"detail"`
}
type blocker struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
}

func main() {
	outPath := flag.String("out", "", "report output path")
	check := flag.Bool("check", false, "fail if report is stale")
	flag.Parse()
	if *outPath == "" {
		fatalf("-out is required")
	}
	result, err := build()
	if err != nil {
		fatalf("%v", err)
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fatalf("marshal: %v", err)
	}
	data = append(data, '\n')
	if *check {
		current, err := os.ReadFile(*outPath)
		if err != nil {
			fatalf("read %s: %v", *outPath, err)
		}
		if !bytes.Equal(current, data) {
			fatalf("%s is stale; regenerate the Go binding report", *outPath)
		}
		return
	}
	if err := os.WriteFile(*outPath, data, 0o644); err != nil {
		fatalf("write: %v", err)
	}
}

func build() (report, error) {
	result := report{
		Schema:           "able.semanticabi.heap-contract-reconciliation.v1",
		Decision:         "admit-bounded-production-shared-value-pilot",
		ManifestIdentity: hex.EncodeToString(semanticabi.ManifestIdentity[:]), RuntimeKinds: len(semanticabi.KindManifest),
		ExactKinds: 27, ConditionalKinds: 0, OpaqueHostKinds: 4, BlockedKinds: 0,
		PackageLayoutReconciled: true, TwoPhaseGraphConstruction: true,
		OwnedWideScalars: true, InspectableHasherState: true, ExplicitIteratorDriver: true,
		Blockers: []blocker{}, ProductionMigration: true, NextLane: "runtime-contract-performance-evidence-reconciliation",
		Exclusions: []string{"production runtime migration", "foreign heap", "cgo runtime", "JIT or backend", "executable memory", "benchmark branch", "named-container or non-primitive nominal special case", "WASM"},
	}
	shared := &runtime.ArrayValue{Elements: []runtime.Value{runtime.NewSmallInt(7, runtime.IntegerI32)}}
	root := &runtime.ArrayValue{}
	root.Elements = []runtime.Value{shared, shared, root}
	decoded, _, err := gobinding.RoundTrip(root)
	if err != nil {
		return report{}, err
	}
	array := decoded[0].(*runtime.ArrayValue)
	if array.Elements[0] != array.Elements[1] || array.Elements[2] != array {
		return report{}, fmt.Errorf("array identity vector failed")
	}
	result.Vectors = append(result.Vectors, vector{"array-cycle-alias", "shared child and self-cycle preserve identity"})

	for _, current := range []struct {
		name   string
		engine *interpreter.Interpreter
	}{{"treewalker", interpreter.New()}, {"bytecode", interpreter.NewBytecode()}} {
		module := ast.Mod([]ast.Statement{ast.Assign(ast.ID("shared"), ast.Arr(ast.Int(1))), ast.Arr(ast.ID("shared"), ast.ID("shared"))}, nil, nil)
		value, _, err := current.engine.EvaluateModule(module)
		if err != nil {
			return report{}, err
		}
		values, _, err := gobinding.RoundTrip(value)
		if err != nil {
			return report{}, err
		}
		alias := values[0].(*runtime.ArrayValue)
		if alias.Elements[0] != alias.Elements[1] {
			return report{}, fmt.Errorf("%s alias vector failed", current.name)
		}
		result.Vectors = append(result.Vectors, vector{current.name + "-alias", "ordinary execution graph round-trips with alias identity"})

		closureModule := ast.Mod([]ast.Statement{ast.Assign(ast.ID("captured"), ast.Int(41)), ast.Lam(nil, ast.ID("captured"))}, nil, nil)
		closureValue, _, err := current.engine.EvaluateModule(closureModule)
		if err != nil {
			return report{}, err
		}
		fn := closureValue.(*runtime.FunctionValue)
		fn.Closure.DefineWithoutMerge("self", fn)
		values, _, err = gobinding.RoundTrip(fn)
		if err != nil {
			return report{}, err
		}
		restored := values[0].(*runtime.FunctionValue)
		call, err := current.engine.CallFunction(restored, nil)
		if err != nil {
			return report{}, err
		}
		number, _ := call.(runtime.IntegerValue).ToInt64()
		if number != 41 {
			return report{}, fmt.Errorf("%s closure result=%d", current.name, number)
		}
		result.Vectors = append(result.Vectors, vector{current.name + "-recursive-closure", "captured binding and self-cycle restore; call returns 41"})
	}

	definition := &runtime.StructDefinitionValue{Node: ast.StructDef("Box", nil, ast.StructKindNamed, nil, nil, false), NamedFieldIndices: map[string]int{"value": 0}}
	instance := &runtime.StructInstanceValue{Definition: definition, Fields: map[string]runtime.Value{"value": runtime.NewSmallInt(5, runtime.IntegerI32)}}
	errValue := &runtime.ErrorValue{TypeName: ast.ID("Wrapped"), Payload: map[string]runtime.Value{"value": instance}, Message: "wrapped"}
	future := runtime.NewFuture()
	future.Resolve(instance)
	values, _, err := gobinding.RoundTrip(&runtime.PackageValue{Name: "example", Public: map[string]runtime.Value{"box": instance}}, errValue, future)
	if err != nil {
		return report{}, err
	}
	box := values[0].(*runtime.PackageValue).Public["box"]
	if values[1].(*runtime.ErrorValue).Payload["value"] != box {
		return report{}, fmt.Errorf("nominal identity vector failed")
	}
	futureResult, _, _ := values[2].(*runtime.FutureValue).Snapshot()
	if futureResult != box {
		return report{}, fmt.Errorf("future identity vector failed")
	}
	result.Vectors = append(result.Vectors, vector{"nominal-error-future", "package, Error, and Future retain one StructInstance identity"})

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
	values, snapshot, err := gobinding.RoundTrip(runtime.NewBigIntValue(wide, runtime.IntegerI128), hasher, iterator, retained)
	if err != nil {
		return report{}, err
	}
	gotWide := values[0].(runtime.IntegerValue)
	if gotWide.TypeSuffix != runtime.IntegerI128 || gotWide.BigInt().Cmp(wide) != 0 {
		return report{}, fmt.Errorf("wide integer vector failed")
	}
	result.Vectors = append(result.Vectors, vector{"owned-wide-integer", "negative i128 magnitude and suffix survive an owned immutable scalar"})
	if values[1].(*runtime.HasherValue).SemanticState() != hasher.SemanticState() {
		return report{}, fmt.Errorf("Hasher state vector failed")
	}
	result.Vectors = append(result.Vectors, vector{"hasher-state", "complete evolving semantic state restores exactly"})
	gotIterator := values[2].(*runtime.IteratorValue)
	driver, closed := gotIterator.HostDriverSnapshot()
	if closed || len(driver.Retained) != 1 || driver.Retained[0] != values[3] {
		return report{}, fmt.Errorf("Iterator retained-root vector failed")
	}
	next, done, err := gotIterator.Next()
	if err != nil || done {
		return report{}, fmt.Errorf("Iterator resume vector failed: (%v, %v, %v)", next, done, err)
	}
	number, _ := next.(runtime.IntegerValue).ToInt64()
	if number != 2 {
		return report{}, fmt.Errorf("Iterator resumed at %d, want 2", number)
	}
	gotIterator.Close()
	if !finalized {
		return report{}, fmt.Errorf("Iterator finalizer vector failed")
	}
	collection, err := snapshot.Heap.Collect()
	if err != nil || collection.Reachable == 0 {
		return report{}, fmt.Errorf("indirect/host-root collection failed: %+v: %v", collection, err)
	}
	result.Vectors = append(result.Vectors, vector{"iterator-host-driver", "driver resumes, finalizer survives, and retained semantic roots preserve identity and lifetime"})
	return result, nil
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "semanticabi-bindingreport: "+format+"\n", args...)
	os.Exit(1)
}
