package interpreter

import (
	"sort"
	"sync/atomic"

	"able/interpreter-go/pkg/runtime"
)

const bytecodePrimitiveMaterializationSiteLimit = 8192
const bytecodePrimitiveMaterializationStatsEnv = "ABLE_BYTECODE_PRIMITIVE_MATERIALIZATION_STATS"

const (
	bytecodeMaterializationCandidateStatic = "candidate_static"
	bytecodeMaterializationRequiredDynamic = "required_dynamic"
)

const (
	bytecodeMaterializationReasonStaticCall       = "static_call"
	bytecodeMaterializationReasonStaticReturn     = "static_return"
	bytecodeMaterializationReasonCast             = "cast"
	bytecodeMaterializationReasonPattern          = "pattern"
	bytecodeMaterializationReasonCollection       = "collection_value"
	bytecodeMaterializationReasonDynamicOperation = "dynamic_operation"
	bytecodeMaterializationReasonInterfaceUnion   = "interface_union"
	bytecodeMaterializationReasonHostNative       = "host_native"
	bytecodeMaterializationReasonEnvironment      = "environment"
	bytecodeMaterializationReasonErrorControl     = "error_control"
	bytecodeMaterializationReasonPublicEscape     = "public_escape"
)

type bytecodePrimitiveMaterializationKey struct {
	Class           string
	Reason          string
	Carrier         string
	Suffix          string
	Op              int
	IP              int
	Program         string
	Origin          string
	Line            int
	Column          int
	ReturnFrame     string
	ConsumerOp      int
	ConsumerIP      int
	ConsumerProgram string
	ConsumerOrigin  string
	ConsumerLine    int
	ConsumerColumn  int
}

// BytecodePrimitiveMaterializationSnapshot describes one raw primitive
// carrier-to-runtime-value transition observed during opt-in bytecode
// diagnostics.
type BytecodePrimitiveMaterializationSnapshot struct {
	Class           string `json:"class"`
	Reason          string `json:"reason"`
	Carrier         string `json:"carrier"`
	Suffix          string `json:"suffix"`
	Op              int    `json:"op"`
	Opcode          string `json:"opcode"`
	IP              int    `json:"ip"`
	Program         string `json:"program,omitempty"`
	Origin          string `json:"origin,omitempty"`
	Line            int    `json:"line,omitempty"`
	Column          int    `json:"column,omitempty"`
	ReturnFrame     string `json:"return_frame,omitempty"`
	ConsumerOp      int    `json:"consumer_op,omitempty"`
	ConsumerOpcode  string `json:"consumer_opcode,omitempty"`
	ConsumerIP      int    `json:"consumer_ip,omitempty"`
	ConsumerProgram string `json:"consumer_program,omitempty"`
	ConsumerOrigin  string `json:"consumer_origin,omitempty"`
	ConsumerLine    int    `json:"consumer_line,omitempty"`
	ConsumerColumn  int    `json:"consumer_column,omitempty"`
	Count           uint64 `json:"count"`
}

func bytecodeProgramDiagnosticName(program *bytecodeProgram) string {
	if program == nil || program.reach == nil {
		return ""
	}
	return program.reach.kind + ":" + program.reach.name
}

func (vm *bytecodeVM) bytecodeReturnConsumer() (frameKind string, program *bytecodeProgram, ip int) {
	if vm == nil {
		return "", nil, -1
	}
	if vm.selfFastMinimalSuffix > 0 && len(vm.selfFastMinimal) > 0 {
		frame := &vm.selfFastMinimal[len(vm.selfFastMinimal)-1]
		return "self_fast_minimal", vm.currentProgram, frame.returnIP
	}
	if len(vm.callFrameKinds) == 0 {
		return "root", nil, -1
	}
	switch vm.callFrameKinds[len(vm.callFrameKinds)-1] {
	case bytecodeCallFrameKindSelfFastMinimal:
		if len(vm.selfFastMinimal) > 0 {
			frame := &vm.selfFastMinimal[len(vm.selfFastMinimal)-1]
			return "self_fast_minimal", vm.currentProgram, frame.returnIP
		}
	case bytecodeCallFrameKindSelfFast:
		if len(vm.selfFastCallFrames) > 0 {
			frame := &vm.selfFastCallFrames[len(vm.selfFastCallFrames)-1]
			return "self_fast", vm.currentProgram, frame.returnIP
		}
	case bytecodeCallFrameKindFull:
		if len(vm.callFrames) > 0 {
			frame := &vm.callFrames[len(vm.callFrames)-1]
			return "full", frame.program, frame.returnIP
		}
	}
	return "invalid", nil, -1
}

func bytecodeRawPrimitiveCarrier(value runtime.Value) (carrier, suffix string, ok bool) {
	switch raw := value.(type) {
	case bytecodeRawI32SlotValue:
		return "i32_slot_value", string(runtime.IntegerI32), true
	case *bytecodeRawI32StackCell:
		if raw != nil {
			return "i32_stack_cell", string(runtime.IntegerI32), true
		}
	case bytecodeRawU8ResultValue:
		return "integer_result", string(runtime.IntegerU8), true
	case bytecodeRawU16ResultValue:
		return "integer_result", string(runtime.IntegerU16), true
	case bytecodeRawU32ResultValue:
		return "integer_result", string(runtime.IntegerU32), true
	case bytecodeRawU64ResultValue:
		return "integer_result", string(runtime.IntegerU64), true
	case bytecodeRawUsizeResultValue:
		return "integer_result", string(runtime.IntegerUsize), true
	case bytecodeRawI64ResultValue:
		return "integer_result", string(runtime.IntegerI64), true
	case bytecodeRawIntegerValue:
		return "integer_value", string(raw.TypeSuffix), true
	case *bytecodeRawIntegerSlotCell:
		if raw != nil {
			return "integer_slot_cell", string(raw.TypeSuffix), true
		}
	case *bytecodeRawIntegerReturnScratch:
		if raw != nil {
			return "integer_return_scratch", string(raw.TypeSuffix), true
		}
	case *bytecodeRawI64SlotCell:
		if raw != nil {
			return "i64_slot_cell", string(runtime.IntegerI64), true
		}
	case bytecodeRawF32SlotValue:
		return "float_slot_value", string(runtime.FloatF32), true
	case bytecodeRawF64SlotValue:
		return "float_slot_value", string(runtime.FloatF64), true
	}
	return "", "", false
}

func (vm *bytecodeVM) materializePrimitiveValue(class, reason string, value runtime.Value) runtime.Value {
	if vm != nil && vm.interp != nil && vm.interp.bytecodePrimitiveMaterializationStatsEnabled {
		vm.interp.recordBytecodePrimitiveMaterialization(vm, class, reason, value)
	}
	return bytecodeMaterializeRawValue(value)
}

func (vm *bytecodeVM) recordPrimitiveMaterializationValues(class, reason string, values []runtime.Value) {
	if vm == nil || vm.interp == nil || !vm.interp.bytecodePrimitiveMaterializationStatsEnabled {
		return
	}
	for _, value := range values {
		vm.interp.recordBytecodePrimitiveMaterialization(vm, class, reason, value)
	}
}

func (i *Interpreter) recordBytecodePrimitiveMaterialization(vm *bytecodeVM, class, reason string, value runtime.Value) {
	carrier, suffix, ok := bytecodeRawPrimitiveCarrier(value)
	if i == nil || !i.bytecodePrimitiveMaterializationStatsEnabled || !ok {
		return
	}
	key := bytecodePrimitiveMaterializationKey{
		Class: class, Reason: reason, Carrier: carrier, Suffix: suffix, Op: -1, IP: -1,
		ConsumerOp: -1, ConsumerIP: -1,
	}
	if vm != nil {
		key.IP = vm.ip
		if program := vm.currentProgram; program != nil {
			key.Program = bytecodeProgramDiagnosticName(program)
			if vm.ip >= 0 && vm.ip < len(program.instructions) {
				instr := &program.instructions[vm.ip]
				key.Op = int(instr.op)
				if instr.node != nil {
					key.Origin = i.nodeOrigins[instr.node]
					span := instr.node.Span()
					key.Line = span.Start.Line
					key.Column = span.Start.Column
				}
			}
		}
		if reason == bytecodeMaterializationReasonStaticReturn {
			var consumerProgram *bytecodeProgram
			key.ReturnFrame, consumerProgram, key.ConsumerIP = vm.bytecodeReturnConsumer()
			key.ConsumerProgram = bytecodeProgramDiagnosticName(consumerProgram)
			if consumerProgram != nil && key.ConsumerIP >= 0 && key.ConsumerIP < len(consumerProgram.instructions) {
				instr := &consumerProgram.instructions[key.ConsumerIP]
				key.ConsumerOp = int(instr.op)
				if instr.node != nil {
					key.ConsumerOrigin = i.nodeOrigins[instr.node]
					span := instr.node.Span()
					key.ConsumerLine = span.Start.Line
					key.ConsumerColumn = span.Start.Column
				}
			}
		}
	}
	var counter *uint64
	if vm != nil {
		counter = vm.bytecodePrimitiveMaterializationCounters[key]
	}
	if counter == nil {
		i.bytecodePrimitiveMaterializationsMu.Lock()
		if i.bytecodePrimitiveMaterializations == nil {
			i.bytecodePrimitiveMaterializations = make(map[bytecodePrimitiveMaterializationKey]*uint64)
		}
		counter = i.bytecodePrimitiveMaterializations[key]
		if counter == nil {
			if len(i.bytecodePrimitiveMaterializations) >= bytecodePrimitiveMaterializationSiteLimit {
				i.bytecodePrimitiveMaterializationsDropped++
				i.bytecodePrimitiveMaterializationsMu.Unlock()
				return
			}
			counter = new(uint64)
			i.bytecodePrimitiveMaterializations[key] = counter
		}
		i.bytecodePrimitiveMaterializationsMu.Unlock()
		if vm != nil {
			if vm.bytecodePrimitiveMaterializationCounters == nil {
				vm.bytecodePrimitiveMaterializationCounters = make(map[bytecodePrimitiveMaterializationKey]*uint64)
			}
			vm.bytecodePrimitiveMaterializationCounters[key] = counter
		}
	}
	atomic.AddUint64(counter, 1)
}

func (i *Interpreter) bytecodePrimitiveMaterializationSnapshot() ([]BytecodePrimitiveMaterializationSnapshot, uint64) {
	if i == nil {
		return nil, 0
	}
	i.bytecodePrimitiveMaterializationsMu.Lock()
	rows := make([]BytecodePrimitiveMaterializationSnapshot, 0, len(i.bytecodePrimitiveMaterializations))
	for key, counter := range i.bytecodePrimitiveMaterializations {
		count := atomic.LoadUint64(counter)
		if count == 0 {
			continue
		}
		opcode := ""
		if key.Op >= 0 && key.Op < bytecodeOpCount {
			opcode = bytecodeOpName(bytecodeOp(key.Op))
		}
		consumerOpcode := ""
		if key.ConsumerOp >= 0 && key.ConsumerOp < bytecodeOpCount {
			consumerOpcode = bytecodeOpName(bytecodeOp(key.ConsumerOp))
		}
		rows = append(rows, BytecodePrimitiveMaterializationSnapshot{
			Class: key.Class, Reason: key.Reason, Carrier: key.Carrier, Suffix: key.Suffix,
			Op: key.Op, Opcode: opcode, IP: key.IP, Program: key.Program, Origin: key.Origin,
			Line: key.Line, Column: key.Column, ReturnFrame: key.ReturnFrame,
			ConsumerOp: key.ConsumerOp, ConsumerOpcode: consumerOpcode, ConsumerIP: key.ConsumerIP,
			ConsumerProgram: key.ConsumerProgram, ConsumerOrigin: key.ConsumerOrigin,
			ConsumerLine: key.ConsumerLine, ConsumerColumn: key.ConsumerColumn, Count: count,
		})
	}
	dropped := i.bytecodePrimitiveMaterializationsDropped
	i.bytecodePrimitiveMaterializationsMu.Unlock()
	sort.Slice(rows, func(left, right int) bool {
		if rows[left].Count != rows[right].Count {
			return rows[left].Count > rows[right].Count
		}
		if rows[left].Reason != rows[right].Reason {
			return rows[left].Reason < rows[right].Reason
		}
		if rows[left].Origin != rows[right].Origin {
			return rows[left].Origin < rows[right].Origin
		}
		if rows[left].Line != rows[right].Line {
			return rows[left].Line < rows[right].Line
		}
		return rows[left].IP < rows[right].IP
	})
	return rows, dropped
}

func (i *Interpreter) resetBytecodePrimitiveMaterializations() {
	if i == nil {
		return
	}
	i.bytecodePrimitiveMaterializationsMu.Lock()
	for _, counter := range i.bytecodePrimitiveMaterializations {
		atomic.StoreUint64(counter, 0)
	}
	i.bytecodePrimitiveMaterializationsDropped = 0
	i.bytecodePrimitiveMaterializationsMu.Unlock()
}
