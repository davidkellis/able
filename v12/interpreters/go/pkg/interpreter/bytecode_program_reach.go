package interpreter

import (
	"sort"
	"sync/atomic"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

// The reach registry is diagnostic-only and bounded so an opt-in census cannot
// retain an unbounded number of dynamically-created programs.
const bytecodeProgramReachLimit = 4096

type bytecodeProgramReach struct {
	kind                       string
	name                       string
	origin                     string
	line                       int
	column                     int
	staticInstructions         int
	staticPrimitiveEligible    int
	staticEffectBoundaries     int
	staticBackedges            int
	maxStaticPrimitiveSpan     int
	paramCount                 int
	boxedParamCount            int
	boxedReturn                bool
	requiresEnvironment        bool
	entries                    uint64
	dynamicInstructions        uint64
	dynamicPrimitiveEligible   uint64
	dynamicEffectBoundaries    uint64
	dynamicBackedges           uint64
	instructionScalarProofs    []bytecodeScalarProofKind
	integerLoadUses            []bytecodeIntegerLoadUse
	staticScalarProofCounts    [bytecodeScalarProofTargetCount][bytecodeScalarProofCount]int
	dynamicScalarProofCounts   [bytecodeScalarProofTargetCount][bytecodeScalarProofCount]uint64
	dynamicIntegerLoadCarriers [][bytecodeIntegerLoadCarrierCount]uint64
}

// BytecodeProgramReachSnapshot attributes opt-in VM work to a source program.
// PrimitiveEligible is deliberately conservative: it includes only operations
// whose semantics can be represented by the proposed pointer-free primitive
// native leaf tier without invoking general Able runtime dispatch.
type BytecodeProgramReachSnapshot struct {
	Kind                     string                             `json:"kind"`
	Name                     string                             `json:"name"`
	Origin                   string                             `json:"origin,omitempty"`
	Line                     int                                `json:"line,omitempty"`
	Column                   int                                `json:"column,omitempty"`
	Entries                  uint64                             `json:"entries"`
	DynamicInstructions      uint64                             `json:"dynamic_instructions"`
	DynamicPrimitiveEligible uint64                             `json:"dynamic_primitive_eligible"`
	DynamicEffectBoundaries  uint64                             `json:"dynamic_effect_boundaries"`
	DynamicBackedges         uint64                             `json:"dynamic_backedges"`
	StaticInstructions       int                                `json:"static_instructions"`
	StaticPrimitiveEligible  int                                `json:"static_primitive_eligible"`
	StaticEffectBoundaries   int                                `json:"static_effect_boundaries"`
	StaticBackedges          int                                `json:"static_backedges"`
	MaxStaticPrimitiveSpan   int                                `json:"max_static_primitive_span"`
	ParamCount               int                                `json:"param_count"`
	BoxedParamCount          int                                `json:"boxed_param_count"`
	BoxedReturn              bool                               `json:"boxed_return"`
	RequiresEnvironment      bool                               `json:"requires_environment"`
	ScalarProofs             []BytecodeScalarProofSnapshot      `json:"scalar_proofs,omitempty"`
	IntegerLoadShapes        []BytecodeIntegerLoadShapeSnapshot `json:"integer_load_shapes,omitempty"`
}

func (i *Interpreter) annotateBytecodeProgramReach(program *bytecodeProgram, kind, name string, node ast.Node) *bytecodeProgram {
	return i.annotateBytecodeProgramReachWithScalarChecks(program, kind, name, node, nil)
}

func (i *Interpreter) annotateBytecodeProgramReachWithScalarChecks(program *bytecodeProgram, kind, name string, node ast.Node, loweringChecks []bytecodeSimpleTypeCheck) *bytecodeProgram {
	if program == nil {
		return nil
	}
	if i == nil || !i.bytecodeStatsEnabled {
		return program
	}
	if name == "" {
		name = "<anonymous>"
	}
	reach := &bytecodeProgramReach{kind: kind, name: name}
	if node != nil {
		reach.origin = i.nodeOrigins[node]
		span := node.Span()
		reach.line = span.Start.Line
		reach.column = span.Start.Column
	}
	if layout := program.frameLayout; layout != nil {
		reach.paramCount = layout.paramSlots
		for _, simple := range layout.paramSimpleTypes {
			if !bytecodeProgramPrimitiveSimpleType(simple) {
				reach.boxedParamCount++
			}
		}
		reach.boxedReturn = layout.returnSimpleType != "" && layout.returnSimpleType != "void" && layout.returnSimpleType != "nil" && !bytecodeProgramPrimitiveSimpleType(layout.returnSimpleType)
		reach.requiresEnvironment = layout.needsEnvScopes
	} else if kind == "function" {
		reach.requiresEnvironment = true
	}
	reach.instructionScalarProofs = make([]bytecodeScalarProofKind, len(program.instructions))
	reach.integerLoadUses = make([]bytecodeIntegerLoadUse, len(program.instructions))
	reach.dynamicIntegerLoadCarriers = make([][bytecodeIntegerLoadCarrierCount]uint64, len(program.instructions))
	program.reach = reach
	inferred := i.bytecodeInferenceFactsSnapshot()
	primitiveSpan := 0
	for ip := range program.instructions {
		instr := &program.instructions[ip]
		loweringCheck := bytecodeSimpleTypeCheckUnknown
		if ip < len(loweringChecks) {
			loweringCheck = loweringChecks[ip]
		}
		proof := bytecodeScalarProofForInstruction(program, ip, instr, loweringCheck, inferred)
		reach.instructionScalarProofs[ip] = proof
		if instr.op == bytecodeOpLoadSlot && proof == bytecodeScalarProofSlotInteger {
			reach.integerLoadUses[ip] = bytecodeIntegerLoadConsumerForProgram(program, ip)
		}
		if target, ok := bytecodeScalarProofTargetForOp(instr.op); ok && proof > bytecodeScalarProofNotTarget && proof < bytecodeScalarProofCount {
			reach.staticScalarProofCounts[target][proof]++
		}
		reach.staticInstructions++
		if bytecodeProgramInstructionPrimitiveEligible(program, ip, instr) {
			reach.staticPrimitiveEligible++
			primitiveSpan++
			if primitiveSpan > reach.maxStaticPrimitiveSpan {
				reach.maxStaticPrimitiveSpan = primitiveSpan
			}
		} else {
			reach.staticEffectBoundaries++
			primitiveSpan = 0
		}
		if bytecodeProgramInstructionBackedge(ip, instr) {
			reach.staticBackedges++
			primitiveSpan = 0
		}
	}
	return program
}

func bytecodeProgramPrimitiveSimpleType(name string) bool {
	if name == "bool" || name == "char" {
		return true
	}
	if _, ok := integerTypes[name]; ok {
		return true
	}
	_, ok := floatTypes[name]
	return ok
}

func (i *Interpreter) recordBytecodeProgramEntry(program *bytecodeProgram) {
	if i == nil || !i.bytecodeStatsEnabled || program == nil || program.reach == nil {
		return
	}
	i.bytecodeProgramReachMu.Lock()
	if i.bytecodeProgramReach == nil {
		i.bytecodeProgramReach = make(map[*bytecodeProgram]struct{})
	}
	if _, ok := i.bytecodeProgramReach[program]; !ok {
		if len(i.bytecodeProgramReach) >= bytecodeProgramReachLimit {
			i.bytecodeProgramReachDropped++
			i.bytecodeProgramReachMu.Unlock()
			return
		}
		i.bytecodeProgramReach[program] = struct{}{}
	}
	i.bytecodeProgramReachMu.Unlock()
	atomic.AddUint64(&program.reach.entries, 1)
}

func (i *Interpreter) recordBytecodeProgramInstruction(program *bytecodeProgram, ip int, instr *bytecodeInstruction) {
	if i == nil || !i.bytecodeStatsEnabled || program == nil || program.reach == nil || instr == nil {
		return
	}
	atomic.AddUint64(&program.reach.dynamicInstructions, 1)
	if bytecodeProgramInstructionPrimitiveEligible(program, ip, instr) {
		atomic.AddUint64(&program.reach.dynamicPrimitiveEligible, 1)
	} else {
		atomic.AddUint64(&program.reach.dynamicEffectBoundaries, 1)
	}
	if bytecodeProgramInstructionBackedge(ip, instr) {
		atomic.AddUint64(&program.reach.dynamicBackedges, 1)
	}
	if ip >= 0 && ip < len(program.reach.instructionScalarProofs) {
		proof := program.reach.instructionScalarProofs[ip]
		if target, ok := bytecodeScalarProofTargetForOp(instr.op); ok && proof > bytecodeScalarProofNotTarget && proof < bytecodeScalarProofCount {
			atomic.AddUint64(&program.reach.dynamicScalarProofCounts[target][proof], 1)
		}
	}
}

func (i *Interpreter) bytecodeProgramReachSnapshot() ([]BytecodeProgramReachSnapshot, uint64) {
	if i == nil {
		return nil, 0
	}
	i.bytecodeProgramReachMu.Lock()
	programs := make([]*bytecodeProgram, 0, len(i.bytecodeProgramReach))
	for program := range i.bytecodeProgramReach {
		programs = append(programs, program)
	}
	dropped := i.bytecodeProgramReachDropped
	i.bytecodeProgramReachMu.Unlock()
	rows := make([]BytecodeProgramReachSnapshot, 0, len(programs))
	for _, program := range programs {
		r := program.reach
		if r == nil {
			continue
		}
		row := BytecodeProgramReachSnapshot{
			Kind: r.kind, Name: r.name, Origin: r.origin, Line: r.line, Column: r.column,
			Entries:                  atomic.LoadUint64(&r.entries),
			DynamicInstructions:      atomic.LoadUint64(&r.dynamicInstructions),
			DynamicPrimitiveEligible: atomic.LoadUint64(&r.dynamicPrimitiveEligible),
			DynamicEffectBoundaries:  atomic.LoadUint64(&r.dynamicEffectBoundaries),
			DynamicBackedges:         atomic.LoadUint64(&r.dynamicBackedges),
			StaticInstructions:       r.staticInstructions, StaticPrimitiveEligible: r.staticPrimitiveEligible,
			StaticEffectBoundaries: r.staticEffectBoundaries, StaticBackedges: r.staticBackedges,
			MaxStaticPrimitiveSpan: r.maxStaticPrimitiveSpan, ParamCount: r.paramCount,
			BoxedParamCount: r.boxedParamCount, BoxedReturn: r.boxedReturn,
			RequiresEnvironment: r.requiresEnvironment,
		}
		for target := bytecodeScalarProofTarget(0); target < bytecodeScalarProofTargetCount; target++ {
			for proof := bytecodeScalarProofUnproven; proof < bytecodeScalarProofCount; proof++ {
				staticCount := r.staticScalarProofCounts[target][proof]
				dynamicCount := atomic.LoadUint64(&r.dynamicScalarProofCounts[target][proof])
				if staticCount == 0 && dynamicCount == 0 {
					continue
				}
				row.ScalarProofs = append(row.ScalarProofs, BytecodeScalarProofSnapshot{
					Opcode: target.String(), Proof: proof.String(), StaticInstructions: staticCount, DynamicInstructions: dynamicCount,
				})
			}
		}
		type integerLoadShapeKey struct {
			carrier   bytecodeIntegerLoadCarrier
			consumer  bytecodeIntegerLoadConsumer
			op        bytecodeOp
			operation string
			role      bytecodeIntegerLoadOperandRole
		}
		shapeCounts := make(map[integerLoadShapeKey]uint64)
		for ip, carriers := range r.dynamicIntegerLoadCarriers {
			use := bytecodeIntegerLoadUse{}
			if ip < len(r.integerLoadUses) {
				use = r.integerLoadUses[ip]
			}
			for carrier := bytecodeIntegerLoadCarrier(0); carrier < bytecodeIntegerLoadCarrierCount; carrier++ {
				count := atomic.LoadUint64(&carriers[carrier])
				if count != 0 {
					shapeCounts[integerLoadShapeKey{carrier: carrier, consumer: use.consumer, op: use.op, operation: use.operation, role: use.role}] += count
				}
			}
		}
		for key, count := range shapeCounts {
			opcode := ""
			if key.consumer != bytecodeIntegerLoadConsumerUnknown {
				opcode = bytecodeOpName(key.op)
			}
			row.IntegerLoadShapes = append(row.IntegerLoadShapes, BytecodeIntegerLoadShapeSnapshot{
				Carrier: key.carrier.String(), Consumer: key.consumer.String(), ConsumerOpcode: opcode,
				ConsumerOperation: key.operation, ConsumerOperandRole: key.role.String(), DynamicInstructions: count,
			})
		}
		sort.Slice(row.IntegerLoadShapes, func(left, right int) bool {
			if row.IntegerLoadShapes[left].DynamicInstructions != row.IntegerLoadShapes[right].DynamicInstructions {
				return row.IntegerLoadShapes[left].DynamicInstructions > row.IntegerLoadShapes[right].DynamicInstructions
			}
			if row.IntegerLoadShapes[left].Carrier != row.IntegerLoadShapes[right].Carrier {
				return row.IntegerLoadShapes[left].Carrier < row.IntegerLoadShapes[right].Carrier
			}
			if row.IntegerLoadShapes[left].ConsumerOpcode != row.IntegerLoadShapes[right].ConsumerOpcode {
				return row.IntegerLoadShapes[left].ConsumerOpcode < row.IntegerLoadShapes[right].ConsumerOpcode
			}
			if row.IntegerLoadShapes[left].ConsumerOperation != row.IntegerLoadShapes[right].ConsumerOperation {
				return row.IntegerLoadShapes[left].ConsumerOperation < row.IntegerLoadShapes[right].ConsumerOperation
			}
			return row.IntegerLoadShapes[left].ConsumerOperandRole < row.IntegerLoadShapes[right].ConsumerOperandRole
		})
		rows = append(rows, row)
	}
	sort.Slice(rows, func(left, right int) bool {
		if rows[left].DynamicInstructions != rows[right].DynamicInstructions {
			return rows[left].DynamicInstructions > rows[right].DynamicInstructions
		}
		if rows[left].Origin != rows[right].Origin {
			return rows[left].Origin < rows[right].Origin
		}
		if rows[left].Line != rows[right].Line {
			return rows[left].Line < rows[right].Line
		}
		return rows[left].Name < rows[right].Name
	})
	return rows, dropped
}

func (i *Interpreter) resetBytecodeProgramReach() {
	if i == nil {
		return
	}
	i.bytecodeProgramReachMu.Lock()
	for program := range i.bytecodeProgramReach {
		if r := program.reach; r != nil {
			atomic.StoreUint64(&r.entries, 0)
			atomic.StoreUint64(&r.dynamicInstructions, 0)
			atomic.StoreUint64(&r.dynamicPrimitiveEligible, 0)
			atomic.StoreUint64(&r.dynamicEffectBoundaries, 0)
			atomic.StoreUint64(&r.dynamicBackedges, 0)
			for target := bytecodeScalarProofTarget(0); target < bytecodeScalarProofTargetCount; target++ {
				for proof := bytecodeScalarProofUnproven; proof < bytecodeScalarProofCount; proof++ {
					atomic.StoreUint64(&r.dynamicScalarProofCounts[target][proof], 0)
				}
			}
			for ip := range r.dynamicIntegerLoadCarriers {
				for carrier := bytecodeIntegerLoadCarrier(0); carrier < bytecodeIntegerLoadCarrierCount; carrier++ {
					atomic.StoreUint64(&r.dynamicIntegerLoadCarriers[ip][carrier], 0)
				}
			}
		}
	}
	clear(i.bytecodeProgramReach)
	i.bytecodeProgramReachDropped = 0
	i.bytecodeProgramReachMu.Unlock()
}

func bytecodeProgramInstructionPrimitiveEligible(program *bytecodeProgram, ip int, instr *bytecodeInstruction) bool {
	if instr == nil {
		return false
	}
	if program != nil && program.reach != nil && ip >= 0 && ip < len(program.reach.instructionScalarProofs) && program.reach.instructionScalarProofs[ip].primitiveEligible() {
		return true
	}
	switch instr.op {
	case bytecodeOpConst:
		switch instr.value.(type) {
		case runtime.BoolValue, runtime.CharValue, runtime.IntegerValue, runtime.FloatValue:
			return true
		}
		return false
	case bytecodeOpLoadSlot, bytecodeOpLoadImplicitSlot, bytecodeOpStoreSlot,
		bytecodeOpStoreSlotNew, bytecodeOpStoreImplicitSlot:
		return bytecodeProgramSlotHasPrimitiveType(program, instr.target)
	case bytecodeOpDup, bytecodeOpPop,
		bytecodeOpBinaryIntAdd, bytecodeOpBinaryIntSub, bytecodeOpBinaryIntLessEqual,
		bytecodeOpBinaryIntDivCast, bytecodeOpBinaryCastSlotFloatConstDiv,
		bytecodeOpBinaryFloatMulSlotConst, bytecodeOpBinaryIntAddSlotConst,
		bytecodeOpBinaryIntSubSlotConst, bytecodeOpBinaryIntMulSlotConst,
		bytecodeOpBinaryIntModSlotConst, bytecodeOpBinaryIntLessEqualSlotConst,
		bytecodeOpBinaryIntCompareSlotConst, bytecodeOpJump,
		bytecodeOpJumpIfBoolSlotFalse, bytecodeOpJumpIfIntLessEqualSlotConstFalse,
		bytecodeOpJumpIfIntCompareSlotConstFalse, bytecodeOpJumpIfFloatMulAddMulCompareConstFalse,
		bytecodeOpJumpIfFloatAddCompareConstFalse, bytecodeOpJumpIfIntCompareSlotFalse,
		bytecodeOpReturnIfIntLessEqualSlotConst, bytecodeOpReturnConstIfIntLessEqualSlotConst,
		bytecodeOpConstI32, bytecodeOpBinaryI32Add, bytecodeOpBinaryI32Sub,
		bytecodeOpLoadSlotI32, bytecodeOpStoreSlotI32, bytecodeOpCompoundAssignSlotI32,
		bytecodeOpStoreSlotCastSlotFloatConstDiv, bytecodeOpStoreSlotFloatAffine,
		bytecodeOpStoreSlotFloatRegion, bytecodeOpStoreSlotBinaryIntSlotConst,
		bytecodeOpStoreSlotIntMulConstAdd, bytecodeOpStoreSlotIntMulConstModConst,
		bytecodeOpStoreSlotIntMulConstAddFromSlot, bytecodeOpStoreSlotFloatBinary,
		bytecodeOpStoreSlotFloatAddSub, bytecodeOpStoreSlotFloatAddMul,
		bytecodeOpStoreSlotFloatAddMulSlot:
		return true
	default:
		return false
	}
}

func bytecodeProgramSlotHasPrimitiveType(program *bytecodeProgram, slot int) bool {
	if program == nil || program.frameLayout == nil || slot < 0 {
		return false
	}
	layout := program.frameLayout
	return slot < len(layout.slotKinds) && layout.slotKinds[slot] != bytecodeCellKindValue
}

func bytecodeProgramInstructionBackedge(ip int, instr *bytecodeInstruction) bool {
	if instr == nil || instr.target < 0 || instr.target > ip {
		return false
	}
	switch instr.op {
	case bytecodeOpJump, bytecodeOpJumpIfFalse, bytecodeOpJumpIfBoolSlotFalse,
		bytecodeOpJumpIfIntLessEqualSlotConstFalse, bytecodeOpJumpIfIntCompareSlotConstFalse,
		bytecodeOpJumpIfArrayReadSlotCompareSlotFalse, bytecodeOpJumpIfArrayIndexSlotCompareSlotFalse,
		bytecodeOpJumpIfFloatMulAddMulCompareConstFalse, bytecodeOpJumpIfFloatAddCompareConstFalse,
		bytecodeOpJumpIfIntCompareSlotFalse, bytecodeOpJumpIfNil, bytecodeOpJumpIfNotNil,
		bytecodeOpJumpIfNotTypedPattern, bytecodeOpJumpIfBinaryCompareFalse:
		return true
	default:
		return false
	}
}
