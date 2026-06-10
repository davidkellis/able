package interpreter

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

const bytecodeOpcodeAuditEnabledEnv = "ABLE_BYTECODE_OPCODE_AUDIT"

type bytecodeOpcodeAuditTarget struct {
	Name string
	Path string
}

type bytecodeOpcodeAuditOp struct {
	Name string `json:"name"`
	Op   bytecodeOp
}

type bytecodeOpcodeAuditBenchmarkRow struct {
	Name              string                           `json:"name"`
	Path              string                           `json:"path"`
	FunctionCount     int                              `json:"function_count"`
	InstructionCount  int                              `json:"instruction_count"`
	TrackedOpcodeHits map[string]int                   `json:"tracked_opcode_hits"`
	Functions         []bytecodeOpcodeAuditFunctionRow `json:"functions,omitempty"`
}

type bytecodeOpcodeAuditFunctionRow struct {
	Name              string         `json:"name"`
	SlotFrame         bool           `json:"slot_frame"`
	InstructionCount  int            `json:"instruction_count"`
	TrackedOpcodeHits map[string]int `json:"tracked_opcode_hits"`
	LoadNameHits      map[string]int `json:"load_name_hits,omitempty"`
}

type bytecodeOpcodeAuditSummary struct {
	GeneratedAtUTC string                            `json:"generated_at_utc"`
	Suite          string                            `json:"suite"`
	Benchmarks     []bytecodeOpcodeAuditBenchmarkRow `json:"benchmarks"`
	Totals         bytecodeOpcodeAuditTotals         `json:"totals"`
}

type bytecodeOpcodeAuditTotals struct {
	BenchmarkCount     int            `json:"benchmark_count"`
	FunctionCount      int            `json:"function_count"`
	InstructionCount   int            `json:"instruction_count"`
	TrackedOpcodeHits  map[string]int `json:"tracked_opcode_hits"`
	BenchmarksWithHits map[string]int `json:"benchmarks_with_hits"`
}

var bytecodeOpcodeAuditTrackedOps = []bytecodeOpcodeAuditOp{
	{Name: "LoadName", Op: bytecodeOpLoadName},
	{Name: "LoadSlot", Op: bytecodeOpLoadSlot},
	{Name: "LoadImplicitSlot", Op: bytecodeOpLoadImplicitSlot},
	{Name: "LoadSlotI32", Op: bytecodeOpLoadSlotI32},
	{Name: "StoreImplicitSlot", Op: bytecodeOpStoreImplicitSlot},
	{Name: "Match", Op: bytecodeOpMatch},
	{Name: "JumpIfNotTypedPattern", Op: bytecodeOpJumpIfNotTypedPattern},
	{Name: "LoadSlotStructField", Op: bytecodeOpLoadSlotStructField},
	{Name: "TryFloatUpdatePair", Op: bytecodeOpTryFloatUpdatePair},
	{Name: "JumpIfFloatMulAddMulCompareConstFalse", Op: bytecodeOpJumpIfFloatMulAddMulCompareConstFalse},
	{Name: "JumpIfFloatAddCompareConstFalse", Op: bytecodeOpJumpIfFloatAddCompareConstFalse},
	{Name: "StoreSlotFloatAddMulSlot", Op: bytecodeOpStoreSlotFloatAddMulSlot},
	{Name: "StoreSlotFloatAddSub", Op: bytecodeOpStoreSlotFloatAddSub},
}

func TestBytecodeBenchmarkOpcodeAudit(t *testing.T) {
	if strings.TrimSpace(os.Getenv(bytecodeOpcodeAuditEnabledEnv)) == "" {
		t.Skipf("set %s=1 to run benchmark opcode audit", bytecodeOpcodeAuditEnabledEnv)
	}

	targets, err := parseBytecodeOpcodeAuditTargets(os.Getenv("ABLE_BYTECODE_OPCODE_AUDIT_BENCHMARK_TARGETS"))
	if err != nil {
		t.Fatalf("parse benchmark targets: %v", err)
	}
	if len(targets) == 0 {
		t.Fatalf("no benchmark targets provided")
	}

	suite := strings.TrimSpace(os.Getenv("ABLE_BYTECODE_OPCODE_AUDIT_SUITE"))
	if suite == "" {
		suite = "unspecified"
	}

	rows := make([]bytecodeOpcodeAuditBenchmarkRow, 0, len(targets))
	totals := bytecodeOpcodeAuditTotals{
		BenchmarkCount:     len(targets),
		TrackedOpcodeHits:  make(map[string]int, len(bytecodeOpcodeAuditTrackedOps)),
		BenchmarksWithHits: make(map[string]int, len(bytecodeOpcodeAuditTrackedOps)),
	}
	for _, tracked := range bytecodeOpcodeAuditTrackedOps {
		totals.TrackedOpcodeHits[tracked.Name] = 0
		totals.BenchmarksWithHits[tracked.Name] = 0
	}

	for _, target := range targets {
		row, err := runBytecodeOpcodeAuditTarget(target)
		if err != nil {
			t.Fatalf("audit %s: %v", target.Name, err)
		}
		rows = append(rows, row)
		totals.FunctionCount += row.FunctionCount
		totals.InstructionCount += row.InstructionCount
		for _, tracked := range bytecodeOpcodeAuditTrackedOps {
			hits := row.TrackedOpcodeHits[tracked.Name]
			totals.TrackedOpcodeHits[tracked.Name] += hits
			if hits > 0 {
				totals.BenchmarksWithHits[tracked.Name]++
			}
		}
	}

	summary := bytecodeOpcodeAuditSummary{
		GeneratedAtUTC: time.Now().UTC().Format(time.RFC3339),
		Suite:          suite,
		Benchmarks:     rows,
		Totals:         totals,
	}

	t.Logf("bytecode opcode audit suite=%s benchmarks=%d functions=%d instructions=%d", summary.Suite, summary.Totals.BenchmarkCount, summary.Totals.FunctionCount, summary.Totals.InstructionCount)
	for _, row := range summary.Benchmarks {
		t.Logf("%s funcs=%d instructions=%d tracked=%s", row.Name, row.FunctionCount, row.InstructionCount, formatBytecodeOpcodeAuditCounts(row.TrackedOpcodeHits))
	}
	t.Logf("totals tracked=%s", formatBytecodeOpcodeAuditCounts(summary.Totals.TrackedOpcodeHits))
	t.Logf("benchmark coverage=%s", formatBytecodeOpcodeAuditCounts(summary.Totals.BenchmarksWithHits))

	if err := writeBytecodeOpcodeAuditOutputs(summary, os.Getenv("ABLE_BYTECODE_OPCODE_AUDIT_OUTPUT_JSON"), os.Getenv("ABLE_BYTECODE_OPCODE_AUDIT_OUTPUT_MD")); err != nil {
		t.Fatalf("write audit outputs: %v", err)
	}
}

func parseBytecodeOpcodeAuditTargets(raw string) ([]bytecodeOpcodeAuditTarget, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	targets := make([]bytecodeOpcodeAuditTarget, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, path, ok := strings.Cut(part, "=")
		if !ok {
			return nil, fmt.Errorf("invalid target %q", part)
		}
		name = strings.TrimSpace(name)
		path = strings.TrimSpace(path)
		if name == "" || path == "" {
			return nil, fmt.Errorf("invalid target %q", part)
		}
		targets = append(targets, bytecodeOpcodeAuditTarget{Name: name, Path: path})
	}
	return targets, nil
}

func runBytecodeOpcodeAuditTarget(target bytecodeOpcodeAuditTarget) (bytecodeOpcodeAuditBenchmarkRow, error) {
	module, err := parseSourceModule(target.Path)
	if err != nil {
		return bytecodeOpcodeAuditBenchmarkRow{}, err
	}
	interp := NewBytecode()
	env := interp.GlobalEnvironment()
	if err := seedBytecodeOpcodeAuditModule(interp, env, module); err != nil {
		return bytecodeOpcodeAuditBenchmarkRow{}, err
	}

	row := bytecodeOpcodeAuditBenchmarkRow{
		Name:              target.Name,
		Path:              target.Path,
		TrackedOpcodeHits: make(map[string]int, len(bytecodeOpcodeAuditTrackedOps)),
	}
	for _, tracked := range bytecodeOpcodeAuditTrackedOps {
		row.TrackedOpcodeHits[tracked.Name] = 0
	}

	functions := collectBytecodeOpcodeAuditFunctions(module)
	row.FunctionCount = len(functions)
	for _, def := range functions {
		program, err := interp.lowerFunctionDefinitionBytecodeWithEnv(def, env)
		if err != nil {
			fnName := "<anonymous>"
			if def != nil && def.ID != nil && def.ID.Name != "" {
				fnName = def.ID.Name
			}
			return bytecodeOpcodeAuditBenchmarkRow{}, fmt.Errorf("lower %s: %w", fnName, err)
		}
		functionRow := bytecodeOpcodeAuditFunctionRow{
			Name:              bytecodeOpcodeAuditFunctionName(def),
			SlotFrame:         program != nil && program.frameLayout != nil,
			TrackedOpcodeHits: make(map[string]int, len(bytecodeOpcodeAuditTrackedOps)),
			LoadNameHits:      make(map[string]int),
		}
		for _, tracked := range bytecodeOpcodeAuditTrackedOps {
			functionRow.TrackedOpcodeHits[tracked.Name] = 0
		}
		functionRow.InstructionCount = bytecodeOpcodeAuditCountProgram(program, functionRow.TrackedOpcodeHits, functionRow.LoadNameHits)
		if len(functionRow.LoadNameHits) == 0 {
			functionRow.LoadNameHits = nil
		}
		row.Functions = append(row.Functions, functionRow)
		row.InstructionCount += functionRow.InstructionCount
		for _, tracked := range bytecodeOpcodeAuditTrackedOps {
			row.TrackedOpcodeHits[tracked.Name] += functionRow.TrackedOpcodeHits[tracked.Name]
		}
	}

	return row, nil
}

func bytecodeOpcodeAuditFunctionName(def *ast.FunctionDefinition) string {
	if def == nil || def.ID == nil || def.ID.Name == "" {
		return "<anonymous>"
	}
	return def.ID.Name
}

func seedBytecodeOpcodeAuditModule(interp *Interpreter, env *runtime.Environment, module *ast.Module) error {
	if interp == nil || env == nil || module == nil {
		return nil
	}
	for _, stmt := range module.Body {
		switch def := stmt.(type) {
		case *ast.StructDefinition:
			if _, err := interp.evaluateStructDefinition(def, env); err != nil {
				return err
			}
		case *ast.UnionDefinition:
			if _, err := interp.evaluateUnionDefinition(def, env); err != nil {
				return err
			}
		case *ast.InterfaceDefinition:
			if _, err := interp.evaluateInterfaceDefinition(def, env); err != nil {
				return err
			}
		}
	}
	return nil
}

func collectBytecodeOpcodeAuditFunctions(module *ast.Module) []*ast.FunctionDefinition {
	if module == nil {
		return nil
	}
	functions := make([]*ast.FunctionDefinition, 0)
	for _, stmt := range module.Body {
		switch def := stmt.(type) {
		case *ast.FunctionDefinition:
			functions = append(functions, def)
		case *ast.MethodsDefinition:
			functions = append(functions, def.Definitions...)
		case *ast.ImplementationDefinition:
			functions = append(functions, def.Definitions...)
		}
	}
	return functions
}

func bytecodeOpcodeAuditCountProgram(program *bytecodeProgram, trackedHits map[string]int, loadNameHits map[string]int) int {
	if program == nil {
		return 0
	}
	trackedByOp := make(map[bytecodeOp]string, len(bytecodeOpcodeAuditTrackedOps))
	for _, tracked := range bytecodeOpcodeAuditTrackedOps {
		trackedByOp[tracked.Op] = tracked.Name
	}
	var count int
	var walk func(*bytecodeProgram)
	walk = func(program *bytecodeProgram) {
		if program == nil {
			return
		}
		for _, instr := range program.instructions {
			count++
			if name, ok := trackedByOp[instr.op]; ok {
				trackedHits[name]++
			}
			if instr.op == bytecodeOpLoadName && loadNameHits != nil {
				loadNameHits[instr.name]++
			}
			if instr.program != nil {
				walk(instr.program)
			}
		}
	}
	walk(program)
	return count
}

func formatBytecodeOpcodeAuditCounts(counts map[string]int) string {
	if len(counts) == 0 {
		return "none"
	}
	keys := make([]string, 0, len(counts))
	for key := range counts {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", key, counts[key]))
	}
	return strings.Join(parts, " ")
}

func writeBytecodeOpcodeAuditOutputs(summary bytecodeOpcodeAuditSummary, jsonPath string, mdPath string) error {
	if strings.TrimSpace(jsonPath) != "" {
		payload, err := json.MarshalIndent(summary, "", "  ")
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(jsonPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(jsonPath, append(payload, '\n'), 0o644); err != nil {
			return err
		}
	}
	if strings.TrimSpace(mdPath) != "" {
		if err := os.MkdirAll(filepath.Dir(mdPath), 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(mdPath, []byte(renderBytecodeOpcodeAuditMarkdown(summary)), 0o644); err != nil {
			return err
		}
	}
	return nil
}

func renderBytecodeOpcodeAuditMarkdown(summary bytecodeOpcodeAuditSummary) string {
	var b strings.Builder
	b.WriteString("# Bytecode Opcode Audit\n\n")
	b.WriteString(fmt.Sprintf("- Generated: `%s`\n", summary.GeneratedAtUTC))
	b.WriteString(fmt.Sprintf("- Suite: `%s`\n", summary.Suite))
	b.WriteString(fmt.Sprintf("- Benchmarks: `%d`\n", summary.Totals.BenchmarkCount))
	b.WriteString(fmt.Sprintf("- Functions lowered: `%d`\n", summary.Totals.FunctionCount))
	b.WriteString(fmt.Sprintf("- Instructions lowered: `%d`\n\n", summary.Totals.InstructionCount))

	b.WriteString("| Benchmark | Functions | Instructions |")
	for _, tracked := range bytecodeOpcodeAuditTrackedOps {
		b.WriteString(" `")
		b.WriteString(tracked.Name)
		b.WriteString("` |")
	}
	b.WriteString("\n| --- | ---: | ---: |")
	for range bytecodeOpcodeAuditTrackedOps {
		b.WriteString(" ---: |")
	}
	b.WriteString("\n")
	for _, row := range summary.Benchmarks {
		b.WriteString(fmt.Sprintf("| `%s` | %d | %d |", row.Name, row.FunctionCount, row.InstructionCount))
		for _, tracked := range bytecodeOpcodeAuditTrackedOps {
			b.WriteString(fmt.Sprintf(" %d |", row.TrackedOpcodeHits[tracked.Name]))
		}
		b.WriteString("\n")
	}

	b.WriteString("\n## Totals\n\n")
	b.WriteString("| Opcode | Total Hits | Benchmarks With Hits |\n")
	b.WriteString("| --- | ---: | ---: |\n")
	for _, tracked := range bytecodeOpcodeAuditTrackedOps {
		b.WriteString(fmt.Sprintf("| `%s` | %d | %d |\n", tracked.Name, summary.Totals.TrackedOpcodeHits[tracked.Name], summary.Totals.BenchmarksWithHits[tracked.Name]))
	}
	return b.String()
}
