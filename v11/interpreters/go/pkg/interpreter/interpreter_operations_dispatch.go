package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type operatorDispatch struct {
	interfaceName string
	methodName    string
}

var operatorInterfaces = map[string]operatorDispatch{
	"+":   {interfaceName: "Add", methodName: "add"},
	"-":   {interfaceName: "Sub", methodName: "sub"},
	"*":   {interfaceName: "Mul", methodName: "mul"},
	"/":   {interfaceName: "Div", methodName: "div"},
	"%":   {interfaceName: "Rem", methodName: "rem"},
	".&":  {interfaceName: "BitAnd", methodName: "bit_and"},
	".|":  {interfaceName: "BitOr", methodName: "bit_or"},
	".^":  {interfaceName: "BitXor", methodName: "bit_xor"},
	".<<": {interfaceName: "Shl", methodName: "shl"},
	".>>": {interfaceName: "Shr", methodName: "shr"},
}

var unaryInterfaces = map[string]operatorDispatch{
	"-":  {interfaceName: "Neg", methodName: "neg"},
	"~":  {interfaceName: "Not", methodName: "not"},
	".~": {interfaceName: "Not", methodName: "not"},
}

var equalityInterfaces = []operatorDispatch{
	{interfaceName: "Eq", methodName: "eq"},
	{interfaceName: "PartialEq", methodName: "eq"},
}

var orderingInterfaces = []operatorDispatch{
	{interfaceName: "Ord", methodName: "cmp"},
	{interfaceName: "PartialOrd", methodName: "partial_cmp"},
}

func binaryOpForAssignment(op ast.AssignmentOperator) (string, bool) {
	switch op {
	case ast.AssignmentAdd:
		return "+", true
	case ast.AssignmentSub:
		return "-", true
	case ast.AssignmentMul:
		return "*", true
	case ast.AssignmentDiv:
		return "/", true
	case ast.AssignmentMod:
		return "%", true
	case ast.AssignmentBitAnd:
		return ".&", true
	case ast.AssignmentBitOr:
		return ".|", true
	case ast.AssignmentBitXor:
		return ".^", true
	case ast.AssignmentShiftL:
		return ".<<", true
	case ast.AssignmentShiftR:
		return ".>>", true
	default:
		return "", false
	}
}

func normalizeOperator(op string) (string, bool) {
	switch op {
	case ".&":
		return "&", true
	case ".|":
		return "|", true
	case ".^":
		return "^", true
	case ".<<":
		return "<<", true
	case ".>>":
		return ">>", true
	case ".~":
		return "~", true
	case "\\xor":
		return "^", false
	default:
		return op, false
	}
}

func isIntegerValue(val runtime.Value) bool {
	_, ok := val.(runtime.IntegerValue)
	return ok
}

func (i *Interpreter) resolveOperatorMethod(receiver runtime.Value, op string) (runtime.Value, error) {
	dispatch, ok := operatorInterfaces[op]
	if !ok {
		return nil, nil
	}
	info, ok := i.getTypeInfoForValue(receiver)
	if !ok {
		return nil, nil
	}
	return i.findMethod(info, dispatch.methodName, dispatch.interfaceName)
}

func (i *Interpreter) applyOperatorInterface(op string, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	method, err := i.resolveOperatorMethod(left, op)
	if err != nil {
		return nil, true, err
	}
	if method == nil {
		return nil, false, nil
	}
	result, err := i.CallFunction(method, []runtime.Value{unwrapInterfaceValue(left), unwrapInterfaceValue(right)})
	return result, true, err
}

func (i *Interpreter) resolveComparisonMethod(receiver runtime.Value, dispatch operatorDispatch) (runtime.Value, error) {
	info, ok := i.getTypeInfoForValue(receiver)
	if !ok {
		return nil, nil
	}
	return i.findMethod(info, dispatch.methodName, dispatch.interfaceName)
}

func (i *Interpreter) applyEqualityInterface(op string, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	for _, dispatch := range equalityInterfaces {
		method, err := i.resolveComparisonMethod(left, dispatch)
		if err != nil {
			return nil, true, err
		}
		if method == nil {
			continue
		}
		result, err := i.CallFunction(method, []runtime.Value{unwrapInterfaceValue(left), unwrapInterfaceValue(right)})
		if err != nil {
			return nil, true, err
		}
		boolVal, ok := result.(runtime.BoolValue)
		if !ok {
			if ptr, okPtr := result.(*runtime.BoolValue); okPtr && ptr != nil {
				boolVal = *ptr
				ok = true
			}
		}
		if !ok {
			return nil, true, fmt.Errorf("comparison '%s' requires bool result from %s.%s", op, dispatch.interfaceName, dispatch.methodName)
		}
		if op == "!=" {
			boolVal.Val = !boolVal.Val
		}
		return boolVal, true, nil
	}
	return nil, false, nil
}

func orderingName(value runtime.Value) string {
	switch v := value.(type) {
	case runtime.InterfaceValue:
		return orderingName(v.Underlying)
	case *runtime.StructInstanceValue:
		return structInstanceName(v)
	case runtime.StructDefinitionValue:
		return structDefName(v)
	case *runtime.StructDefinitionValue:
		if v == nil {
			return ""
		}
		return structDefName(*v)
	default:
		return ""
	}
}

func orderingToCmp(value runtime.Value) (int, bool) {
	switch orderingName(value) {
	case "Less":
		return -1, true
	case "Equal":
		return 0, true
	case "Greater":
		return 1, true
	default:
		return 0, false
	}
}

func (i *Interpreter) applyOrderingInterface(op string, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	for _, dispatch := range orderingInterfaces {
		method, err := i.resolveComparisonMethod(left, dispatch)
		if err != nil {
			return nil, true, err
		}
		if method == nil {
			continue
		}
		result, err := i.CallFunction(method, []runtime.Value{unwrapInterfaceValue(left), unwrapInterfaceValue(right)})
		if err != nil {
			return nil, true, err
		}
		cmp, ok := orderingToCmp(result)
		if !ok {
			return nil, true, fmt.Errorf("comparison '%s' requires Ordering result from %s.%s", op, dispatch.interfaceName, dispatch.methodName)
		}
		return runtime.BoolValue{Val: comparisonOp(op, cmp)}, true, nil
	}
	return nil, false, nil
}

func (i *Interpreter) applyUnaryInterface(op string, operand runtime.Value) (runtime.Value, bool, error) {
	dispatch, ok := unaryInterfaces[op]
	if !ok {
		return nil, false, nil
	}
	method, err := i.resolveComparisonMethod(operand, dispatch)
	if err != nil {
		return nil, true, err
	}
	if method == nil {
		return nil, false, nil
	}
	result, err := i.CallFunction(method, []runtime.Value{unwrapInterfaceValue(operand)})
	return result, true, err
}
