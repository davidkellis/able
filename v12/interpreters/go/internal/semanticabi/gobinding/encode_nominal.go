package gobinding

import (
	"strings"

	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/internal/semanticabi/heapmodel"
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (e *encoder) interfaceDefinition(value *runtime.InterfaceDefinitionValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutInterfaceDefinition, semanticabi.TagKindInterfaceDefinition)
	if err != nil || seen {
		return cell, err
	}
	env, err := e.environment(value.Env)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutInterfaceDefinition)
	_ = heapmodel.SetScalar(semanticabi.LayoutInterfaceDefinition, result, "definition", e.metadata(value.Node))
	_ = heapmodel.SetObjects(semanticabi.LayoutInterfaceDefinition, result, "environment", env)
	_ = heapmodel.SetBytes(semanticabi.LayoutInterfaceDefinition, result, "qualified_name", []byte(value.QualifiedName))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) interfaceValue(value *runtime.InterfaceValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutInterfaceValue, semanticabi.TagKindInterfaceValue)
	if err != nil || seen {
		return cell, err
	}
	definition, err := e.interfaceDefinition(value.Interface)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	underlying, err := e.value(value.Underlying)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	names, shared := sortedKeys(value.Methods), sortedKeys(value.SharedMethods)
	methodValues := make([]runtime.Value, 0, len(names)+len(shared)+1)
	for _, name := range names {
		methodValues = append(methodValues, value.Methods[name])
	}
	for _, name := range shared {
		methodValues = append(methodValues, value.SharedMethods[name])
	}
	if value.BoundMethod != nil {
		methodValues = append(methodValues, value.BoundMethod)
	}
	methodCells, err := e.values(methodValues)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutInterfaceValue)
	_ = heapmodel.SetObjects(semanticabi.LayoutInterfaceValue, result, "interface", mustObject(definition))
	_ = heapmodel.SetCells(semanticabi.LayoutInterfaceValue, result, "underlying", underlying)
	_ = heapmodel.SetCells(semanticabi.LayoutInterfaceValue, result, "methods", methodCells...)
	meta := interfaceValueMeta{names, shared, value.BoundMethodName, value.InterfaceArgs}
	_ = heapmodel.SetScalar(semanticabi.LayoutInterfaceValue, result, "type_arguments", e.metadata(meta))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) unionDefinition(value *runtime.UnionDefinitionValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutUnionDefinition, semanticabi.TagKindUnionDefinition)
	if err != nil || seen {
		return cell, err
	}
	result := fields(semanticabi.LayoutUnionDefinition)
	_ = heapmodel.SetScalar(semanticabi.LayoutUnionDefinition, result, "definition", e.metadata(value.Node))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) packageValue(value *runtime.PackageValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutPackage, semanticabi.TagKindPackage)
	if err != nil || seen {
		return cell, err
	}
	names := sortedKeys(value.Public)
	vals := make([]runtime.Value, 0, len(names))
	for _, name := range names {
		vals = append(vals, value.Public[name])
	}
	cells, err := e.values(vals)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutPackage)
	_ = heapmodel.SetBytes(semanticabi.LayoutPackage, result, "name", []byte(value.Name))
	_ = heapmodel.SetBytes(semanticabi.LayoutPackage, result, "name_path", []byte(strings.Join(value.NamePath, "\x00")))
	_ = heapmodel.SetBytes(semanticabi.LayoutPackage, result, "identity", []byte(value.IdentityKey))
	if value.IsPrivate {
		_ = heapmodel.SetScalar(semanticabi.LayoutPackage, result, "flags", 1)
	}
	_ = heapmodel.SetBytes(semanticabi.LayoutPackage, result, "public_names", []byte(strings.Join(names, "\x00")))
	_ = heapmodel.SetCells(semanticabi.LayoutPackage, result, "public_bindings", cells...)
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) dynPackage(value *runtime.DynPackageValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutDynPackage, semanticabi.TagKindDynPackage)
	if err != nil || seen {
		return cell, err
	}
	result := fields(semanticabi.LayoutDynPackage)
	_ = heapmodel.SetBytes(semanticabi.LayoutDynPackage, result, "name", []byte(value.Name))
	_ = heapmodel.SetBytes(semanticabi.LayoutDynPackage, result, "name_path", []byte(strings.Join(value.NamePath, "\x00")))
	_ = heapmodel.SetBytes(semanticabi.LayoutDynPackage, result, "identity", []byte(value.IdentityKey))
	if value.IsPrivate {
		_ = heapmodel.SetScalar(semanticabi.LayoutDynPackage, result, "flags", 1)
	}
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) dynRef(value *runtime.DynRefValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutDynRef, semanticabi.TagKindDynRef)
	if err != nil || seen {
		return cell, err
	}
	result := fields(semanticabi.LayoutDynRef)
	_ = heapmodel.SetBytes(semanticabi.LayoutDynRef, result, "package", []byte(value.Package))
	_ = heapmodel.SetBytes(semanticabi.LayoutDynRef, result, "name", []byte(value.Name))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) errorValue(value *runtime.ErrorValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutError, semanticabi.TagKindError)
	if err != nil || seen {
		return cell, err
	}
	names := sortedKeys(value.Payload)
	vals := make([]runtime.Value, 0, len(names))
	for _, name := range names {
		vals = append(vals, value.Payload[name])
	}
	cells, err := e.values(vals)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutError)
	_ = heapmodel.SetScalar(semanticabi.LayoutError, result, "type", e.metadata(errorMeta{value.TypeName, names}))
	_ = heapmodel.SetCells(semanticabi.LayoutError, result, "payload", cells...)
	_ = heapmodel.SetBytes(semanticabi.LayoutError, result, "message", []byte(value.Message))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) boundMethod(value *runtime.BoundMethodValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutBoundMethod, semanticabi.TagKindBoundMethod)
	if err != nil || seen {
		return cell, err
	}
	receiver, err := e.value(value.Receiver)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	method, err := e.value(value.Method)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutBoundMethod)
	_ = heapmodel.SetCells(semanticabi.LayoutBoundMethod, result, "receiver", receiver)
	_ = heapmodel.SetCells(semanticabi.LayoutBoundMethod, result, "method", method)
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) implementationNamespace(value *runtime.ImplementationNamespaceValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutImplementationNamespace, semanticabi.TagKindImplementationNamespace)
	if err != nil || seen {
		return cell, err
	}
	names := sortedKeys(value.Methods)
	vals := make([]runtime.Value, 0, len(names))
	for _, name := range names {
		vals = append(vals, value.Methods[name])
	}
	cells, err := e.values(vals)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutImplementationNamespace)
	meta := struct {
		Name, Interface *ast.Identifier
		Target          ast.TypeExpression
		Private         bool
		Methods         []string
	}{value.Name, value.InterfaceName, value.TargetType, value.IsPrivate, names}
	_ = heapmodel.SetScalar(semanticabi.LayoutImplementationNamespace, result, "definition", e.metadata(meta))
	_ = heapmodel.SetCells(semanticabi.LayoutImplementationNamespace, result, "methods", cells...)
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func (e *encoder) partialFunction(value *runtime.PartialFunctionValue) (semanticabi.Cell, error) {
	id, cell, seen, err := e.reserve(value, semanticabi.LayoutPartialFunction, semanticabi.TagKindPartialFunction)
	if err != nil || seen {
		return cell, err
	}
	target, err := e.value(value.Target)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	args, err := e.values(value.BoundArgs)
	if err != nil {
		return semanticabi.Cell{}, err
	}
	result := fields(semanticabi.LayoutPartialFunction)
	_ = heapmodel.SetCells(semanticabi.LayoutPartialFunction, result, "target", target)
	_ = heapmodel.SetCells(semanticabi.LayoutPartialFunction, result, "bound_arguments", args...)
	_ = heapmodel.SetScalar(semanticabi.LayoutPartialFunction, result, "call", e.metadata(value.Call))
	return cell, e.snapshot.Heap.Initialize(id, result)
}

func splitNames(value []byte) []string {
	if len(value) == 0 {
		return nil
	}
	return strings.Split(string(value), "\x00")
}
