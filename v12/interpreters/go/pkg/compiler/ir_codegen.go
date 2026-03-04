package compiler

import (
	"bytes"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

// IRGoOptions configures IR -> Go emission.
type IRGoOptions struct {
	PackageName string
}

// EmitIRFunction renders a single IR function into a Go source file.
func EmitIRFunction(fn *IRFunction, opts IRGoOptions) ([]byte, error) {
	if fn == nil {
		return nil, fmt.Errorf("compiler: IR function is nil")
	}
	if opts.PackageName == "" {
		opts.PackageName = "main"
	}
	emitter := newIRGoEmitter(opts)
	return emitter.emitFunctionFile(fn)
}

type irGoEmitter struct {
	opts          IRGoOptions
	buf           bytes.Buffer
	mangler       *nameMangler
	valueNames    map[*IRValue]string
	slotNames     map[*IRSlot]string
	labelNames    map[string]string
	funcNames     map[*IRFunction]string
	paramValues   map[*IRValue]struct{}
	callCallees   map[*IRValue]struct{}
	declaredVals  []*IRValue
	patternSlots  map[ast.Node]*IRSlot
	needsIterator bool
}

func newIRGoEmitter(opts IRGoOptions) *irGoEmitter {
	mangler := newNameMangler()
	mangler.seen["rt"] = 1
	mangler.seen["err"] = 1
	return &irGoEmitter{
		opts:         opts,
		mangler:      mangler,
		valueNames:   make(map[*IRValue]string),
		slotNames:    make(map[*IRSlot]string),
		labelNames:   make(map[string]string),
		funcNames:    make(map[*IRFunction]string),
		paramValues:  make(map[*IRValue]struct{}),
		callCallees:  make(map[*IRValue]struct{}),
		patternSlots: make(map[ast.Node]*IRSlot),
	}
}

func (e *irGoEmitter) prepare(fn *IRFunction) error {
	e.resetFunctionState()
	if err := ValidateIR(fn); err != nil {
		return err
	}
	for _, param := range fn.Params {
		if param.Value == nil {
			continue
		}
		e.paramValues[param.Value] = struct{}{}
		e.nameForValue(param.Value)
	}
	for _, slot := range fn.Locals {
		if slot == nil {
			continue
		}
		e.nameForSlot(slot)
		if slot.Source != nil {
			e.patternSlots[slot.Source] = slot
		}
	}
	for _, slot := range fn.Captured {
		if slot == nil {
			continue
		}
		e.nameForSlot(slot)
	}
	for label := range fn.Blocks {
		e.nameForLabel(label)
	}
	for _, block := range fn.Blocks {
		if block == nil {
			continue
		}
		for _, instr := range block.Instructions {
			switch node := instr.(type) {
			case *IRCompute:
				e.collectValue(node.Dest)
			case *IRInvoke:
				e.collectValue(node.Value)
				e.collectValue(node.Error)
				if node.Op == IROpCall {
					if use, ok := node.Callee.(IRValueUse); ok && use.Value != nil {
						e.callCallees[use.Value] = struct{}{}
					}
				}
			case *IRSpawn:
				e.collectValue(node.Value)
				e.collectValue(node.Error)
				for _, slot := range node.Captures {
					if slot == nil {
						continue
					}
					e.nameForSlot(slot)
				}
			case *IRLoad:
				e.collectValue(node.Dest)
			case *IRIterNext:
				e.collectValue(node.Value)
				e.collectValue(node.Done)
				e.collectValue(node.Error)
			case *IRArrayLiteral:
				e.collectValue(node.Dest)
			case *IRStructLiteral:
				e.collectValue(node.Dest)
			case *IRMapLiteral:
				e.collectValue(node.Dest)
			case *IRStringInterpolation:
				e.collectValue(node.Dest)
			case *IRIteratorLiteral:
				e.collectValue(node.Dest)
				e.needsIterator = true
				for _, slot := range node.Captures {
					if slot == nil {
						continue
					}
					e.nameForSlot(slot)
				}
			}
		}
	}
	return nil
}

func (e *irGoEmitter) collectValue(val *IRValue) {
	if val == nil {
		return
	}
	if _, ok := e.valueNames[val]; ok {
		return
	}
	e.nameForValue(val)
	e.declaredVals = append(e.declaredVals, val)
}

func (e *irGoEmitter) nameForValue(val *IRValue) string {
	if val == nil {
		return ""
	}
	if name, ok := e.valueNames[val]; ok {
		return name
	}
	base := sanitizeIdent(val.Name)
	if base == "" || base == "_" {
		base = fmt.Sprintf("v%d", val.ID)
	}
	name := e.mangler.unique(base)
	e.valueNames[val] = name
	return name
}

func (e *irGoEmitter) nameForSlot(slot *IRSlot) string {
	if slot == nil {
		return ""
	}
	if name, ok := e.slotNames[slot]; ok {
		return name
	}
	base := sanitizeIdent(slot.Name)
	if base == "" || base == "_" {
		base = "slot"
	}
	name := e.mangler.unique("slot_" + base)
	e.slotNames[slot] = name
	return name
}

func (e *irGoEmitter) nameForLabel(label string) string {
	if label == "" {
		return ""
	}
	if name, ok := e.labelNames[label]; ok {
		return name
	}
	base := sanitizeIdent(label)
	if base == "" || base == "_" {
		base = "block"
	}
	name := e.mangler.unique("block_" + base)
	e.labelNames[label] = name
	return name
}

func (e *irGoEmitter) emitFunctionFile(fn *IRFunction) ([]byte, error) {
	if fn == nil {
		return nil, fmt.Errorf("compiler: IR function is nil")
	}
	functions := collectIRFunctions(fn)
	e.scanFunctions(functions)
	fmt.Fprintf(&e.buf, "package %s\n\n", e.opts.PackageName)
	e.emitImports()
	e.emitHelpers()
	for _, f := range functions {
		if err := e.prepare(f); err != nil {
			return nil, err
		}
		if err := e.emitFunction(f); err != nil {
			return nil, err
		}
	}
	return formatSource(e.buf.Bytes())
}

func (e *irGoEmitter) emitImports() {
	fmt.Fprintf(&e.buf, "import (\n")
	if e.needsIterator {
		fmt.Fprintf(&e.buf, "\t%q\n", "errors")
	}
	fmt.Fprintf(&e.buf, "\t%q\n", "fmt")
	fmt.Fprintf(&e.buf, "\t%q\n", "math/big")
	fmt.Fprintf(&e.buf, "\t%q\n", "strings")
	if e.needsIterator {
		fmt.Fprintf(&e.buf, "\t%q\n", "sync")
	}
	fmt.Fprintf(&e.buf, "\t%q\n", "able/interpreter-go/pkg/ast")
	fmt.Fprintf(&e.buf, "\t%q\n", "able/interpreter-go/pkg/compiler/bridge")
	fmt.Fprintf(&e.buf, "\t%q\n", "able/interpreter-go/pkg/runtime")
	fmt.Fprintf(&e.buf, ")\n\n")
}

func (e *irGoEmitter) emitHelpers() {
	fmt.Fprintf(&e.buf, "func __able_is_nil(val runtime.Value) bool {\n")
	fmt.Fprintf(&e.buf, "\tif val == nil {\n")
	fmt.Fprintf(&e.buf, "\t\treturn true\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tswitch val.(type) {\n")
	fmt.Fprintf(&e.buf, "\tcase runtime.NilValue, *runtime.NilValue:\n")
	fmt.Fprintf(&e.buf, "\t\treturn true\n")
	fmt.Fprintf(&e.buf, "\tdefault:\n")
	fmt.Fprintf(&e.buf, "\t\treturn false\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_new_cell(value runtime.Value) *runtime.Value {\n")
	fmt.Fprintf(&e.buf, "\tv := value\n")
	fmt.Fprintf(&e.buf, "\treturn &v\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_cell_value(cell *runtime.Value) runtime.Value {\n")
	fmt.Fprintf(&e.buf, "\tif cell == nil || *cell == nil {\n")
	fmt.Fprintf(&e.buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\treturn *cell\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_is_error(rt *bridge.Runtime, val runtime.Value) bool {\n")
	fmt.Fprintf(&e.buf, "\treturn bridge.IsError(rt, val)\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_truthy(rt *bridge.Runtime, val runtime.Value) bool {\n")
	fmt.Fprintf(&e.buf, "\treturn bridge.IsTruthy(rt, val)\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_error_value(rt *bridge.Runtime, val runtime.Value) runtime.Value {\n")
	fmt.Fprintf(&e.buf, "\treturn bridge.ErrorValue(rt, val)\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_stringify(rt *bridge.Runtime, val runtime.Value) string {\n")
	fmt.Fprintf(&e.buf, "\tstr, err := bridge.Stringify(rt, val)\n")
	fmt.Fprintf(&e.buf, "\tif err != nil {\n")
	fmt.Fprintf(&e.buf, "\t\tpanic(err)\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\treturn str\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_struct_instance(val runtime.Value) *runtime.StructInstanceValue {\n")
	fmt.Fprintf(&e.buf, "\tcurrent := val\n")
	fmt.Fprintf(&e.buf, "\tfor {\n")
	fmt.Fprintf(&e.buf, "\t\tswitch v := current.(type) {\n")
	fmt.Fprintf(&e.buf, "\t\tcase runtime.InterfaceValue:\n")
	fmt.Fprintf(&e.buf, "\t\t\tcurrent = v.Underlying\n")
	fmt.Fprintf(&e.buf, "\t\t\tcontinue\n")
	fmt.Fprintf(&e.buf, "\t\tcase *runtime.InterfaceValue:\n")
	fmt.Fprintf(&e.buf, "\t\t\tif v != nil {\n")
	fmt.Fprintf(&e.buf, "\t\t\t\tcurrent = v.Underlying\n")
	fmt.Fprintf(&e.buf, "\t\t\t\tcontinue\n")
	fmt.Fprintf(&e.buf, "\t\t\t}\n")
	fmt.Fprintf(&e.buf, "\t\tcase runtime.ErrorValue:\n")
	fmt.Fprintf(&e.buf, "\t\t\treturn __able_error_to_struct(v)\n")
	fmt.Fprintf(&e.buf, "\t\tcase *runtime.ErrorValue:\n")
	fmt.Fprintf(&e.buf, "\t\t\tif v != nil {\n")
	fmt.Fprintf(&e.buf, "\t\t\t\treturn __able_error_to_struct(*v)\n")
	fmt.Fprintf(&e.buf, "\t\t\t}\n")
	fmt.Fprintf(&e.buf, "\t\t}\n")
	fmt.Fprintf(&e.buf, "\t\tbreak\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tif inst, ok := current.(*runtime.StructInstanceValue); ok {\n")
	fmt.Fprintf(&e.buf, "\t\treturn inst\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\treturn nil\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_error_to_struct(err runtime.ErrorValue) *runtime.StructInstanceValue {\n")
	fmt.Fprintf(&e.buf, "\tfields := make(map[string]runtime.Value)\n")
	fmt.Fprintf(&e.buf, "\tif err.Payload != nil {\n")
	fmt.Fprintf(&e.buf, "\t\tfor k, v := range err.Payload {\n")
	fmt.Fprintf(&e.buf, "\t\t\tfields[k] = v\n")
	fmt.Fprintf(&e.buf, "\t\t}\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tfields[\"message\"] = runtime.StringValue{Val: err.Message}\n")
	fmt.Fprintf(&e.buf, "\treturn &runtime.StructInstanceValue{Fields: fields}\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_error_from_go(err error) runtime.Value {\n")
	fmt.Fprintf(&e.buf, "\tif err == nil {\n")
	fmt.Fprintf(&e.buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\treturn runtime.ErrorValue{Message: err.Error()}\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_int_literal(text string, suffix runtime.IntegerType) runtime.Value {\n")
	fmt.Fprintf(&e.buf, "\tval, ok := new(big.Int).SetString(text, 10)\n")
	fmt.Fprintf(&e.buf, "\tif !ok {\n")
	fmt.Fprintf(&e.buf, "\t\tpanic(fmt.Errorf(\"invalid integer literal: %%s\", text))\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\treturn runtime.NewBigIntValue(val, suffix)\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_float_literal(value float64, suffix runtime.FloatType) runtime.Value {\n")
	fmt.Fprintf(&e.buf, "\treturn runtime.FloatValue{Val: value, TypeSuffix: suffix}\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	fmt.Fprintf(&e.buf, "func __able_global(rt *bridge.Runtime, name string) runtime.Value {\n")
	fmt.Fprintf(&e.buf, "\tif rt == nil {\n")
	fmt.Fprintf(&e.buf, "\t\tpanic(fmt.Errorf(\"compiler: missing runtime\"))\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tval, err := bridge.Get(rt, name)\n")
	fmt.Fprintf(&e.buf, "\tif err != nil {\n")
	fmt.Fprintf(&e.buf, "\t\tpanic(err)\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tif val == nil {\n")
	fmt.Fprintf(&e.buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\treturn val\n")
	fmt.Fprintf(&e.buf, "}\n\n")
	if e.needsIterator {
		e.emitIteratorHelpers()
	}
}

func (e *irGoEmitter) emitFunction(fn *IRFunction) error {
	funcName := e.functionName(fn)
	captured := make(map[*IRSlot]struct{}, len(fn.Captured))
	for _, slot := range fn.Captured {
		if slot == nil {
			continue
		}
		captured[slot] = struct{}{}
	}
	fmt.Fprintf(&e.buf, "func %s(rt *bridge.Runtime", funcName)
	for _, slot := range fn.Captured {
		if slot == nil {
			continue
		}
		name := e.nameForSlot(slot)
		fmt.Fprintf(&e.buf, ", %s *runtime.Value", name)
	}
	for _, param := range fn.Params {
		if param.Value == nil {
			continue
		}
		name := e.nameForValue(param.Value)
		fmt.Fprintf(&e.buf, ", %s runtime.Value", name)
	}
	fmt.Fprintf(&e.buf, ") (runtime.Value, error) {\n")
	fmt.Fprintf(&e.buf, "\tvar err error\n")
	for _, slot := range fn.Locals {
		if slot == nil {
			continue
		}
		if _, ok := captured[slot]; ok {
			continue
		}
		fmt.Fprintf(&e.buf, "\t%s := __able_new_cell(runtime.NilValue{})\n", e.nameForSlot(slot))
	}
	for _, val := range e.sortedDeclaredValues() {
		if val == nil {
			continue
		}
		if _, isParam := e.paramValues[val]; isParam {
			continue
		}
		fmt.Fprintf(&e.buf, "\tvar %s runtime.Value\n", e.nameForValue(val))
	}
	entryLabel := e.nameForLabel(fn.EntryLabel)
	fmt.Fprintf(&e.buf, "\tgoto %s\n", entryLabel)

	labels := e.sortedLabels(fn)
	for _, label := range labels {
		block := fn.Blocks[label]
		if block == nil {
			continue
		}
		fmt.Fprintf(&e.buf, "%s:\n", e.nameForLabel(label))
		for _, instr := range block.Instructions {
			if err := e.emitInstruction(instr); err != nil {
				return err
			}
		}
		if block.Terminator == nil {
			return fmt.Errorf("compiler: IR block %q missing terminator", label)
		}
		if err := e.emitTerminator(block.Terminator); err != nil {
			return err
		}
	}
	fmt.Fprintf(&e.buf, "}\n\n")
	return nil
}

func (e *irGoEmitter) emitInstruction(instr IRInstruction) error {
	switch node := instr.(type) {
	case *IRNoop:
		return nil
	case *IRLoad:
		dest := e.nameForValue(node.Dest)
		slot := e.nameForSlot(node.Slot)
		fmt.Fprintf(&e.buf, "\t%s = __able_cell_value(%s)\n", dest, slot)
		return nil
	case *IRStore:
		slot := e.nameForSlot(node.Slot)
		value, err := e.emitValueRef(node.Value)
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t*%s = %s\n", slot, value)
		return nil
	case *IRCompute:
		return e.emitCompute(node)
	case *IRInvoke:
		return e.emitInvoke(node)
	case *IRSpawn:
		return e.emitSpawn(node)
	case *IRIterNext:
		return e.emitIterNext(node)
	case *IRArrayLiteral:
		return e.emitArrayLiteral(node)
	case *IRStructLiteral:
		return e.emitStructLiteral(node)
	case *IRMapLiteral:
		return e.emitMapLiteral(node)
	case *IRStringInterpolation:
		return e.emitStringInterpolation(node)
	case *IRIteratorLiteral:
		return e.emitIteratorLiteral(node)
	case *IRDestructure:
		return e.emitDestructure(node)
	default:
		return fmt.Errorf("compiler: unsupported IR instruction %T", instr)
	}
}

func (e *irGoEmitter) emitCompute(node *IRCompute) error {
	if node == nil || node.Dest == nil {
		return fmt.Errorf("compiler: IR compute missing destination")
	}
	dest := e.nameForValue(node.Dest)
	switch node.Op {
	case IROpUnary:
		if len(node.Args) != 1 {
			return fmt.Errorf("compiler: unary expects 1 arg")
		}
		arg, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t%s, err = bridge.ApplyUnaryOperator(rt, %q, %s)\n", dest, node.Operator, arg)
		fmt.Fprintf(&e.buf, "\tif err != nil {\n\t\treturn nil, err\n\t}\n")
		fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\t%s = runtime.NilValue{}\n\t}\n", dest, dest)
		return nil
	case IROpBinary:
		if len(node.Args) != 2 {
			return fmt.Errorf("compiler: binary expects 2 args")
		}
		left, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		right, err := e.emitValueRef(node.Args[1])
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t%s, err = bridge.ApplyBinaryOperator(rt, %q, %s, %s)\n", dest, node.Operator, left, right)
		fmt.Fprintf(&e.buf, "\tif err != nil {\n\t\treturn nil, err\n\t}\n")
		fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\t%s = runtime.NilValue{}\n\t}\n", dest, dest)
		return nil
	case IROpIsNil:
		if len(node.Args) != 1 {
			return fmt.Errorf("compiler: is_nil expects 1 arg")
		}
		value, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t%s = runtime.BoolValue{Val: __able_is_nil(%s)}\n", dest, value)
		return nil
	case IROpIsError:
		if len(node.Args) != 1 {
			return fmt.Errorf("compiler: is_error expects 1 arg")
		}
		value, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t%s = runtime.BoolValue{Val: __able_is_error(rt, %s)}\n", dest, value)
		return nil
	case IROpAsError:
		if len(node.Args) != 1 {
			return fmt.Errorf("compiler: as_error expects 1 arg")
		}
		value, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t%s = __able_error_value(rt, %s)\n", dest, value)
		return nil
	default:
		return fmt.Errorf("compiler: unsupported compute op %s", node.Op)
	}
}

func (e *irGoEmitter) emitInvoke(node *IRInvoke) error {
	if node == nil || node.Value == nil || node.Error == nil {
		return fmt.Errorf("compiler: IR invoke missing destination")
	}
	dest := e.nameForValue(node.Value)
	errDest := e.nameForValue(node.Error)
	switch node.Op {
	case IROpCall:
		callee, err := e.emitValueRef(node.Callee)
		if err != nil {
			return err
		}
		args, err := e.emitArgs(node.Args)
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t%s, err = bridge.CallValue(rt, %s, %s)\n", dest, callee, args)
		fmt.Fprintf(&e.buf, "\t%s = __able_error_from_go(err)\n", errDest)
		fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\t%s = runtime.NilValue{}\n\t}\n", dest, dest)
		return nil
	case IROpMember:
		if len(node.Args) != 2 {
			return fmt.Errorf("compiler: member expects 2 args")
		}
		obj, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		member, err := e.emitValueRef(node.Args[1])
		if err != nil {
			return err
		}
		memberFn := "bridge.MemberGet"
		if _, ok := e.callCallees[node.Value]; ok {
			memberFn = "bridge.MemberGetPreferMethods"
		}
		fmt.Fprintf(&e.buf, "\t%s, err = %s(rt, %s, %s)\n", dest, memberFn, obj, member)
		fmt.Fprintf(&e.buf, "\t%s = __able_error_from_go(err)\n", errDest)
		fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\t%s = runtime.NilValue{}\n\t}\n", dest, dest)
		return nil
	case IROpIndex:
		if len(node.Args) != 2 {
			return fmt.Errorf("compiler: index expects 2 args")
		}
		obj, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		idx, err := e.emitValueRef(node.Args[1])
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t%s, err = bridge.Index(rt, %s, %s)\n", dest, obj, idx)
		fmt.Fprintf(&e.buf, "\t%s = __able_error_from_go(err)\n", errDest)
		fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\t%s = runtime.NilValue{}\n\t}\n", dest, dest)
		return nil
	case IROpRange:
		if len(node.Args) != 2 {
			return fmt.Errorf("compiler: range expects 2 args")
		}
		start, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		end, err := e.emitValueRef(node.Args[1])
		if err != nil {
			return err
		}
		inclusive := false
		if expr, ok := node.Source.(*ast.RangeExpression); ok && expr != nil {
			inclusive = expr.Inclusive
		}
		fmt.Fprintf(&e.buf, "\t%s, err = bridge.Range(rt, %s, %s, %t)\n", dest, start, end, inclusive)
		fmt.Fprintf(&e.buf, "\t%s = __able_error_from_go(err)\n", errDest)
		fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\t%s = runtime.NilValue{}\n\t}\n", dest, dest)
		return nil
	case IROpIterator:
		if len(node.Args) != 1 {
			return fmt.Errorf("compiler: iterator expects 1 arg")
		}
		iterable, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\tvar __able_iter *runtime.IteratorValue\n")
		fmt.Fprintf(&e.buf, "\t__able_iter, err = bridge.ResolveIterator(rt, %s)\n", iterable)
		fmt.Fprintf(&e.buf, "\t%s = __able_error_from_go(err)\n", errDest)
		fmt.Fprintf(&e.buf, "\tif __able_iter == nil {\n\t\t%s = runtime.NilValue{}\n\t} else {\n\t\t%s = __able_iter\n\t}\n", dest, dest)
		return nil
	case IROpPropagate:
		if len(node.Args) != 1 {
			return fmt.Errorf("compiler: propagate expects 1 arg")
		}
		value, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\tif __able_is_error(rt, %s) {\n", value)
		fmt.Fprintf(&e.buf, "\t\t%s = %s\n", errDest, value)
		fmt.Fprintf(&e.buf, "\t\t%s = runtime.NilValue{}\n", dest)
		fmt.Fprintf(&e.buf, "\t} else {\n")
		fmt.Fprintf(&e.buf, "\t\t%s = runtime.NilValue{}\n", errDest)
		fmt.Fprintf(&e.buf, "\t\t%s = %s\n", dest, value)
		fmt.Fprintf(&e.buf, "\t}\n")
		return nil
	case IROpCast:
		if len(node.Args) != 1 {
			return fmt.Errorf("compiler: cast expects 1 arg")
		}
		value, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		castExpr, ok := node.Source.(*ast.TypeCastExpression)
		if !ok || castExpr.TargetType == nil {
			return fmt.Errorf("compiler: cast missing target type")
		}
		rendered, ok := e.renderTypeExpression(castExpr.TargetType)
		if !ok {
			return fmt.Errorf("compiler: unsupported cast target type")
		}
		fmt.Fprintf(&e.buf, "\t%s, err = bridge.Cast(rt, %s, %s)\n", dest, rendered, value)
		fmt.Fprintf(&e.buf, "\t%s = __able_error_from_go(err)\n", errDest)
		fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\t%s = runtime.NilValue{}\n\t}\n", dest, dest)
		return nil
	case IROpSpawn:
		return fmt.Errorf("compiler: IR spawn not supported in codegen yet")
	case IROpAwait:
		if len(node.Args) != 1 {
			return fmt.Errorf("compiler: await expects 1 arg")
		}
		value, err := e.emitValueRef(node.Args[0])
		if err != nil {
			return err
		}
		awaitExpr := e.mangler.unique("await_expr")
		fmt.Fprintf(&e.buf, "\t%s := &ast.AwaitExpression{}\n", awaitExpr)
		fmt.Fprintf(&e.buf, "\t%s, err = bridge.Await(rt, %s, %s)\n", dest, awaitExpr, value)
		fmt.Fprintf(&e.buf, "\t%s = __able_error_from_go(err)\n", errDest)
		fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\t%s = runtime.NilValue{}\n\t}\n", dest, dest)
		return nil
	default:
		return fmt.Errorf("compiler: unsupported invoke op %s", node.Op)
	}
}

func (e *irGoEmitter) emitIterNext(node *IRIterNext) error {
	if node == nil || node.Value == nil || node.Done == nil || node.Error == nil {
		return fmt.Errorf("compiler: IR iter_next missing outputs")
	}
	iterable, err := e.emitValueRef(node.Iterator)
	if err != nil {
		return err
	}
	valueDest := e.nameForValue(node.Value)
	doneDest := e.nameForValue(node.Done)
	errDest := e.nameForValue(node.Error)
	fmt.Fprintf(&e.buf, "\tvar __able_iter_val runtime.Value\n")
	fmt.Fprintf(&e.buf, "\tvar __able_iter_done bool\n")
	fmt.Fprintf(&e.buf, "\tvar __able_iter *runtime.IteratorValue\n")
	fmt.Fprintf(&e.buf, "\tif cast, ok := %s.(*runtime.IteratorValue); ok {\n", iterable)
	fmt.Fprintf(&e.buf, "\t\t__able_iter = cast\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	fmt.Fprintf(&e.buf, "\tif __able_iter == nil {\n")
	fmt.Fprintf(&e.buf, "\t\t%s = runtime.ErrorValue{Message: \"expected iterator\"}\n", errDest)
	fmt.Fprintf(&e.buf, "\t\t%s = runtime.BoolValue{Val: true}\n", doneDest)
	fmt.Fprintf(&e.buf, "\t\t%s = runtime.NilValue{}\n", valueDest)
	fmt.Fprintf(&e.buf, "\t} else {\n")
	fmt.Fprintf(&e.buf, "\t\t__able_iter_val, __able_iter_done, err = __able_iter.Next()\n")
	fmt.Fprintf(&e.buf, "\t\t%s = __able_error_from_go(err)\n", errDest)
	fmt.Fprintf(&e.buf, "\t\t%s = runtime.BoolValue{Val: __able_iter_done}\n", doneDest)
	fmt.Fprintf(&e.buf, "\t\tif __able_iter_val == nil {\n")
	fmt.Fprintf(&e.buf, "\t\t\t%s = runtime.NilValue{}\n", valueDest)
	fmt.Fprintf(&e.buf, "\t\t} else {\n")
	fmt.Fprintf(&e.buf, "\t\t\t%s = __able_iter_val\n", valueDest)
	fmt.Fprintf(&e.buf, "\t\t}\n")
	fmt.Fprintf(&e.buf, "\t}\n")
	return nil
}

func (e *irGoEmitter) emitArrayLiteral(node *IRArrayLiteral) error {
	if node == nil || node.Dest == nil {
		return fmt.Errorf("compiler: IR array literal missing destination")
	}
	dest := e.nameForValue(node.Dest)
	defTemp := e.mangler.unique("array_def")
	handleTemp := e.mangler.unique("array_handle")
	fmt.Fprintf(&e.buf, "\tif rt == nil {\n\t\treturn nil, fmt.Errorf(\"compiler: missing runtime\")\n\t}\n")
	fmt.Fprintf(&e.buf, "\t%s, err := rt.StructDefinition(\"Array\")\n", defTemp)
	fmt.Fprintf(&e.buf, "\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\treturn nil, fmt.Errorf(\"array definition unavailable\")\n\t}\n", defTemp)
	fmt.Fprintf(&e.buf, "\t%s, err := __able_array_with_capacity_impl([]runtime.Value{bridge.ToInt(int64(%d), runtime.IntegerType(\"i32\"))})\n", handleTemp, len(node.Elements))
	fmt.Fprintf(&e.buf, "\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	for idx, elem := range node.Elements {
		expr, err := e.emitValueRef(elem)
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t_, err = __able_array_write_impl([]runtime.Value{%s, bridge.ToInt(int64(%d), runtime.IntegerType(\"i32\")), %s})\n", handleTemp, idx, expr)
		fmt.Fprintf(&e.buf, "\tif err != nil { return nil, err }\n")
	}
	fmt.Fprintf(&e.buf, "\t%s = &runtime.StructInstanceValue{Definition: %s, Fields: map[string]runtime.Value{\"length\": bridge.ToInt(int64(%d), runtime.IntegerType(\"i32\")), \"capacity\": bridge.ToInt(int64(%d), runtime.IntegerType(\"i32\")), \"storage_handle\": %s}, TypeArguments: []ast.TypeExpression{ast.NewWildcardTypeExpression()}}\n", dest, defTemp, len(node.Elements), len(node.Elements), handleTemp)
	return nil
}

func (e *irGoEmitter) emitStructLiteral(node *IRStructLiteral) error {
	if node == nil || node.Dest == nil {
		return fmt.Errorf("compiler: IR struct literal missing destination")
	}
	if node.StructName == "" {
		return fmt.Errorf("compiler: IR struct literal missing struct name")
	}
	dest := e.nameForValue(node.Dest)
	defTemp := e.mangler.unique("struct_def")
	defNodeTemp := e.mangler.unique("struct_node")
	typeArgsTemp := e.mangler.unique("struct_type_args")
	fmt.Fprintf(&e.buf, "\tif rt == nil {\n\t\treturn nil, fmt.Errorf(\"compiler: missing runtime\")\n\t}\n")
	fmt.Fprintf(&e.buf, "\t%s, err := rt.StructDefinition(%q)\n", defTemp, node.StructName)
	fmt.Fprintf(&e.buf, "\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	fmt.Fprintf(&e.buf, "\tif %s == nil || %s.Node == nil || %s.Node.ID == nil {\n\t\treturn nil, fmt.Errorf(\"struct definition '%s' unavailable\")\n\t}\n", defTemp, defTemp, defTemp, node.StructName)
	fmt.Fprintf(&e.buf, "\t%s := %s.Node\n", defNodeTemp, defTemp)
	fmt.Fprintf(&e.buf, "\t%s := []ast.TypeExpression(nil)\n", typeArgsTemp)
	if len(node.TypeArguments) > 0 {
		rendered := make([]string, 0, len(node.TypeArguments))
		for _, arg := range node.TypeArguments {
			expr, ok := e.renderTypeExpression(arg)
			if !ok {
				return fmt.Errorf("compiler: unsupported struct literal type arguments")
			}
			rendered = append(rendered, expr)
		}
		fmt.Fprintf(&e.buf, "\t%s = []ast.TypeExpression{%s}\n", typeArgsTemp, strings.Join(rendered, ", "))
	}
	if node.Positional {
		if len(node.Updates) > 0 {
			return fmt.Errorf("compiler: positional struct literal does not support functional updates")
		}
		fmt.Fprintf(&e.buf, "\tif %s.Kind != ast.StructKindPositional && %s.Kind != ast.StructKindSingleton {\n\t\treturn nil, fmt.Errorf(\"positional struct literal not allowed for struct '%s'\")\n\t}\n", defNodeTemp, defNodeTemp, node.StructName)
		valuesTemp := e.mangler.unique("struct_vals")
		values := make([]string, 0, len(node.Fields))
		for _, field := range node.Fields {
			expr, err := e.emitValueRef(field.Value)
			if err != nil {
				return err
			}
			values = append(values, expr)
		}
		fmt.Fprintf(&e.buf, "\t%s := []runtime.Value{%s}\n", valuesTemp, strings.Join(values, ", "))
		fmt.Fprintf(&e.buf, "\tif len(%s) != len(%s.Fields) {\n\t\treturn nil, fmt.Errorf(\"struct '%s' expects %%d fields, got %%d\", len(%s.Fields), len(%s))\n\t}\n", valuesTemp, defNodeTemp, node.StructName, defNodeTemp, valuesTemp)
		fmt.Fprintf(&e.buf, "\t%s = &runtime.StructInstanceValue{Definition: %s, Positional: %s, TypeArguments: %s}\n", dest, defTemp, valuesTemp, typeArgsTemp)
		return nil
	}

	if len(node.Updates) == 0 {
		fmt.Fprintf(&e.buf, "\tif %s.Kind == ast.StructKindPositional {\n\t\treturn nil, fmt.Errorf(\"named struct literal not allowed for positional struct '%s'\")\n\t}\n", defNodeTemp, node.StructName)
	} else {
		fmt.Fprintf(&e.buf, "\tif %s.Kind == ast.StructKindPositional {\n\t\treturn nil, fmt.Errorf(\"functional update only supported for named structs\")\n\t}\n", defNodeTemp)
	}

	fieldsTemp := e.mangler.unique("struct_fields")
	fmt.Fprintf(&e.buf, "\t%s := make(map[string]runtime.Value, %d)\n", fieldsTemp, len(node.Fields))

	if len(node.Updates) > 0 {
		baseTemp := e.mangler.unique("struct_base")
		fmt.Fprintf(&e.buf, "\tvar %s *runtime.StructInstanceValue\n", baseTemp)
		for _, update := range node.Updates {
			updateExpr, err := e.emitValueRef(update)
			if err != nil {
				return err
			}
			updateTemp := e.mangler.unique("struct_update")
			instTemp := e.mangler.unique("struct_update_inst")
			fmt.Fprintf(&e.buf, "\t%s := %s\n", updateTemp, updateExpr)
			fmt.Fprintf(&e.buf, "\t%s := __able_struct_instance(%s)\n", instTemp, updateTemp)
			fmt.Fprintf(&e.buf, "\tif %s == nil { return nil, fmt.Errorf(\"functional update source must be a struct instance\") }\n", instTemp)
			fmt.Fprintf(&e.buf, "\tif %s.Definition == nil || %s.Definition.Node == nil || %s.Definition.Node.ID == nil || %s.Definition.Node.ID.Name != %q {\n\t\treturn nil, fmt.Errorf(\"functional update source must be same struct type\")\n\t}\n", instTemp, instTemp, instTemp, instTemp, node.StructName)
			fmt.Fprintf(&e.buf, "\tif %s.Fields == nil { return nil, fmt.Errorf(\"functional update only supported for named structs\") }\n", instTemp)
			fmt.Fprintf(&e.buf, "\tif %s == nil { %s = %s }\n", baseTemp, baseTemp, instTemp)
			fmt.Fprintf(&e.buf, "\tfor k, v := range %s.Fields { %s[k] = v }\n", instTemp, fieldsTemp)
		}
		fmt.Fprintf(&e.buf, "\tif len(%s) == 0 && %s != nil {\n", typeArgsTemp, baseTemp)
		fmt.Fprintf(&e.buf, "\t\t%s = %s.TypeArguments\n", typeArgsTemp, baseTemp)
		fmt.Fprintf(&e.buf, "\t}\n")
	}

	for _, field := range node.Fields {
		name := field.Name
		if name == "" && field.IsShorthand {
			if valName, ok := e.valueNameFromRef(field.Value); ok {
				name = valName
			}
		}
		if name == "" {
			return fmt.Errorf("compiler: struct literal missing field name")
		}
		valueExpr, err := e.emitValueRef(field.Value)
		if err != nil {
			return err
		}
		fmt.Fprintf(&e.buf, "\t%s[%q] = %s\n", fieldsTemp, name, valueExpr)
	}

	fmt.Fprintf(&e.buf, "\tif %s.Kind == ast.StructKindNamed {\n", defNodeTemp)
	fmt.Fprintf(&e.buf, "\t\tfor _, defField := range %s.Fields {\n", defNodeTemp)
	fmt.Fprintf(&e.buf, "\t\t\tif defField == nil || defField.Name == nil { continue }\n")
	fmt.Fprintf(&e.buf, "\t\t\tif _, ok := %s[defField.Name.Name]; !ok {\n", fieldsTemp)
	fmt.Fprintf(&e.buf, "\t\t\t\treturn nil, fmt.Errorf(\"missing field '%%s' for struct '%s'\", defField.Name.Name)\n", node.StructName)
	fmt.Fprintf(&e.buf, "\t\t\t}\n")
	fmt.Fprintf(&e.buf, "\t\t}\n")
	fmt.Fprintf(&e.buf, "\t}\n")

	fmt.Fprintf(&e.buf, "\t%s = &runtime.StructInstanceValue{Definition: %s, Fields: %s, TypeArguments: %s}\n", dest, defTemp, fieldsTemp, typeArgsTemp)
	return nil
}

func (e *irGoEmitter) emitMapLiteral(node *IRMapLiteral) error {
	if node == nil || node.Dest == nil {
		return fmt.Errorf("compiler: IR map literal missing destination")
	}
	dest := e.nameForValue(node.Dest)
	defTemp := e.mangler.unique("map_def")
	handleTemp := e.mangler.unique("map_handle")
	instTemp := e.mangler.unique("map_inst")
	fmt.Fprintf(&e.buf, "\tif rt == nil {\n\t\treturn nil, fmt.Errorf(\"compiler: missing runtime\")\n\t}\n")
	fmt.Fprintf(&e.buf, "\t%s, err := rt.StructDefinition(\"HashMap\")\n", defTemp)
	fmt.Fprintf(&e.buf, "\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	fmt.Fprintf(&e.buf, "\tif %s == nil {\n\t\treturn nil, fmt.Errorf(\"hash map definition unavailable\")\n\t}\n", defTemp)
	fmt.Fprintf(&e.buf, "\t%s, err := __able_hash_map_new_impl(nil)\n", handleTemp)
	fmt.Fprintf(&e.buf, "\tif err != nil {\n\t\treturn nil, err\n\t}\n")
	fmt.Fprintf(&e.buf, "\t%s := &runtime.StructInstanceValue{Definition: %s, Fields: map[string]runtime.Value{\"handle\": %s}, TypeArguments: []ast.TypeExpression{ast.NewWildcardTypeExpression(), ast.NewWildcardTypeExpression()}}\n", instTemp, defTemp, handleTemp)

	for _, element := range node.Elements {
		switch entry := element.(type) {
		case IRMapEntry:
			keyExpr, err := e.emitValueRef(entry.Key)
			if err != nil {
				return err
			}
			valExpr, err := e.emitValueRef(entry.Value)
			if err != nil {
				return err
			}
			fmt.Fprintf(&e.buf, "\t_, err = __able_hash_map_set_impl([]runtime.Value{%s, %s, %s})\n", handleTemp, keyExpr, valExpr)
			fmt.Fprintf(&e.buf, "\tif err != nil { return nil, err }\n")
		case IRMapSpread:
			spreadExpr, err := e.emitValueRef(entry.Value)
			if err != nil {
				return err
			}
			spreadTemp := e.mangler.unique("map_spread")
			handleVar := e.mangler.unique("map_spread_handle")
			callbackTemp := e.mangler.unique("map_spread_cb")
			fmt.Fprintf(&e.buf, "\t%s := %s\n", spreadTemp, spreadExpr)
			fmt.Fprintf(&e.buf, "\t%s := func(val runtime.Value) runtime.Value {\n", handleVar)
			fmt.Fprintf(&e.buf, "\t\tinst := __able_struct_instance(val)\n")
			fmt.Fprintf(&e.buf, "\t\tif inst == nil || inst.Definition == nil || inst.Definition.Node == nil || inst.Definition.Node.ID == nil || inst.Definition.Node.ID.Name != \"HashMap\" {\n")
			fmt.Fprintf(&e.buf, "\t\t\tpanic(fmt.Errorf(\"map literal spread expects HashMap value\"))\n")
			fmt.Fprintf(&e.buf, "\t\t}\n")
			fmt.Fprintf(&e.buf, "\t\thandle, ok := inst.Fields[\"handle\"]\n")
			fmt.Fprintf(&e.buf, "\t\tif !ok {\n")
			fmt.Fprintf(&e.buf, "\t\t\tpanic(fmt.Errorf(\"map literal spread expects HashMap value\"))\n")
			fmt.Fprintf(&e.buf, "\t\t}\n")
			fmt.Fprintf(&e.buf, "\t\treturn handle\n")
			fmt.Fprintf(&e.buf, "\t}(%s)\n", spreadTemp)
			fmt.Fprintf(&e.buf, "\t%s := runtime.NativeFunctionValue{\n", callbackTemp)
			fmt.Fprintf(&e.buf, "\t\tName: \"__able_map_spread_cb\",\n")
			fmt.Fprintf(&e.buf, "\t\tArity: 2,\n")
			fmt.Fprintf(&e.buf, "\t\tImpl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {\n")
			fmt.Fprintf(&e.buf, "\t\t\tif len(args) != 2 {\n")
			fmt.Fprintf(&e.buf, "\t\t\t\treturn nil, fmt.Errorf(\"map literal spread callback expects key and value\")\n")
			fmt.Fprintf(&e.buf, "\t\t\t}\n")
			fmt.Fprintf(&e.buf, "\t\t\t_, err := __able_hash_map_set_impl([]runtime.Value{%s, args[0], args[1]})\n", handleTemp)
			fmt.Fprintf(&e.buf, "\t\t\tif err != nil {\n")
			fmt.Fprintf(&e.buf, "\t\t\t\treturn nil, err\n")
			fmt.Fprintf(&e.buf, "\t\t\t}\n")
			fmt.Fprintf(&e.buf, "\t\t\treturn runtime.NilValue{}, nil\n")
			fmt.Fprintf(&e.buf, "\t\t},\n")
			fmt.Fprintf(&e.buf, "\t}\n")
			fmt.Fprintf(&e.buf, "\t_, err = __able_hash_map_for_each_impl([]runtime.Value{%s, %s})\n", handleVar, callbackTemp)
			fmt.Fprintf(&e.buf, "\tif err != nil { return nil, err }\n")
		default:
			return fmt.Errorf("compiler: unsupported map literal element %T", element)
		}
	}
	fmt.Fprintf(&e.buf, "\t%s = %s\n", dest, instTemp)
	return nil
}
