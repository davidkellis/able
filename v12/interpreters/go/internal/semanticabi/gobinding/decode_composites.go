package gobinding

import (
	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (d *decoder) array(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutArray)
	if err != nil {
		return nil, err
	}
	value := &runtime.ArrayValue{}
	d.objects[id] = value
	value.Elements, err = d.values(field(object, "elements").Cells)
	return value, err
}

func (d *decoder) hashMap(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutHashMap)
	if err != nil {
		return nil, err
	}
	value := &runtime.HashMapValue{}
	d.objects[id] = value
	flat, err := d.values(field(object, "entries").Cells)
	if err != nil {
		return nil, err
	}
	meta, err := d.metadata(field(object, "hashes").Scalar)
	if err != nil {
		return nil, err
	}
	hashes, ok := meta.([]uint64)
	if !ok || len(flat)%2 != 0 || len(hashes) != len(flat)/2 {
		return nil, unexpected("hash map")
	}
	for index := 0; index < len(flat); index += 2 {
		value.AppendEntry(runtime.HashMapEntry{Key: flat[index], Value: flat[index+1], Hash: hashes[index/2]})
	}
	return value, nil
}

func (d *decoder) hasher(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutHasher)
	if err != nil {
		return nil, err
	}
	value := runtime.NewHasherValueFromState(field(object, "state").Scalar)
	d.objects[id] = value
	return value, nil
}

func (d *decoder) iterator(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutIterator)
	if err != nil {
		return nil, err
	}
	value := runtime.NewIteratorValueFromHostDriver(runtime.IteratorHostDriver{}, false)
	d.objects[id] = value
	driverCell := field(object, "driver").Cells[0]
	hostID := semanticabi.ObjectID(driverCell.Payload)
	host, err := d.snapshot.Heap.Hosts().Resolve(hostID)
	if err != nil {
		return nil, err
	}
	if driverCell.Tag != semanticabi.TagKindHostHandle {
		return nil, unexpected("iterator driver tag")
	}
	meta, ok := d.snapshot.hostMetadata[hostID].(iteratorDriverMeta)
	if !ok {
		return nil, unexpected("iterator driver")
	}
	retained, err := d.values(host.Retained)
	if err != nil {
		return nil, err
	}
	value.RestoreHostDriver(runtime.IteratorHostDriver{Next: meta.Next, NextRaw: meta.NextRaw, Finalize: meta.Finalize, Retained: retained}, field(object, "closed").Scalar != 0)
	return value, nil
}

func (d *decoder) function(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutFunction)
	if err != nil {
		return nil, err
	}
	value := &runtime.FunctionValue{}
	d.objects[id] = value
	metaRaw, err := d.metadata(field(object, "declaration").Scalar)
	if err != nil {
		return nil, err
	}
	meta, ok := metaRaw.(functionMeta)
	if !ok {
		return nil, unexpected("function")
	}
	value.Declaration, value.MethodPriority, value.TypeQualified, value.MethodSet, value.Bytecode = meta.Declaration, meta.MethodPriority, meta.TypeQualified, meta.MethodSet, meta.Bytecode
	value.Closure, err = d.environment(field(object, "environment").Objects[0])
	return value, err
}

func (d *decoder) functionOverload(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutFunctionOverload)
	if err != nil {
		return nil, err
	}
	value := &runtime.FunctionOverloadValue{}
	d.objects[id] = value
	for _, ref := range field(object, "functions").Objects {
		decoded, err := d.value(semanticabi.Cell{Tag: semanticabi.TagKindFunction, Payload: uint64(ref)})
		if err != nil {
			return nil, err
		}
		fn, ok := decoded.(*runtime.FunctionValue)
		if !ok {
			return nil, unexpected("function overload")
		}
		value.Overloads = append(value.Overloads, fn)
	}
	return value, nil
}

func (d *decoder) environment(id semanticabi.ObjectID) (*runtime.Environment, error) {
	if id == 0 {
		return nil, nil
	}
	if value, ok := d.envs[id]; ok {
		return value, nil
	}
	object, err := d.object(id, semanticabi.LayoutEnvironment)
	if err != nil {
		return nil, err
	}
	value := runtime.NewEnvironment(nil)
	d.envs[id] = value
	parent, err := d.environment(field(object, "parent").Objects[0])
	if err != nil {
		return nil, err
	}
	// Parent cannot be assigned after construction. Recreate before bindings and
	// replace the memo entry; cycles through values still see the final pointer
	// because parent cycles are not legal lexical environments.
	value = runtime.NewEnvironment(parent)
	d.envs[id] = value
	for _, bindingID := range field(object, "bindings").Objects {
		binding, err := d.object(bindingID, semanticabi.LayoutBindingCell)
		if err != nil {
			return nil, err
		}
		name := string(field(binding, "name").Bytes)
		decoded, err := d.value(field(binding, "value").Cells[0])
		if err != nil {
			return nil, err
		}
		value.DefineWithoutMerge(name, decoded)
	}
	return value, nil
}

func (d *decoder) structDefinition(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutStructDefinition)
	if err != nil {
		return nil, err
	}
	meta, err := d.metadata(field(object, "definition").Scalar)
	if err != nil {
		return nil, err
	}
	original, ok := meta.(*runtime.StructDefinitionValue)
	if !ok {
		return nil, unexpected("struct definition")
	}
	value := &runtime.StructDefinitionValue{Node: original.Node, NamedFieldIndices: make(map[string]int, len(original.NamedFieldIndices))}
	for name, index := range original.NamedFieldIndices {
		value.NamedFieldIndices[name] = index
	}
	d.objects[id] = value
	return value, nil
}

func (d *decoder) typeRef(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutTypeRef)
	if err != nil {
		return nil, err
	}
	meta, err := d.metadata(field(object, "type_arguments").Scalar)
	if err != nil {
		return nil, err
	}
	args, ok := meta.([]ast.TypeExpression)
	if !ok {
		return nil, unexpected("type ref")
	}
	value := &runtime.TypeRefValue{TypeName: string(field(object, "name").Bytes), TypeArgs: append([]ast.TypeExpression(nil), args...)}
	d.objects[id] = value
	return value, nil
}

func (d *decoder) structInstance(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutStructInstance)
	if err != nil {
		return nil, err
	}
	value := &runtime.StructInstanceValue{}
	d.objects[id] = value
	definition, err := d.structDefinition(field(object, "definition").Objects[0])
	if err != nil {
		return nil, err
	}
	value.Definition = definition.(*runtime.StructDefinitionValue)
	metaRaw, err := d.metadata(field(object, "type_arguments").Scalar)
	if err != nil {
		return nil, err
	}
	meta, ok := metaRaw.(structInstanceMeta)
	if !ok {
		return nil, unexpected("struct instance")
	}
	flat, err := d.values(field(object, "fields").Cells)
	if err != nil {
		return nil, err
	}
	if len(flat) != len(meta.Names)+meta.Positional {
		return nil, unexpected("struct instance fields")
	}
	value.Fields = make(map[string]runtime.Value, len(meta.Names))
	for index, name := range meta.Names {
		value.Fields[name] = flat[index]
	}
	value.Positional = append([]runtime.Value(nil), flat[len(meta.Names):]...)
	value.TypeArguments, value.Native = append([]ast.TypeExpression(nil), meta.TypeArguments...), meta.Native
	return value, nil
}

func (d *decoder) values(cells []semanticabi.Cell) ([]runtime.Value, error) {
	result := make([]runtime.Value, len(cells))
	for index, cell := range cells {
		value, err := d.value(cell)
		if err != nil {
			return nil, err
		}
		result[index] = value
	}
	return result, nil
}
