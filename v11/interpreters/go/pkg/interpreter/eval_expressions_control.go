package interpreter

import (
	"fmt"

	"able/interpreter10-go/pkg/ast"
	"able/interpreter10-go/pkg/runtime"
)

func comparisonOp(op string, cmp int) bool {
	switch op {
	case "<":
		return cmp < 0
	case "<=":
		return cmp <= 0
	case ">":
		return cmp > 0
	case ">=":
		return cmp >= 0
	case "==":
		return cmp == 0
	case "!=":
		return cmp != 0
	default:
		return false
	}
}

func isTruthy(val runtime.Value) bool {
	switch v := val.(type) {
	case runtime.BoolValue:
		return v.Val
	case runtime.NilValue:
		return false
	default:
		return true
	}
}

func isNumericValue(val runtime.Value) bool {
	switch val.(type) {
	case runtime.IntegerValue, runtime.FloatValue:
		return true
	default:
		return false
	}
}

func numericToFloat(val runtime.Value) (float64, error) {
	switch v := val.(type) {
	case runtime.FloatValue:
		return v.Val, nil
	case runtime.IntegerValue:
		return bigIntToFloat(v.Val), nil
	default:
		return 0, fmt.Errorf("Arithmetic requires numeric operands")
	}
}

func assignStructMember(interp *Interpreter, inst *runtime.StructInstanceValue, member ast.Expression, value runtime.Value, operator ast.AssignmentOperator, binaryOp string, isCompound bool) (runtime.Value, error) {
	if inst == nil {
		return nil, fmt.Errorf("struct instance is nil")
	}
	switch mem := member.(type) {
	case *ast.Identifier:
		if inst.Fields == nil {
			return nil, fmt.Errorf("Expected named struct instance")
		}
		current, ok := inst.Fields[mem.Name]
		if !ok {
			return nil, fmt.Errorf("No field named '%s'", mem.Name)
		}
		if operator == ast.AssignmentAssign {
			inst.Fields[mem.Name] = value
			return value, nil
		}
		if !isCompound {
			return nil, fmt.Errorf("unsupported assignment operator %s", operator)
		}
		computed, err := applyBinaryOperator(interp, binaryOp, current, value)
		if err != nil {
			return nil, err
		}
		inst.Fields[mem.Name] = computed
		return computed, nil
	case *ast.IntegerLiteral:
		if inst.Positional == nil {
			return nil, fmt.Errorf("Expected positional struct instance")
		}
		if mem.Value == nil {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		idx := int(mem.Value.Int64())
		if idx < 0 || idx >= len(inst.Positional) {
			return nil, fmt.Errorf("Struct field index out of bounds")
		}
		if operator == ast.AssignmentAssign {
			inst.Positional[idx] = value
			return value, nil
		}
		if !isCompound {
			return nil, fmt.Errorf("unsupported assignment operator %s", operator)
		}
		current := inst.Positional[idx]
		computed, err := applyBinaryOperator(interp, binaryOp, current, value)
		if err != nil {
			return nil, err
		}
		inst.Positional[idx] = computed
		return computed, nil
	default:
		return nil, fmt.Errorf("Unsupported member assignment target %s", mem.NodeType())
	}
}

func integerBitWidth(t runtime.IntegerType) int {
	switch t {
	case runtime.IntegerI8, runtime.IntegerU8:
		return 8
	case runtime.IntegerI16, runtime.IntegerU16:
		return 16
	case runtime.IntegerI32, runtime.IntegerU32:
		return 32
	case runtime.IntegerI64, runtime.IntegerU64:
		return 64
	case runtime.IntegerI128, runtime.IntegerU128:
		return 128
	default:
		return 0
	}
}
