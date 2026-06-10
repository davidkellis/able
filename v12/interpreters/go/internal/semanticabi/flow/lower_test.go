package flow

import (
	"bytes"
	"os"
	"path/filepath"
	goruntime "runtime"
	"testing"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/pkg/parser"
)

func TestRepresentativeFunctionsLowerToValidatedFlowImages(t *testing.T) {
	tests := []struct {
		application string
		function    string
		programID   uint64
		minBlocks   int
		minCalls    int
		hostEffects int
	}{
		{application: "fixed_width_128", function: "ordered_select_checksum", programID: 201, minBlocks: 8, minCalls: 3},
		{application: "distance_field", function: "main", programID: 202, minBlocks: 12, minCalls: 0, hostEffects: 2},
		{application: "array_slice_window", function: "rolling_checksum", programID: 203, minBlocks: 14, minCalls: 3},
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
			image, coverage, err := LowerFunctionWithOptions(module, test.function, filepath.ToSlash(path), test.programID, Options{
				HostFunctions: map[string]string{"able.math.hypot": "f64"},
			})
			if err != nil {
				t.Fatalf("LowerFunction: %v; coverage=%#v", err, coverage)
			}
			if !coverage.Complete() || coverage.ASTFallbacks != 0 {
				t.Fatalf("coverage is incomplete: %#v", coverage)
			}
			if coverage.Blocks < test.minBlocks || coverage.CallTargets < test.minCalls || len(coverage.HostEffects) != test.hostEffects {
				t.Fatalf("coverage = %#v, want blocks >= %d calls >= %d host effects = %d", coverage, test.minBlocks, test.minCalls, test.hostEffects)
			}
			if len(image.Functions) != 1 || image.Functions[0].Flags&semanticabi.FunctionFlagFlowValidated == 0 {
				t.Fatalf("function is not flow validated: %#v", image.Functions)
			}
			assertOnlyFlowOpcodes(t, image)
			assertApplicationContracts(t, test.application, image)
			first, err := semanticabi.Encode(image)
			if err != nil {
				t.Fatal(err)
			}
			second, err := semanticabi.Encode(image)
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := semanticabi.Decode(first)
			if err != nil {
				t.Fatal(err)
			}
			roundTrip, err := semanticabi.Encode(decoded)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(first, second) || !bytes.Equal(first, roundTrip) {
				t.Fatal("flow image is not byte-identical across encode/round trip")
			}
			t.Logf("ast=%d instructions=%d registers=%d blocks=%d calls=%d effects=%v bytes=%d", coverage.ASTNodes, coverage.Instructions, coverage.Registers, coverage.Blocks, coverage.CallTargets, coverage.HostEffects, len(first))
		})
	}
}

func assertApplicationContracts(t *testing.T, application string, image *semanticabi.Image) {
	t.Helper()
	hasBackedge, hasBranch := false, false
	opcodes := make(map[uint16]int)
	for _, instruction := range image.Instructions {
		opcodes[instruction.Opcode]++
		if instruction.Opcode == semanticabi.OpBranch {
			hasBranch = true
		}
		if instruction.Opcode == semanticabi.OpJump && instruction.Operands[0] <= instruction.Block {
			hasBackedge = true
		}
	}
	if !hasBackedge || !hasBranch {
		t.Fatalf("graph lacks backedge=%v or branch=%v", hasBackedge, hasBranch)
	}
	targets := make(map[string]semanticabi.CallTargetKind)
	for _, target := range image.CallTargets {
		targets[image.Strings[target.Name]] = target.Kind
	}
	switch application {
	case "fixed_width_128":
		if targets["new"] != semanticabi.CallTargetImported || targets["zero"] != semanticabi.CallTargetImported || targets["words_less"] != semanticabi.CallTargetLocal {
			t.Fatalf("fixed-width targets = %#v", targets)
		}
		for _, target := range image.CallTargets {
			if target.Kind == semanticabi.CallTargetImported && typeNameAt(image, target.ReturnType) != "dynamic" {
				t.Fatalf("imported nominal target %s inferred special return type %s", image.Strings[target.Name], typeNameAt(image, target.ReturnType))
			}
		}
	case "distance_field":
		if opcodes[semanticabi.OpHostEffectResume] != 2 {
			t.Fatalf("distance host-effect resumes = %d, want 2", opcodes[semanticabi.OpHostEffectResume])
		}
	case "array_slice_window":
		for _, name := range []string{"len", "slice", "read_slot"} {
			if targets[name] != semanticabi.CallTargetMember {
				t.Fatalf("array target %s = %d, want member; all=%#v", name, targets[name], targets)
			}
		}
		for _, target := range image.CallTargets {
			if target.Kind == semanticabi.CallTargetMember && typeNameAt(image, target.ReturnType) != "dynamic" {
				t.Fatalf("member target %s inferred nominal-specific return type %s", image.Strings[target.Name], typeNameAt(image, target.ReturnType))
			}
		}
		if opcodes[semanticabi.OpTypeTest] != 2 || opcodes[semanticabi.OpMatchFail] != 1 || opcodes[semanticabi.OpRaiseValue] != 1 {
			t.Fatalf("array match ops: type-test=%d match-fail=%d raise=%d", opcodes[semanticabi.OpTypeTest], opcodes[semanticabi.OpMatchFail], opcodes[semanticabi.OpRaiseValue])
		}
	}
}

func typeNameAt(image *semanticabi.Image, index uint32) string {
	return image.Strings[image.Types[index].Name]
}

func assertOnlyFlowOpcodes(t *testing.T, image *semanticabi.Image) {
	t.Helper()
	for index, instruction := range image.Instructions {
		if instruction.Opcode < semanticabi.OpLoadConst || instruction.Opcode > semanticabi.OpHostEffectResume {
			t.Fatalf("instruction %d uses structural opcode %d", index, instruction.Opcode)
		}
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
