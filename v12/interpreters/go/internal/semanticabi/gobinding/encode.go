package gobinding

import (
	"math"
	"math/big"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/internal/semanticabi/heapmodel"
	"able/interpreter-go/pkg/runtime"
)

type futureMeta struct {
	Status          runtime.FutureStatus
	Started         bool
	CancelRequested bool
}

func (e *encoder) value(value runtime.Value) (semanticabi.Cell, error) {
	if value == nil {
		return semanticabi.ImmediateCell(semanticabi.TagKindNil, 0, 0)
	}
	switch current := value.(type) {
	case runtime.BoolValue:
		payload := uint64(0)
		if current.Val {
			payload = 1
		}
		return semanticabi.ImmediateCell(semanticabi.TagKindBool, 0, payload)
	case runtime.CharValue:
		return semanticabi.ImmediateCell(semanticabi.TagKindChar, 0, uint64(current.Val))
	case runtime.NilValue:
		return semanticabi.ImmediateCell(semanticabi.TagKindNil, 0, 0)
	case runtime.VoidValue:
		return semanticabi.ImmediateCell(semanticabi.TagKindVoid, 0, 0)
	case runtime.IntegerValue:
		integer, ok := current.ToInt64()
		format, err := integerFormat(current.TypeSuffix)
		if err != nil {
			return semanticabi.Cell{}, err
		}
		if ok {
			return semanticabi.ImmediateCell(semanticabi.TagKindInteger, format, uint64(integer))
		}
		wide := fields(semanticabi.LayoutWideScalar)
		bigValue := new(big.Int).Set(current.BigInt())
		sign := uint64(0)
		if bigValue.Sign() < 0 {
			sign = 1
			bigValue.Abs(bigValue)
		}
		_ = heapmodel.SetScalar(semanticabi.LayoutWideScalar, wide, "format", sign)
		_ = heapmodel.SetBytes(semanticabi.LayoutWideScalar, wide, "payload", bigValue.Bytes())
		id, err := e.snapshot.Heap.AllocateLayout(semanticabi.LayoutWideScalar, wide)
		if err != nil {
			return semanticabi.Cell{}, err
		}
		return semanticabi.IndirectImmediateCell(semanticabi.TagKindInteger, format, id)
	case runtime.FloatValue:
		aux := uint32(64)
		if current.TypeSuffix == runtime.FloatF32 {
			aux = 32
		}
		return semanticabi.ImmediateCell(semanticabi.TagKindFloat, aux, math.Float64bits(current.Val))
	case runtime.IteratorEndValue:
		return semanticabi.ImmediateCell(semanticabi.TagKindIteratorEnd, 0, 0)
	case runtime.StringValue:
		return e.string(current, nil)
	case *runtime.StringValue:
		return e.string(*current, current)
	case *runtime.ArrayValue:
		return e.array(current)
	case *runtime.HashMapValue:
		return e.hashMap(current)
	case *runtime.HasherValue:
		return e.hasher(current)
	case *runtime.FunctionValue:
		return e.function(current)
	case *runtime.FunctionOverloadValue:
		return e.functionOverload(current)
	case runtime.StructDefinitionValue:
		copy := current
		return e.structDefinition(&copy)
	case *runtime.StructDefinitionValue:
		return e.structDefinition(current)
	case runtime.TypeRefValue:
		copy := current
		return e.typeRef(&copy)
	case *runtime.TypeRefValue:
		return e.typeRef(current)
	case *runtime.StructInstanceValue:
		return e.structInstance(current)
	case runtime.InterfaceDefinitionValue:
		copy := current
		return e.interfaceDefinition(&copy)
	case *runtime.InterfaceDefinitionValue:
		return e.interfaceDefinition(current)
	case runtime.InterfaceValue:
		copy := current
		return e.interfaceValue(&copy)
	case *runtime.InterfaceValue:
		return e.interfaceValue(current)
	case runtime.UnionDefinitionValue:
		copy := current
		return e.unionDefinition(&copy)
	case *runtime.UnionDefinitionValue:
		return e.unionDefinition(current)
	case runtime.PackageValue:
		copy := current
		return e.packageValue(&copy)
	case *runtime.PackageValue:
		return e.packageValue(current)
	case runtime.DynPackageValue:
		copy := current
		return e.dynPackage(&copy)
	case *runtime.DynPackageValue:
		return e.dynPackage(current)
	case runtime.DynRefValue:
		copy := current
		return e.dynRef(&copy)
	case *runtime.DynRefValue:
		return e.dynRef(current)
	case runtime.ErrorValue:
		copy := current
		return e.errorValue(&copy)
	case *runtime.ErrorValue:
		return e.errorValue(current)
	case runtime.BoundMethodValue:
		copy := current
		return e.boundMethod(&copy)
	case *runtime.BoundMethodValue:
		return e.boundMethod(current)
	case runtime.ImplementationNamespaceValue:
		copy := current
		return e.implementationNamespace(&copy)
	case *runtime.ImplementationNamespaceValue:
		return e.implementationNamespace(current)
	case *runtime.IteratorValue:
		return e.iterator(current)
	case runtime.PartialFunctionValue:
		copy := current
		return e.partialFunction(&copy)
	case *runtime.PartialFunctionValue:
		return e.partialFunction(current)
	case runtime.NativeFunctionValue, *runtime.NativeFunctionValue, *runtime.HostHandleValue, runtime.NativeBoundMethodValue, *runtime.NativeBoundMethodValue, *runtime.FutureValue:
		return e.host(value)
	default:
		return semanticabi.Cell{}, &UnsupportedError{Kind: value.Kind().String(), Reason: "concrete Go value shape is not mapped"}
	}
}

func integerFormat(suffix runtime.IntegerType) (uint32, error) {
	formats := map[runtime.IntegerType]uint32{
		"":                0,
		runtime.IntegerI8: 1, runtime.IntegerI16: 2, runtime.IntegerI32: 3, runtime.IntegerI64: 4,
		runtime.IntegerI128: 5, runtime.IntegerU8: 6, runtime.IntegerU16: 7, runtime.IntegerU32: 8,
		runtime.IntegerU64: 9, runtime.IntegerU128: 10, runtime.IntegerIsize: 11, runtime.IntegerUsize: 12,
	}
	format, ok := formats[suffix]
	if !ok {
		return 0, &UnsupportedError{Kind: "KindInteger", Reason: "unknown integer suffix " + string(suffix)}
	}
	return format, nil
}

func (e *encoder) string(value runtime.StringValue, identity any) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(identity, semanticabi.LayoutString, semanticabi.TagKindString)
	if err != nil || seen {
		return cell, err
	}
	result := fields(semanticabi.LayoutString)
	_ = heapSetBytes(semanticabi.LayoutString, result, "utf8", []byte(value.Val))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) host(value runtime.Value) (semanticabi.Cell, error) {
	tag := uint32(value.Kind()) + 1
	key, keyed := keyOf(value)
	if keyed {
		if id, ok := e.hosts[key]; ok {
			return e.snapshot.Heap.Hosts().Cell(tag, id)
		}
	}
	var retained []semanticabi.Cell
	metadata := any(value)
	if future, ok := value.(*runtime.FutureValue); ok {
		result, failure, status := future.Snapshot()
		for _, retainedValue := range []runtime.Value{result, failure} {
			cell, err := e.value(retainedValue)
			if err != nil {
				return semanticabi.Cell{}, err
			}
			retained = append(retained, cell)
		}
		metadata = futureMeta{Status: status, Started: future.Started(), CancelRequested: future.CancelRequested()}
	}
	id, err := e.snapshot.Heap.Hosts().Register(tag, retained...)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	if keyed {
		e.hosts[key] = id
	}
	e.snapshot.hostMetadata[id] = metadata
	return e.snapshot.Heap.Hosts().Cell(tag, id)
}

func heapSetBytes(layout uint16, values []heapmodel.FieldValue, name string, value []byte) error {
	return heapmodel.SetBytes(layout, values, name, value)
}
