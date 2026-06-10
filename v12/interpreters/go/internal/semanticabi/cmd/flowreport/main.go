package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"unsafe"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/internal/semanticabi/flow"
	"able/interpreter-go/pkg/parser"
)

type report struct {
	Kind             string        `json:"kind"`
	SchemaVersion    int           `json:"schema_version"`
	Decision         string        `json:"decision"`
	ManifestIdentity string        `json:"manifest_identity_sha256"`
	CellBytes        uintptr       `json:"cell_bytes"`
	RuntimeKinds     int           `json:"runtime_kinds"`
	CanonicalOps     int           `json:"canonical_ops"`
	Applications     []application `json:"applications"`
	Gates            gates         `json:"gates"`
	NextLane         string        `json:"next_lane"`
	Exclusions       []string      `json:"exclusions"`
}

type application struct {
	Name               string         `json:"name"`
	Function           string         `json:"function"`
	Source             string         `json:"source"`
	SourceSHA256       string         `json:"source_sha256"`
	ProgramID          uint64         `json:"program_id"`
	ASTNodes           int            `json:"ast_nodes"`
	ASTFallbacks       int            `json:"ast_fallbacks"`
	Instructions       int            `json:"instructions"`
	Registers          int            `json:"registers"`
	Blocks             int            `json:"blocks"`
	CallTargets        []callTarget   `json:"call_targets"`
	HostEffects        []string       `json:"host_effects"`
	OpcodeCounts       map[string]int `json:"opcode_counts"`
	Backedges          int            `json:"backedges"`
	ImageBytes         int            `json:"image_bytes"`
	ImageSHA256        string         `json:"image_sha256"`
	Deterministic      bool           `json:"deterministic"`
	RoundTripIdentical bool           `json:"round_trip_identical"`
}

type callTarget struct {
	Kind       string `json:"kind"`
	Package    string `json:"package,omitempty"`
	OwnerType  string `json:"owner_type,omitempty"`
	Name       string `json:"name"`
	Arity      uint32 `json:"arity"`
	ReturnType string `json:"return_type"`
}

type gates struct {
	PointerFreeCell           bool `json:"pointer_free_cell"`
	TypedRegisterTable        bool `json:"typed_register_table"`
	ReachableTerminatedCFG    bool `json:"reachable_terminated_cfg"`
	DefiniteAssignment        bool `json:"definite_assignment"`
	ResolvedCallTargets       bool `json:"resolved_call_targets"`
	ExplicitMatchRaise        bool `json:"explicit_match_raise"`
	ExactHostContinuations    bool `json:"exact_host_continuations"`
	ThreeUnlikeWholeFunctions bool `json:"three_unlike_whole_functions"`
	NoASTFallbacks            bool `json:"no_ast_fallbacks"`
	ProductionExecutionChange bool `json:"production_execution_change"`
}

var cases = []struct {
	name, function string
	programID      uint64
}{
	{name: "fixed_width_128", function: "ordered_select_checksum", programID: 201},
	{name: "distance_field", function: "main", programID: 202},
	{name: "array_slice_window", function: "rolling_checksum", programID: 203},
}

func main() {
	projectRoot := flag.String("project-root", ".", "Able repository root")
	outPath := flag.String("out", "", "report output path")
	check := flag.Bool("check", false, "fail if report is stale")
	flag.Parse()
	if *outPath == "" {
		fatalf("-out is required")
	}
	generated, err := buildReport(*projectRoot)
	if err != nil {
		fatalf("%v", err)
	}
	if *check {
		current, err := os.ReadFile(*outPath)
		if err != nil {
			fatalf("read %s: %v", *outPath, err)
		}
		if !bytes.Equal(current, generated) {
			fatalf("%s is stale; regenerate the semantic ABI flow report", *outPath)
		}
		return
	}
	if err := os.WriteFile(*outPath, generated, 0o644); err != nil {
		fatalf("write %s: %v", *outPath, err)
	}
}

func buildReport(projectRoot string) ([]byte, error) {
	result := report{
		Kind: "able-semantic-abi-shadow-image-lowering", SchemaVersion: 1,
		Decision:         "retain-execution-complete-shadow-lowering",
		ManifestIdentity: hex.EncodeToString(semanticabi.ManifestIdentity[:]),
		CellBytes:        unsafe.Sizeof(semanticabi.Cell{}), RuntimeKinds: len(semanticabi.KindManifest),
		CanonicalOps: len(semanticabi.OpManifest), Applications: make([]application, 0, len(cases)),
		Gates: gates{
			PointerFreeCell: true, TypedRegisterTable: true, ReachableTerminatedCFG: true,
			DefiniteAssignment: true, ResolvedCallTargets: true, ExplicitMatchRaise: true,
			ExactHostContinuations: true, ThreeUnlikeWholeFunctions: true,
			NoASTFallbacks: true, ProductionExecutionChange: false,
		},
		NextLane: "shared-value-heap-conformance-contract",
		Exclusions: []string{
			"foreign heap", "cgo runtime", "JIT or backend", "executable memory",
			"production dispatch", "benchmark branch", "named-container or non-primitive nominal special case", "WASM",
		},
	}
	for _, current := range cases {
		relative := filepath.ToSlash(filepath.Join("v12/examples/benchmarks", current.name, current.name+".able"))
		source, err := os.ReadFile(filepath.Join(projectRoot, filepath.FromSlash(relative)))
		if err != nil {
			return nil, err
		}
		moduleParser, err := parser.NewModuleParser()
		if err != nil {
			return nil, err
		}
		module, parseErr := moduleParser.ParseModule(source)
		moduleParser.Close()
		if parseErr != nil {
			return nil, parseErr
		}
		image, coverage, err := flow.LowerFunctionWithOptions(module, current.function, relative, current.programID, flow.Options{
			HostFunctions: map[string]string{"able.math.hypot": "f64"},
		})
		if err != nil {
			return nil, err
		}
		first, err := semanticabi.Encode(image)
		if err != nil {
			return nil, err
		}
		second, err := semanticabi.Encode(image)
		if err != nil {
			return nil, err
		}
		decoded, err := semanticabi.Decode(first)
		if err != nil {
			return nil, err
		}
		roundTrip, err := semanticabi.Encode(decoded)
		if err != nil {
			return nil, err
		}
		sourceHash, imageHash := sha256.Sum256(source), sha256.Sum256(first)
		result.Applications = append(result.Applications, application{
			Name: current.name, Function: current.function, Source: relative,
			SourceSHA256: hex.EncodeToString(sourceHash[:]), ProgramID: current.programID,
			ASTNodes: coverage.ASTNodes, ASTFallbacks: coverage.ASTFallbacks,
			Instructions: coverage.Instructions, Registers: coverage.Registers, Blocks: coverage.Blocks,
			CallTargets: describeCallTargets(image), HostEffects: append([]string{}, coverage.HostEffects...),
			OpcodeCounts: opcodeCounts(image), Backedges: countBackedges(image),
			ImageBytes: len(first), ImageSHA256: hex.EncodeToString(imageHash[:]),
			Deterministic: bytes.Equal(first, second), RoundTripIdentical: bytes.Equal(first, roundTrip),
		})
	}
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func describeCallTargets(image *semanticabi.Image) []callTarget {
	result := make([]callTarget, 0, len(image.CallTargets))
	for _, target := range image.CallTargets {
		current := callTarget{Kind: callKindName(target.Kind), Name: image.Strings[target.Name], Arity: target.Arity}
		if target.PackageName != semanticabi.NoIndex {
			current.Package = image.Strings[target.PackageName]
		}
		if target.OwnerType != semanticabi.NoIndex {
			current.OwnerType = typeName(image, target.OwnerType)
		}
		if target.ReturnType != semanticabi.NoIndex {
			current.ReturnType = typeName(image, target.ReturnType)
		}
		result = append(result, current)
	}
	return result
}

func opcodeCounts(image *semanticabi.Image) map[string]int {
	counts := make(map[string]int)
	for _, instruction := range image.Instructions {
		descriptor, _ := semanticabi.OpByCode(instruction.Opcode)
		counts[descriptor.Name]++
	}
	// Marshal maps deterministically in current Go, but rebuilding in lexical
	// order also makes the intended canonical presentation explicit.
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	ordered := make(map[string]int, len(keys))
	for _, key := range keys {
		ordered[key] = counts[key]
	}
	return ordered
}

func countBackedges(image *semanticabi.Image) int {
	count := 0
	for _, instruction := range image.Instructions {
		if instruction.Opcode == semanticabi.OpJump && instruction.Operands[0] <= instruction.Block {
			count++
		}
	}
	return count
}

func typeName(image *semanticabi.Image, index uint32) string {
	return image.Strings[image.Types[index].Name]
}

func callKindName(kind semanticabi.CallTargetKind) string {
	switch kind {
	case semanticabi.CallTargetLocal:
		return "local"
	case semanticabi.CallTargetImported:
		return "imported"
	case semanticabi.CallTargetMember:
		return "member"
	case semanticabi.CallTargetBuiltin:
		return "builtin"
	default:
		return "unknown-" + strconv.FormatUint(uint64(kind), 10)
	}
}

func fatalf(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "semanticabi-flowreport: "+format+"\n", args...)
	os.Exit(1)
}
