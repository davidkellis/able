package interpreter

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func primitiveCanonicalValue(value runtime.Value) runtime.Value {
	value = bytecodeMaterializeRawValue(value)
	switch v := value.(type) {
	case *runtime.StringValue:
		if v != nil {
			return *v
		}
	case *runtime.BoolValue:
		if v != nil {
			return *v
		}
	case *runtime.CharValue:
		if v != nil {
			return *v
		}
	case *runtime.IntegerValue:
		if v != nil {
			return *v
		}
	case *runtime.FloatValue:
		if v != nil {
			return *v
		}
	}
	return value
}

func primitiveReceiverTypeName(receiver runtime.Value) (string, bool) {
	switch v := primitiveCanonicalValue(receiver).(type) {
	case runtime.StringValue:
		return "String", true
	case runtime.BoolValue:
		return "bool", true
	case runtime.CharValue:
		return "char", true
	case runtime.NilValue:
		return "nil", true
	case *runtime.NilValue:
		if v == nil {
			return "", false
		}
		return "nil", true
	case runtime.IntegerValue:
		return string(v.TypeSuffix), true
	case runtime.FloatValue:
		return string(v.TypeSuffix), true
	default:
		return "", false
	}
}

func primitiveEqInterfaceName(typeName string, ifaceFilter string) (string, bool) {
	switch ifaceFilter {
	case "":
		if primitiveImplementsInterfaceMethod(typeName, "Eq", "eq") {
			return "Eq", true
		}
		if primitiveImplementsInterfaceMethod(typeName, "PartialEq", "eq") {
			return "PartialEq", true
		}
	case "Eq":
		if primitiveImplementsInterfaceMethod(typeName, "Eq", "eq") {
			return "Eq", true
		}
	case "PartialEq":
		if primitiveImplementsInterfaceMethod(typeName, "PartialEq", "eq") {
			return "PartialEq", true
		}
	}
	return "", false
}

func unboundPrimitiveInterfaceMethodCallable(callable runtime.Value) runtime.Value {
	switch native := callable.(type) {
	case runtime.NativeFunctionValue:
		native.Arity++
		return native
	case *runtime.NativeFunctionValue:
		if native == nil {
			return nil
		}
		clone := *native
		clone.Arity++
		return &clone
	default:
		return callable
	}
}

func (i *Interpreter) resolvePrimitiveInterfaceMethodCallable(receiver runtime.Value, methodName string, ifaceFilter string) (runtime.Value, bool, error) {
	if i == nil || methodName == "" {
		return nil, false, nil
	}
	receiver = primitiveCanonicalValue(receiver)
	typeName, ok := primitiveReceiverTypeName(receiver)
	if !ok {
		return nil, false, nil
	}

	switch methodName {
	case "clone":
		if ifaceFilter != "" && ifaceFilter != "Clone" {
			return nil, false, nil
		}
		if !primitiveImplementsInterfaceMethod(typeName, "Clone", methodName) {
			return nil, false, nil
		}
		return i.primitiveCloneNativeMethod(typeName), true, nil
	case "eq", "ne":
		ifaceName, ok := primitiveEqInterfaceName(typeName, ifaceFilter)
		if !ok {
			return nil, false, nil
		}
		return i.primitiveEqNativeMethod(typeName, ifaceName, methodName), true, nil
	case "cmp":
		if ifaceFilter != "" && ifaceFilter != "Ord" {
			return nil, false, nil
		}
		if !primitiveImplementsInterfaceMethod(typeName, "Ord", methodName) {
			return nil, false, nil
		}
		return i.primitiveCmpNativeMethod(typeName, "Ord"), true, nil
	case "partial_cmp":
		if ifaceFilter != "" && ifaceFilter != "PartialOrd" {
			return nil, false, nil
		}
		if !primitiveImplementsInterfaceMethod(typeName, "PartialOrd", methodName) {
			return nil, false, nil
		}
		return i.primitiveCmpNativeMethod(typeName, "PartialOrd"), true, nil
	case "hash":
		if ifaceFilter != "" && ifaceFilter != "Hash" {
			return nil, false, nil
		}
		if !primitiveImplementsInterfaceMethod(typeName, "Hash", methodName) {
			return nil, false, nil
		}
		return i.primitiveHashNativeMethod(typeName), true, nil
	default:
		return nil, false, nil
	}
}

func (i *Interpreter) primitiveCloneNativeMethod(typeName string) runtime.NativeFunctionValue {
	return runtime.NativeFunctionValue{
		Name:        fmt.Sprintf("%s.clone", strings.ToLower(typeName)),
		Arity:       0,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("clone expects only a receiver")
			}
			return primitiveCanonicalValue(args[0]), nil
		},
	}
}

func (i *Interpreter) primitiveEqNativeMethod(typeName string, ifaceName string, methodName string) runtime.NativeFunctionValue {
	return runtime.NativeFunctionValue{
		Name:        fmt.Sprintf("%s.%s", strings.ToLower(typeName), methodName),
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("%s.%s expects a receiver and one argument", ifaceName, methodName)
			}
			left := primitiveCanonicalValue(args[0])
			right, err := i.coercePrimitiveInterfaceArg(left, args[1])
			if err != nil {
				return nil, err
			}
			equal, err := primitiveEqualValues(left, right)
			if err != nil {
				return nil, err
			}
			if methodName == "ne" {
				equal = !equal
			}
			return runtime.BoolValue{Val: equal}, nil
		},
	}
}

func (i *Interpreter) primitiveCmpNativeMethod(typeName string, ifaceName string) runtime.NativeFunctionValue {
	methodName := "cmp"
	if ifaceName == "PartialOrd" {
		methodName = "partial_cmp"
	}
	return runtime.NativeFunctionValue{
		Name:        fmt.Sprintf("%s.%s", strings.ToLower(typeName), methodName),
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("%s.%s expects a receiver and one argument", ifaceName, methodName)
			}
			left := primitiveCanonicalValue(args[0])
			right, err := i.coercePrimitiveInterfaceArg(left, args[1])
			if err != nil {
				return nil, err
			}
			cmp, err := primitiveCompareValues(left, right)
			if err != nil {
				return nil, err
			}
			return i.primitiveOrderingValue(cmp)
		},
	}
}

func (i *Interpreter) primitiveHashNativeMethod(typeName string) runtime.NativeFunctionValue {
	return runtime.NativeFunctionValue{
		Name:        fmt.Sprintf("%s.hash", strings.ToLower(typeName)),
		Arity:       1,
		BorrowArgs:  true,
		SkipContext: true,
		Impl: func(_ *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("Hash.hash expects a receiver and hasher")
			}
			if err := i.writePrimitiveHash(primitiveCanonicalValue(args[0]), args[1]); err != nil {
				return nil, err
			}
			return runtime.NilValue{}, nil
		},
	}
}

func (i *Interpreter) writePrimitiveHash(receiver runtime.Value, hasher runtime.Value) error {
	switch v := primitiveCanonicalValue(receiver).(type) {
	case runtime.StringValue:
		return i.writeHasherBytes(hasher, []byte(v.Val))
	case runtime.BoolValue:
		if v.Val {
			return i.writeHasherBytes(hasher, []byte{1})
		}
		return i.writeHasherBytes(hasher, []byte{0})
	case runtime.CharValue:
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], uint32(v.Val))
		return i.writeHasherBytes(hasher, buf[:])
	case runtime.IntegerValue:
		return i.writePrimitiveIntegerHash(v, hasher)
	default:
		typeName, _ := primitiveReceiverTypeName(receiver)
		if typeName == "" {
			typeName = fmt.Sprintf("%T", receiver)
		}
		return fmt.Errorf("Hash.hash unsupported for primitive type %s", typeName)
	}
}

func (i *Interpreter) writePrimitiveIntegerHash(value runtime.IntegerValue, hasher runtime.Value) error {
	switch value.TypeSuffix {
	case runtime.IntegerI8:
		n, err := integerToInt64(value)
		if err != nil {
			return err
		}
		return i.writeHasherBytes(hasher, []byte{byte(int8(n))})
	case runtime.IntegerI16:
		n, err := integerToInt64(value)
		if err != nil {
			return err
		}
		var buf [2]byte
		binary.BigEndian.PutUint16(buf[:], uint16(int16(n)))
		return i.writeHasherBytes(hasher, buf[:])
	case runtime.IntegerI32:
		n, err := integerToInt64(value)
		if err != nil {
			return err
		}
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], uint32(int32(n)))
		return i.writeHasherBytes(hasher, buf[:])
	case runtime.IntegerI64, runtime.IntegerIsize:
		n, err := integerToInt64(value)
		if err != nil {
			return err
		}
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], uint64(n))
		return i.writeHasherBytes(hasher, buf[:])
	case runtime.IntegerU8:
		n, err := integerToUint64(value)
		if err != nil {
			return err
		}
		return i.writeHasherBytes(hasher, []byte{byte(n)})
	case runtime.IntegerU16:
		n, err := integerToUint64(value)
		if err != nil {
			return err
		}
		var buf [2]byte
		binary.BigEndian.PutUint16(buf[:], uint16(n))
		return i.writeHasherBytes(hasher, buf[:])
	case runtime.IntegerU32:
		n, err := integerToUint64(value)
		if err != nil {
			return err
		}
		var buf [4]byte
		binary.BigEndian.PutUint32(buf[:], uint32(n))
		return i.writeHasherBytes(hasher, buf[:])
	case runtime.IntegerU64, runtime.IntegerUsize:
		n, err := integerToUint64(value)
		if err != nil {
			return err
		}
		var buf [8]byte
		binary.BigEndian.PutUint64(buf[:], n)
		return i.writeHasherBytes(hasher, buf[:])
	case runtime.IntegerI128, runtime.IntegerU128:
		high, low := primitiveIntegerHashU128Parts(value)
		var buf [16]byte
		binary.BigEndian.PutUint64(buf[0:8], high.Uint64())
		binary.BigEndian.PutUint64(buf[8:16], low.Uint64())
		return i.writeHasherBytes(hasher, buf[:])
	default:
		return fmt.Errorf("Hash.hash unsupported for integer type %s", value.TypeSuffix)
	}
}

func primitiveIntegerHashU128Parts(value runtime.IntegerValue) (*big.Int, *big.Int) {
	bits := runtime.CloneBigInt(value.BigInt())
	if value.TypeSuffix == runtime.IntegerI128 && bits.Sign() < 0 {
		bits.Add(bits, new(big.Int).Lsh(big.NewInt(1), 128))
	}
	mask64 := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 64), big.NewInt(1))
	high := new(big.Int).Rsh(new(big.Int).Set(bits), 64)
	low := new(big.Int).And(bits, mask64)
	return high, low
}

func (i *Interpreter) writeHasherBytes(hasher runtime.Value, bytes []byte) error {
	for _, b := range bytes {
		if err := i.callHasherWrite(hasher, "write_u8", runtime.NewSmallInt(int64(b), runtime.IntegerU8)); err != nil {
			return err
		}
	}
	return nil
}

func (i *Interpreter) callHasherWrite(hasher runtime.Value, methodName string, value runtime.Value) error {
	if methodName == "" {
		return fmt.Errorf("hasher write method required")
	}
	switch h := hasher.(type) {
	case *runtime.InterfaceValue:
		if h != nil {
			return i.callHasherInterfaceMember(h, methodName, value)
		}
	case runtime.InterfaceValue:
		return i.callHasherInterfaceMember(&h, methodName, value)
	}
	method, err := i.resolveInterfaceMethod(hasher, "Hasher", methodName)
	if err != nil {
		return err
	}
	if method == nil {
		return fmt.Errorf("Hasher.%s is not available", methodName)
	}
	_, err = i.callCallableValue(method, []runtime.Value{hasher, value}, nil, nil)
	return err
}

func (i *Interpreter) callHasherInterfaceMember(hasher *runtime.InterfaceValue, methodName string, value runtime.Value) error {
	method, err := i.interfaceMember(hasher, ast.ID(methodName))
	if err != nil {
		return err
	}
	_, err = i.callCallableValue(method, []runtime.Value{value}, nil, nil)
	return err
}

func (i *Interpreter) coercePrimitiveInterfaceArg(receiver runtime.Value, other runtime.Value) (runtime.Value, error) {
	receiver = primitiveCanonicalValue(receiver)
	other = primitiveCanonicalValue(other)
	return i.coerceCanonicalPrimitiveInterfaceArg(receiver, other)
}

func (i *Interpreter) coerceCanonicalPrimitiveInterfaceArg(receiver runtime.Value, other runtime.Value) (runtime.Value, error) {
	switch rv := receiver.(type) {
	case runtime.StringValue:
		if _, ok := other.(runtime.StringValue); ok {
			return other, nil
		}
	case runtime.BoolValue:
		if _, ok := other.(runtime.BoolValue); ok {
			return other, nil
		}
	case runtime.CharValue:
		if _, ok := other.(runtime.CharValue); ok {
			return other, nil
		}
	case runtime.IntegerValue:
		if ov, ok := other.(runtime.IntegerValue); ok && ov.TypeSuffix == rv.TypeSuffix {
			return other, nil
		}
	case runtime.FloatValue:
		if ov, ok := other.(runtime.FloatValue); ok && ov.TypeSuffix == rv.TypeSuffix {
			return other, nil
		}
	}

	typeName, ok := primitiveReceiverTypeName(receiver)
	if !ok {
		return nil, fmt.Errorf("primitive receiver type unavailable")
	}
	coerced, err := i.coerceValueToType(ast.Ty(typeName), other)
	if err != nil {
		return nil, err
	}
	return primitiveCanonicalValue(coerced), nil
}

func primitiveEqualValues(left runtime.Value, right runtime.Value) (bool, error) {
	left = primitiveCanonicalValue(left)
	right = primitiveCanonicalValue(right)
	return primitiveEqualCanonicalValues(left, right)
}

func primitiveEqualCanonicalValues(left runtime.Value, right runtime.Value) (bool, error) {
	switch lv := left.(type) {
	case runtime.StringValue:
		rv, ok := right.(runtime.StringValue)
		if !ok {
			return false, fmt.Errorf("eq expects string argument")
		}
		return lv.Val == rv.Val, nil
	case runtime.BoolValue:
		rv, ok := right.(runtime.BoolValue)
		if !ok {
			return false, fmt.Errorf("eq expects bool argument")
		}
		return lv.Val == rv.Val, nil
	case runtime.CharValue:
		rv, ok := right.(runtime.CharValue)
		if !ok {
			return false, fmt.Errorf("eq expects char argument")
		}
		return lv.Val == rv.Val, nil
	case runtime.IntegerValue:
		rv, ok := right.(runtime.IntegerValue)
		if !ok {
			return false, fmt.Errorf("eq expects integer argument")
		}
		return lv.CmpInt(rv) == 0, nil
	case runtime.FloatValue:
		rv, ok := right.(runtime.FloatValue)
		if !ok {
			return false, fmt.Errorf("eq expects float argument")
		}
		return lv.Val == rv.Val, nil
	default:
		return false, fmt.Errorf("eq unsupported for %T", left)
	}
}

func primitiveCompareValues(left runtime.Value, right runtime.Value) (int, error) {
	left = primitiveCanonicalValue(left)
	right = primitiveCanonicalValue(right)
	switch lv := left.(type) {
	case runtime.StringValue:
		rv, ok := right.(runtime.StringValue)
		if !ok {
			return 0, fmt.Errorf("cmp expects string argument")
		}
		return strings.Compare(lv.Val, rv.Val), nil
	case runtime.BoolValue:
		rv, ok := right.(runtime.BoolValue)
		if !ok {
			return 0, fmt.Errorf("cmp expects bool argument")
		}
		switch {
		case !lv.Val && rv.Val:
			return -1, nil
		case lv.Val && !rv.Val:
			return 1, nil
		default:
			return 0, nil
		}
	case runtime.CharValue:
		rv, ok := right.(runtime.CharValue)
		if !ok {
			return 0, fmt.Errorf("cmp expects char argument")
		}
		switch {
		case lv.Val < rv.Val:
			return -1, nil
		case lv.Val > rv.Val:
			return 1, nil
		default:
			return 0, nil
		}
	case runtime.IntegerValue:
		rv, ok := right.(runtime.IntegerValue)
		if !ok {
			return 0, fmt.Errorf("cmp expects integer argument")
		}
		return lv.CmpInt(rv), nil
	case runtime.FloatValue:
		rv, ok := right.(runtime.FloatValue)
		if !ok {
			return 0, fmt.Errorf("cmp expects float argument")
		}
		switch {
		case lv.Val < rv.Val:
			return -1, nil
		case lv.Val > rv.Val:
			return 1, nil
		default:
			return 0, nil
		}
	default:
		return 0, fmt.Errorf("cmp unsupported for %T", left)
	}
}

func (i *Interpreter) primitiveOrderingValue(cmp int) (runtime.Value, error) {
	name := "Equal"
	if cmp < 0 {
		name = "Less"
	} else if cmp > 0 {
		name = "Greater"
	}
	def, err := i.lookupOrderingStruct(name)
	if err != nil {
		return nil, err
	}
	if i.orderingValues == nil {
		i.orderingValues = make(map[string]*runtime.StructInstanceValue)
	}
	if value := i.orderingValues[name]; value != nil && value.Definition == def {
		return value, nil
	}
	value := &runtime.StructInstanceValue{Definition: def}
	i.orderingValues[name] = value
	return value, nil
}

func (i *Interpreter) lookupOrderingStruct(name string) (*runtime.StructDefinitionValue, error) {
	if i == nil || name == "" {
		return nil, fmt.Errorf("ordering struct name required")
	}
	if def := i.orderingStructs[name]; def != nil {
		return def, nil
	}
	if i.global != nil {
		if def, ok := i.global.StructDefinition(name); ok && def != nil {
			i.orderingStructs[name] = def
			return def, nil
		}
		if val, err := i.global.Get(name); err == nil {
			if def, conv := toStructDefinitionValue(val, name); conv == nil && def != nil {
				i.orderingStructs[name] = def
				return def, nil
			}
		}
	}
	if val, ok := i.lookupPackageRegistrySymbol("able.kernel", name); ok {
		if def, conv := toStructDefinitionValue(val, name); conv == nil && def != nil {
			i.orderingStructs[name] = def
			return def, nil
		}
	}
	return nil, fmt.Errorf("ordering struct '%s' is not available", name)
}
