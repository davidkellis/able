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
	if _, ok := val.(runtime.IntegerValue); ok {
		return true
	}
	_, _, ok := bytecodeRawIntegerValueInfo(val)
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
	method, err := i.findMethodCached(info, dispatch.methodName, dispatch.interfaceName)
	if method != nil || err != nil {
		return method, err
	}
	if i.interfaceMethodResolver != nil {
		if resolved, found := i.interfaceMethodResolver(receiver, dispatch.interfaceName, dispatch.methodName); found {
			return resolved, nil
		}
	}
	return nil, nil
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
	return i.resolveComparisonMethodForInfo(receiver, info, dispatch)
}

func (i *Interpreter) resolveComparisonMethodForInfo(receiver runtime.Value, info typeInfo, dispatch operatorDispatch) (runtime.Value, error) {
	method, err := i.findMethodCached(info, dispatch.methodName, dispatch.interfaceName)
	if method != nil || err != nil {
		return method, err
	}
	if i.interfaceMethodResolver != nil {
		if resolved, found := i.interfaceMethodResolver(receiver, dispatch.interfaceName, dispatch.methodName); found {
			return resolved, nil
		}
	}
	return nil, nil
}

func (i *Interpreter) resolveEqualityMethod(receiver runtime.Value, dispatch operatorDispatch) (runtime.Value, error) {
	if _, primitive := primitiveReceiverTypeName(receiver); primitive {
		callable, found, err := i.resolvePrimitiveInterfaceMethodCallable(receiver, dispatch.methodName, dispatch.interfaceName)
		if err != nil || !found {
			return nil, err
		}
		return unboundPrimitiveInterfaceMethodCallable(callable), nil
	}
	return i.resolveInterfaceMethod(receiver, dispatch.interfaceName, dispatch.methodName)
}

func (i *Interpreter) applyEqualityInterface(op string, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	info, ok := i.getTypeInfoForValue(left)
	if !ok {
		return nil, false, nil
	}
	typeName := i.cachedTypeInfoName(info)
	if entry, ok := i.lookupEqualityDispatchCache(typeName); ok {
		return i.applyCachedEqualityDispatch(op, left, right, entry)
	}
	for _, dispatch := range equalityInterfaces {
		method, err := i.resolveEqualityMethod(left, dispatch)
		if err != nil {
			i.storeEqualityDispatchCache(typeName, equalityDispatchCacheEntry{
				kind:     equalityDispatchCacheError,
				dispatch: dispatch,
				err:      err,
			})
			return nil, true, err
		}
		if method == nil {
			continue
		}
		entry := equalityDispatchCacheEntry{
			kind:     equalityDispatchCacheMethod,
			dispatch: dispatch,
			method:   method,
		}
		_, entry.primitive = primitiveReceiverTypeName(left)
		i.storeEqualityDispatchCache(typeName, entry)
		return i.applyCachedEqualityDispatch(op, left, right, entry)
	}
	i.storeEqualityDispatchCache(typeName, equalityDispatchCacheEntry{
		kind: equalityDispatchCacheNoMethod,
	})
	return nil, false, nil
}

func (i *Interpreter) applyCachedEqualityDispatch(op string, left runtime.Value, right runtime.Value, entry equalityDispatchCacheEntry) (runtime.Value, bool, error) {
	switch entry.kind {
	case equalityDispatchCacheNoMethod:
		return nil, false, nil
	case equalityDispatchCacheError:
		return nil, true, entry.err
	case equalityDispatchCacheMethod:
		if entry.primitive {
			return i.applyCachedPrimitiveEquality(op, left, right)
		}
		if entry.method == nil {
			return nil, false, nil
		}
		result, err := i.callCallableValue2Mutable(entry.method, unwrapInterfaceValue(left), unwrapInterfaceValue(right), nil, nil)
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
			return nil, true, fmt.Errorf("comparison '%s' requires bool result from %s.%s", op, entry.dispatch.interfaceName, entry.dispatch.methodName)
		}
		if op == "!=" {
			boolVal.Val = !boolVal.Val
		}
		return boolVal, true, nil
	default:
		return nil, false, nil
	}
}

func (i *Interpreter) applyCachedPrimitiveEquality(op string, left runtime.Value, right runtime.Value) (runtime.Value, bool, error) {
	left = primitiveCanonicalValue(unwrapInterfaceValue(left))
	right = primitiveCanonicalValue(unwrapInterfaceValue(right))
	coercedRight, err := i.coerceCanonicalPrimitiveInterfaceArg(left, right)
	if err != nil {
		return nil, true, err
	}
	equal, err := primitiveEqualCanonicalValues(left, coercedRight)
	if err != nil {
		return nil, true, err
	}
	if op == "!=" {
		equal = !equal
	}
	return runtime.BoolValue{Val: equal}, true, nil
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
