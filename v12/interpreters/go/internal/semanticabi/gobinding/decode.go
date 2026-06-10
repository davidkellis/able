package gobinding

import (
	"fmt"
	"math"
	"math/big"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/internal/semanticabi/heapmodel"
	"able/interpreter-go/pkg/runtime"
)

func (d *decoder) value(cell semanticabi.Cell) (runtime.Value, error) {
	kind, ok := semanticabi.KindByTag(cell.Tag)
	if !ok {
		return nil, fmt.Errorf("semanticabi Go binding: unknown tag %d", cell.Tag)
	}
	if kind.Class == semanticabi.KindHostRegistry {
		return d.host(cell)
	}
	if kind.Class == semanticabi.KindImmediate {
		switch cell.Tag {
		case semanticabi.TagKindBool:
			return runtime.BoolValue{Val: cell.Payload != 0}, nil
		case semanticabi.TagKindChar:
			return runtime.CharValue{Val: rune(cell.Payload)}, nil
		case semanticabi.TagKindNil:
			return runtime.NilValue{}, nil
		case semanticabi.TagKindVoid:
			return runtime.VoidValue{}, nil
		case semanticabi.TagKindInteger:
			format := cell.Aux &^ semanticabi.CellAuxIndirect
			suffix, err := integerSuffix(format)
			if err != nil {
				return nil, err
			}
			if cell.IsIndirectImmediate() {
				object, err := d.object(semanticabi.ObjectID(cell.Payload), semanticabi.LayoutWideScalar)
				if err != nil {
					return nil, err
				}
				sign := field(object, "format").Scalar
				if sign > 1 {
					return nil, fmt.Errorf("semanticabi Go binding: invalid wide integer sign %d", sign)
				}
				integer := new(big.Int).SetBytes(field(object, "payload").Bytes)
				if sign == 1 {
					integer.Neg(integer)
				}
				return runtime.NewBigIntValue(integer, suffix), nil
			}
			return runtime.NewSmallInt(int64(cell.Payload), suffix), nil
		case semanticabi.TagKindFloat:
			suffix := runtime.FloatF64
			if cell.Aux == 32 {
				suffix = runtime.FloatF32
			}
			return runtime.FloatValue{Val: math.Float64frombits(cell.Payload), TypeSuffix: suffix}, nil
		case semanticabi.TagKindIteratorEnd:
			return runtime.IteratorEnd, nil
		}
	}
	id := semanticabi.ObjectID(cell.Payload)
	if existing, ok := d.objects[id]; ok {
		value, ok := existing.(runtime.Value)
		if !ok {
			return nil, fmt.Errorf("semanticabi Go binding: object %d is not a value", id)
		}
		return value, nil
	}
	switch cell.Tag {
	case semanticabi.TagKindString:
		return d.string(id)
	case semanticabi.TagKindArray:
		return d.array(id)
	case semanticabi.TagKindHashMap:
		return d.hashMap(id)
	case semanticabi.TagKindHasher:
		return d.hasher(id)
	case semanticabi.TagKindFunction:
		return d.function(id)
	case semanticabi.TagKindFunctionOverload:
		return d.functionOverload(id)
	case semanticabi.TagKindStructDefinition:
		return d.structDefinition(id)
	case semanticabi.TagKindTypeRef:
		return d.typeRef(id)
	case semanticabi.TagKindStructInstance:
		return d.structInstance(id)
	case semanticabi.TagKindInterfaceDefinition:
		return d.interfaceDefinition(id)
	case semanticabi.TagKindInterfaceValue:
		return d.interfaceValue(id)
	case semanticabi.TagKindUnionDefinition:
		return d.unionDefinition(id)
	case semanticabi.TagKindPackage:
		return d.packageValue(id)
	case semanticabi.TagKindDynPackage:
		return d.dynPackage(id)
	case semanticabi.TagKindDynRef:
		return d.dynRef(id)
	case semanticabi.TagKindError:
		return d.errorValue(id)
	case semanticabi.TagKindBoundMethod:
		return d.boundMethod(id)
	case semanticabi.TagKindImplementationNamespace:
		return d.implementationNamespace(id)
	case semanticabi.TagKindIterator:
		return d.iterator(id)
	case semanticabi.TagKindPartialFunction:
		return d.partialFunction(id)
	default:
		return nil, &UnsupportedError{Kind: kind.Name, Reason: "decoder has no lossless concrete mapping"}
	}
}

func integerSuffix(format uint32) (runtime.IntegerType, error) {
	formats := map[uint32]runtime.IntegerType{0: "", 1: runtime.IntegerI8, 2: runtime.IntegerI16, 3: runtime.IntegerI32, 4: runtime.IntegerI64, 5: runtime.IntegerI128, 6: runtime.IntegerU8, 7: runtime.IntegerU16, 8: runtime.IntegerU32, 9: runtime.IntegerU64, 10: runtime.IntegerU128, 11: runtime.IntegerIsize, 12: runtime.IntegerUsize}
	suffix, ok := formats[format]
	if !ok {
		return "", &UnsupportedError{Kind: "KindInteger", Reason: fmt.Sprintf("unknown integer format %d", format)}
	}
	return suffix, nil
}

func (d *decoder) host(cell semanticabi.Cell) (runtime.Value, error) {
	id := semanticabi.ObjectID(cell.Payload)
	if existing, ok := d.hosts[id]; ok {
		return existing, nil
	}
	host, err := d.snapshot.Heap.Hosts().Resolve(id)
	if err != nil {
		return nil, err
	}
	original, ok := d.snapshot.hostMetadata[id]
	if !ok {
		return nil, fmt.Errorf("semanticabi Go binding: host metadata %d is missing", id)
	}
	switch value := original.(type) {
	case futureMeta:
		if len(host.Retained) != 2 {
			return nil, unexpected("future retained cells")
		}
		result, err := d.value(host.Retained[0])
		if err != nil {
			return nil, err
		}
		failure, err := d.value(host.Retained[1])
		if err != nil {
			return nil, err
		}
		copy := runtime.NewFuture()
		d.hosts[id] = copy
		if value.Started {
			copy.MarkStarted()
		}
		switch value.Status {
		case runtime.FutureResolved:
			copy.Resolve(result)
		case runtime.FutureFailed:
			copy.Fail(failure)
		case runtime.FutureCancelled:
			copy.Cancel(failure)
		}
		if value.CancelRequested && value.Status == runtime.FuturePending {
			copy.RequestCancel()
		}
		return copy, nil
	case runtime.Value:
		return value, nil
	default:
		return nil, fmt.Errorf("semanticabi Go binding: host metadata %d is not runtime.Value", id)
	}
}

func (d *decoder) object(id semanticabi.ObjectID, layoutID uint16) (heapmodel.Object, error) {
	object, err := d.snapshot.Heap.Resolve(id)
	if err != nil {
		return heapmodel.Object{}, err
	}
	if object.LayoutID != layoutID {
		return heapmodel.Object{}, fmt.Errorf("semanticabi Go binding: object %d layout=%d want=%d", id, object.LayoutID, layoutID)
	}
	return object, nil
}

func field(object heapmodel.Object, name string) heapmodel.FieldValue {
	layout, _ := semanticabi.ObjectLayoutByID(object.LayoutID)
	for index, descriptor := range layout.Fields {
		if descriptor.Name == name {
			return object.Fields[index]
		}
	}
	panic("missing generated layout field " + name)
}

func (d *decoder) string(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutString)
	if err != nil {
		return nil, err
	}
	value := runtime.StringValue{Val: string(field(object, "utf8").Bytes)}
	d.objects[id] = value
	return value, nil
}
