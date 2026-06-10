package shadow

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	goruntime "runtime"
	"strings"
	"testing"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/parser"
)

func TestRepresentativeWholeFunctionsShadowEncodeWithoutFallback(t *testing.T) {
	tests := []struct {
		application string
		function    string
		programID   uint64
		minNodes    int
		hostEffects int
	}{
		{application: "fixed_width_128", function: "ordered_select_checksum", programID: 101, minNodes: 40},
		{application: "distance_field", function: "main", programID: 102, minNodes: 70, hostEffects: 2},
		{application: "array_slice_window", function: "rolling_checksum", programID: 103, minNodes: 80},
	}
	for _, test := range tests {
		t.Run(test.application, func(t *testing.T) {
			path := benchmarkPath(t, test.application)
			source, err := os.ReadFile(path)
			if err != nil {
				t.Fatal(err)
			}
			moduleParser, err := parser.NewModuleParser()
			if err != nil {
				t.Fatal(err)
			}
			defer moduleParser.Close()
			module, err := moduleParser.ParseModule(source)
			if err != nil {
				t.Fatal(err)
			}
			image, coverage, err := LowerFunction(module, test.function, path, test.programID)
			if err != nil {
				t.Fatalf("LowerFunction: %v; coverage = %#v", err, coverage)
			}
			if !coverage.Complete() || coverage.ASTFallbacks != 0 || len(coverage.Unsupported) != 0 {
				t.Fatalf("incomplete coverage: %#v", coverage)
			}
			if coverage.VisitedNodes < test.minNodes {
				t.Fatalf("visited %d nodes, want at least %d", coverage.VisitedNodes, test.minNodes)
			}
			if len(coverage.HostEffects) != test.hostEffects {
				t.Fatalf("host effects = %v, want %d", coverage.HostEffects, test.hostEffects)
			}
			if len(image.Functions) != 1 || image.Functions[0].Flags&semanticabi.FunctionFlagShadowEligible == 0 {
				t.Fatalf("function is not marked shadow eligible: %#v", image.Functions)
			}
			if len(image.Instructions) != coverage.LoweredNodes {
				t.Fatalf("instructions = %d, lowered nodes = %d", len(image.Instructions), coverage.LoweredNodes)
			}
			callsiteCount := 0
			for _, sourceRecord := range image.Sources {
				if sourceRecord.Callsite != semanticabi.NoIndex {
					callsiteCount++
				}
			}
			if callsiteCount == 0 {
				t.Fatal("image contains no callsite identity")
			}
			first, err := semanticabi.Encode(image)
			if err != nil {
				t.Fatal(err)
			}
			second, err := semanticabi.Encode(image)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(first, second) {
				t.Fatal("repeated shadow encodes differ")
			}
			decoded, err := semanticabi.Decode(first)
			if err != nil {
				t.Fatal(err)
			}
			reencoded, err := semanticabi.Encode(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(first, reencoded) {
				t.Fatal("shadow image decode/re-encode differs")
			}
			t.Logf("nodes=%d image_bytes=%d strings=%d types=%d sources=%d constants=%d effects=%v", coverage.LoweredNodes, len(first), len(image.Strings), len(image.Types), len(image.Sources), len(image.Constants), coverage.HostEffects)
		})
	}
}

func TestShadowImageTypesCannotRetainASTNodes(t *testing.T) {
	astNode := reflect.TypeOf((*ast.Node)(nil)).Elem()
	seen := make(map[reflect.Type]bool)
	var visit func(reflect.Type)
	visit = func(current reflect.Type) {
		if current == nil || seen[current] {
			return
		}
		seen[current] = true
		if current.Implements(astNode) || (current.Kind() != reflect.Pointer && reflect.PointerTo(current).Implements(astNode)) {
			t.Fatalf("semantic image type retains AST node type %s", current)
		}
		switch current.Kind() {
		case reflect.Pointer, reflect.Slice, reflect.Array:
			visit(current.Elem())
		case reflect.Struct:
			for index := 0; index < current.NumField(); index++ {
				visit(current.Field(index).Type)
			}
		}
	}
	visit(reflect.TypeOf(semanticabi.Image{}))
}

func TestShadowLoweringRejectsUnknownFunctionDeterministically(t *testing.T) {
	module := ast.NewModule(nil, nil, nil)
	_, _, first := LowerFunction(module, "missing", "missing.able", 1)
	_, _, second := LowerFunction(module, "missing", "missing.able", 1)
	if first == nil || second == nil || first.Error() != second.Error() || !strings.Contains(first.Error(), `function "missing" not found`) {
		t.Fatalf("errors = %v and %v", first, second)
	}
}

func benchmarkPath(t *testing.T, application string) string {
	t.Helper()
	_, filename, _, ok := goruntime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "../../../../../examples/benchmarks", application, application+".able"))
}
