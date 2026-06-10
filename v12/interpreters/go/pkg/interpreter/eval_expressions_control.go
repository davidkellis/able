package interpreter

import (
	"fmt"
	"math/big"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
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

func (i *Interpreter) isTruthy(val runtime.Value) bool {
	switch v := val.(type) {
	case runtime.BoolValue:
		return v.Val
	case *runtime.BoolValue:
		return v != nil && v.Val
	case runtime.NilValue:
		return false
	case *runtime.NilValue:
		return false
	case runtime.ErrorValue:
		return false
	case *runtime.ErrorValue:
		return false
	}
	return !i.matchesErrorValue(val)
}

func (i *Interpreter) IsTruthy(val runtime.Value) bool {
	return i.isTruthy(val)
}

func isNumericValue(val runtime.Value) bool {
	switch val.(type) {
	case runtime.IntegerValue, runtime.FloatValue:
		return true
	case *runtime.StructInstanceValue:
		return isRatioValue(val)
	default:
		if _, _, ok := bytecodeDirectRawFloatValue(val); ok {
			return true
		}
		_, _, ok := bytecodeRawIntegerValueInfo(val)
		return ok
	}
}

func numericToFloat(val runtime.Value) (float64, error) {
	switch v := val.(type) {
	case runtime.FloatValue:
		return v.Val, nil
	case runtime.IntegerValue:
		if n, ok := v.ToInt64(); ok {
			return float64(n), nil
		}
		return bigIntToFloat(v.BigInt()), nil
	case *runtime.StructInstanceValue:
		if isRatioValue(v) {
			parts, err := coerceToRatio(v)
			if err != nil {
				return 0, err
			}
			num := new(big.Rat).SetFrac(parts.num, parts.den)
			if num == nil {
				return 0, fmt.Errorf("Arithmetic requires numeric operands")
			}
			f, _ := num.Float64()
			return f, nil
		}
		return 0, fmt.Errorf("Arithmetic requires numeric operands")
	default:
		if raw, _, ok := bytecodeDirectRawFloatValue(val); ok {
			return raw, nil
		}
		if _, raw, ok := bytecodeRawIntegerValueInfo(val); ok {
			return float64(raw), nil
		}
		return 0, fmt.Errorf("Arithmetic requires numeric operands")
	}
}

func assignStructMember(interp *Interpreter, inst *runtime.StructInstanceValue, member ast.Expression, value runtime.Value, operator ast.AssignmentOperator, binaryOp string, isCompound bool) (runtime.Value, error) {
	if inst == nil {
		return nil, fmt.Errorf("struct instance is nil")
	}
	switch mem := member.(type) {
	case *ast.Identifier:
		if !structUsesNamedFieldStorage(inst) {
			return nil, fmt.Errorf("Expected named struct instance")
		}
		current, ok := structNamedFieldValue(inst, mem.Name)
		if !ok {
			return nil, fmt.Errorf("No field named '%s'", mem.Name)
		}
		if mem.Name == "storage_handle" && isCanonicalArrayStructInstance(inst) {
			if err := prepareCanonicalArrayStructStorageHandle(inst, value, operator); err != nil {
				return nil, err
			}
			if !structSetNamedFieldValue(inst, "storage_handle", value) {
				return nil, fmt.Errorf("No field named 'storage_handle'")
			}
			return value, nil
		}
		if operator == ast.AssignmentAssign {
			if !structSetNamedFieldValue(inst, mem.Name, value) {
				return nil, fmt.Errorf("No field named '%s'", mem.Name)
			}
			return value, nil
		}
		if !isCompound {
			return nil, fmt.Errorf("unsupported assignment operator %s", operator)
		}
		computed, err := applyBinaryOperator(interp, binaryOp, current, value)
		if err != nil {
			return nil, err
		}
		if !structSetNamedFieldValue(inst, mem.Name, computed) {
			return nil, fmt.Errorf("No field named '%s'", mem.Name)
		}
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
		if isCanonicalArrayStructInstance(inst) &&
			idx < len(inst.Definition.Node.Fields) &&
			inst.Definition.Node.Fields[idx] != nil &&
			inst.Definition.Node.Fields[idx].Name != nil &&
			inst.Definition.Node.Fields[idx].Name.Name == "storage_handle" {
			if err := prepareCanonicalArrayStructStorageHandle(inst, value, operator); err != nil {
				return nil, err
			}
			inst.Positional[idx] = value
			return value, nil
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

func isCanonicalArrayStructInstance(inst *runtime.StructInstanceValue) bool {
	return inst != nil &&
		inst.Definition != nil &&
		inst.Definition.Node != nil &&
		inst.Definition.Node.ID != nil &&
		inst.Definition.Node.ID.Name == "Array"
}

func prepareCanonicalArrayStructStorageHandle(inst *runtime.StructInstanceValue, value runtime.Value, operator ast.AssignmentOperator) error {
	if operator != ast.AssignmentAssign {
		return fmt.Errorf("unsupported assignment operator %s", operator)
	}
	integer, ok := value.(runtime.IntegerValue)
	if !ok {
		return fmt.Errorf("array storage_handle must be an integer")
	}
	handle, fits := integer.ToInt64()
	if !fits || handle <= 0 {
		return fmt.Errorf("array storage_handle must be positive")
	}
	if _, err := runtime.ArrayStoreEnsureHandle(handle, 0, 0); err != nil {
		return err
	}
	if err := runtime.ArrayStoreTrackStructInstanceLease(inst, handle); err != nil {
		return err
	}
	return nil
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
