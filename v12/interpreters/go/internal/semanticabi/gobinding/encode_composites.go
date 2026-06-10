package gobinding

import (
	"fmt"
	"strings"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/internal/semanticabi/heapmodel"
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

type functionMeta struct {
	Declaration    ast.Node
	MethodPriority float64
	TypeQualified  bool
	MethodSet      *runtime.MethodSet
	Bytecode       any
}

type structInstanceMeta struct {
	Names         []string
	Positional    int
	TypeArguments []ast.TypeExpression
	Native        any
}

type interfaceValueMeta struct {
	MethodNames       []string
	SharedMethodNames []string
	BoundMethodName   string
	InterfaceArgs     []ast.TypeExpression
}

type errorMeta struct {
	TypeName *ast.Identifier
	Names    []string
}

func (e *encoder) array(value *runtime.ArrayValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutArray, semanticabi.TagKindArray)
	if err != nil || seen {
		return cell, err
	}
	elements, err := logicalArrayElements(value)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	cells, err := e.values(elements)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutArray)
	_ = heapmodel.SetCells(semanticabi.LayoutArray, result, "elements", cells...)
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func logicalArrayElements(value *runtime.ArrayValue) ([]runtime.Value, error) {
	if value.Handle > 0 {
		size, err := runtime.ArrayStoreSize(value.Handle)
		if err != nil {
			return nil, err
		}
		result := make([]runtime.Value, size)
		for index := range result {
			result[index], err = runtime.ArrayStoreRead(value.Handle, index)
			if err != nil {
				return nil, err
			}
		}
		return result, nil
	}
	if value.State != nil {
		return append([]runtime.Value(nil), value.State.Values...), nil
	}
	return append([]runtime.Value(nil), value.Elements...), nil
}

func (e *encoder) hashMap(value *runtime.HashMapValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutHashMap, semanticabi.TagKindHashMap)
	if err != nil || seen {
		return cell, err
	}
	flat := make([]runtime.Value, 0, len(value.Entries)*2)
	hashes := make([]uint64, 0, len(value.Entries))
	for _, entry := range value.Entries {
		flat = append(flat, entry.Key, entry.Value)
		hashes = append(hashes, entry.Hash)
	}
	cells, err := e.values(flat)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutHashMap)
	_ = heapmodel.SetCells(semanticabi.LayoutHashMap, result, "entries", cells...)
	_ = heapmodel.SetScalar(semanticabi.LayoutHashMap, result, "hashes", e.metadata(hashes))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) hasher(value *runtime.HasherValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutHasher, semanticabi.TagKindHasher)
	if err != nil || seen {
		return cell, err
	}
	result := fields(semanticabi.LayoutHasher)
	_ = heapmodel.SetScalar(semanticabi.LayoutHasher, result, "state", value.SemanticState())
	return cell, e.snapshot.Heap.Initialize(id, result)
}

type iteratorDriverMeta struct {
	Next     func() (runtime.Value, bool, error)
	NextRaw  func() (runtime.RawValue, bool, error)
	Finalize func()
}

func (e *encoder) iterator(value *runtime.IteratorValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutIterator, semanticabi.TagKindIterator)
	if err != nil || seen {
		return cell, err
	}
	driver, closed := value.HostDriverSnapshot()
	retained, err := e.values(driver.Retained)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	hostID, err := e.snapshot.Heap.Hosts().Register(semanticabi.TagKindHostHandle, retained...)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	e.snapshot.hostMetadata[hostID] = iteratorDriverMeta{Next: driver.Next, NextRaw: driver.NextRaw, Finalize: driver.Finalize}
	driverCell, err := e.snapshot.Heap.Hosts().Cell(semanticabi.TagKindHostHandle, hostID)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutIterator)
	_ = heapmodel.SetCells(semanticabi.LayoutIterator, result, "driver", driverCell)
	_ = heapmodel.SetCells(semanticabi.LayoutIterator, result, "retained", retained...)
	if closed {
		_ = heapmodel.SetScalar(semanticabi.LayoutIterator, result, "closed", 1)
	}
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) function(value *runtime.FunctionValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutFunction, semanticabi.TagKindFunction)
	if err != nil || seen {
		return cell, err
	}
	environment, err := e.environment(value.Closure)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutFunction)
	meta := functionMeta{value.Declaration, value.MethodPriority, value.TypeQualified, value.MethodSet, value.Bytecode}
	_ = heapmodel.SetScalar(semanticabi.LayoutFunction, result, "declaration", e.metadata(meta))
	_ = heapmodel.SetObjects(semanticabi.LayoutFunction, result, "environment", environment)
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) functionOverload(value *runtime.FunctionOverloadValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutFunctionOverload, semanticabi.TagKindFunctionOverload)
	if err != nil || seen {
		return cell, err
	}
	refs := make([]semanticabi.ObjectID, 0, len(value.Overloads))
	for _, fn := range value.Overloads {
		encoded, err := e.function(fn)
		if err != nil {
			return semanticabi.Cell{}, err
		}
		refs = append(refs, semanticabi.ObjectID(encoded.Payload))
	}
	result := fields(semanticabi.LayoutFunctionOverload)
	_ = heapmodel.SetObjects(semanticabi.LayoutFunctionOverload, result, "functions", refs...)
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) environment(value *runtime.Environment) (semanticabi.ObjectID, error) {
	if value == nil {
		return 0, nil
	}
	if id, ok := e.envs[value]; ok {
		return id, nil
	}
	id, err := e.snapshot.Heap.ReserveLayout(semanticabi.LayoutEnvironment)
	if err != nil {
		return 0, err
	}
	e.envs[value] = id
	parent, err := e.environment(value.Parent())
	if err != nil {
		return 0, err
	}
	snapshot := value.Snapshot()
	bindings := make([]semanticabi.ObjectID, 0, len(snapshot))
	for _, name := range sortedKeys(snapshot) {
		cell, err := e.value(snapshot[name])
		if err != nil {
			return 0, err
		}
		bindingFields := fields(semanticabi.LayoutBindingCell)
		_ = heapmodel.SetBytes(semanticabi.LayoutBindingCell, bindingFields, "name", []byte(name))
		_ = heapmodel.SetCells(semanticabi.LayoutBindingCell, bindingFields, "value", cell)
		binding, err := e.snapshot.Heap.AllocateLayout(semanticabi.LayoutBindingCell, bindingFields)
		if err != nil {
			return 0, err
		}
		bindings = append(bindings, binding)
	}
	result := fields(semanticabi.LayoutEnvironment)
	_ = heapmodel.SetObjects(semanticabi.LayoutEnvironment, result, "parent", parent)
	_ = heapmodel.SetObjects(semanticabi.LayoutEnvironment, result, "bindings", bindings...)
	return id, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) structDefinition(value *runtime.StructDefinitionValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutStructDefinition, semanticabi.TagKindStructDefinition)
	if err != nil || seen {
		return cell, err
	}
	result := fields(semanticabi.LayoutStructDefinition)
	_ = heapmodel.SetScalar(semanticabi.LayoutStructDefinition, result, "definition", e.metadata(value))
	_ = heapmodel.SetBytes(semanticabi.LayoutStructDefinition, result, "field_names", []byte(strings.Join(sortedKeys(value.NamedFieldIndices), "\x00")))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) typeRef(value *runtime.TypeRefValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutTypeRef, semanticabi.TagKindTypeRef)
	if err != nil || seen {
		return cell, err
	}
	result := fields(semanticabi.LayoutTypeRef)
	_ = heapmodel.SetBytes(semanticabi.LayoutTypeRef, result, "name", []byte(value.TypeName))
	_ = heapmodel.SetScalar(semanticabi.LayoutTypeRef, result, "type_arguments", e.metadata(value.TypeArgs))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) structInstance(value *runtime.StructInstanceValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutStructInstance, semanticabi.TagKindStructInstance)
	if err != nil || seen {
		return cell, err
	}
	definition, err := e.structDefinition(value.Definition)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	names := sortedKeys(value.Fields)
	flat := make([]runtime.Value, 0, len(names)+len(value.Positional))
	for _, name := range names {
		flat = append(flat, value.Fields[name])
	}
	flat = append(flat, value.Positional...)
	cells, err := e.values(flat)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutStructInstance)
	_ = heapmodel.SetObjects(semanticabi.LayoutStructInstance, result, "definition", semanticabi.ObjectID(definition.Payload))
	_ = heapmodel.SetCells(semanticabi.LayoutStructInstance, result, "fields", cells...)
	meta := structInstanceMeta{names, len(value.Positional), value.TypeArguments, value.Native}
	_ = heapmodel.SetScalar(semanticabi.LayoutStructInstance, result, "type_arguments", e.metadata(meta))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) values(values []runtime.Value) ([]semanticabi.Cell, error) {
	result := make([]semanticabi.Cell, len(values))
	for index, value := range values {
		cell, err := e.value(value)
		if err != nil {
			return nil, err
		}
		result[index] = cell
	}
	return result, nil
}

func mustObject(cell semanticabi.Cell) semanticabi.ObjectID {
	return semanticabi.ObjectID(cell.Payload)
}

func unexpected(name string) error {
	return fmt.Errorf("semanticabi Go binding: invalid %s metadata", name)
}
