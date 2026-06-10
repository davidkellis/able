package gobinding

import (
	"able/interpreter-go/internal/semanticabi"
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (d *decoder) interfaceDefinition(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutInterfaceDefinition)
	if err != nil {
		return nil, err
	}
	meta, err := d.metadata(field(object, "definition").Scalar)
	if err != nil {
		return nil, err
	}
	node, ok := meta.(*ast.InterfaceDefinition)
	if !ok && meta != nil {
		return nil, unexpected("interface definition")
	}
	value := &runtime.InterfaceDefinitionValue{Node: node, QualifiedName: string(field(object, "qualified_name").Bytes)}
	d.objects[id] = value
	value.Env, err = d.environment(field(object, "environment").Objects[0])
	return value, err
}

func (d *decoder) interfaceValue(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutInterfaceValue)
	if err != nil {
		return nil, err
	}
	value := &runtime.InterfaceValue{}
	d.objects[id] = value
	definition, err := d.interfaceDefinition(field(object, "interface").Objects[0])
	if err != nil {
		return nil, err
	}
	value.Interface = definition.(*runtime.InterfaceDefinitionValue)
	value.Underlying, err = d.value(field(object, "underlying").Cells[0])
	if err != nil {
		return nil, err
	}
	metaRaw, err := d.metadata(field(object, "type_arguments").Scalar)
	if err != nil {
		return nil, err
	}
	meta, ok := metaRaw.(interfaceValueMeta)
	if !ok {
		return nil, unexpected("interface value")
	}
	methods, err := d.values(field(object, "methods").Cells)
	if err != nil {
		return nil, err
	}
	needed := len(meta.MethodNames) + len(meta.SharedMethodNames)
	if len(methods) < needed || len(methods) > needed+1 {
		return nil, unexpected("interface methods")
	}
	value.Methods = make(map[string]runtime.Value, len(meta.MethodNames))
	value.SharedMethods = make(map[string]runtime.Value, len(meta.SharedMethodNames))
	position := 0
	for _, name := range meta.MethodNames {
		value.Methods[name] = methods[position]
		position++
	}
	for _, name := range meta.SharedMethodNames {
		value.SharedMethods[name] = methods[position]
		position++
	}
	if position < len(methods) {
		value.BoundMethod = methods[position]
	}
	value.BoundMethodName, value.InterfaceArgs = meta.BoundMethodName, append([]ast.TypeExpression(nil), meta.InterfaceArgs...)
	return value, nil
}

func (d *decoder) unionDefinition(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutUnionDefinition)
	if err != nil {
		return nil, err
	}
	meta, err := d.metadata(field(object, "definition").Scalar)
	if err != nil {
		return nil, err
	}
	node, ok := meta.(*ast.UnionDefinition)
	if !ok && meta != nil {
		return nil, unexpected("union definition")
	}
	value := &runtime.UnionDefinitionValue{Node: node}
	d.objects[id] = value
	return value, nil
}

func (d *decoder) packageValue(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutPackage)
	if err != nil {
		return nil, err
	}
	names := splitNames(field(object, "public_names").Bytes)
	value := &runtime.PackageValue{Name: string(field(object, "name").Bytes), IdentityKey: string(field(object, "identity").Bytes), NamePath: splitNames(field(object, "name_path").Bytes), IsPrivate: field(object, "flags").Scalar&1 != 0, Public: make(map[string]runtime.Value)}
	d.objects[id] = value
	values, err := d.values(field(object, "public_bindings").Cells)
	if err != nil {
		return nil, err
	}
	if len(values) != len(names) {
		return nil, unexpected("package bindings")
	}
	value.Public = make(map[string]runtime.Value, len(values))
	for index, name := range names {
		value.Public[name] = values[index]
	}
	return value, nil
}

func (d *decoder) dynPackage(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutDynPackage)
	if err != nil {
		return nil, err
	}
	value := runtime.DynPackageValue{Name: string(field(object, "name").Bytes), NamePath: splitNames(field(object, "name_path").Bytes), IdentityKey: string(field(object, "identity").Bytes), IsPrivate: field(object, "flags").Scalar&1 != 0}
	d.objects[id] = &value
	return &value, nil
}

func (d *decoder) dynRef(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutDynRef)
	if err != nil {
		return nil, err
	}
	value := &runtime.DynRefValue{Package: string(field(object, "package").Bytes), Name: string(field(object, "name").Bytes)}
	d.objects[id] = value
	return value, nil
}

func (d *decoder) errorValue(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutError)
	if err != nil {
		return nil, err
	}
	metaRaw, err := d.metadata(field(object, "type").Scalar)
	if err != nil {
		return nil, err
	}
	meta, ok := metaRaw.(errorMeta)
	if !ok {
		return nil, unexpected("error")
	}
	value := &runtime.ErrorValue{TypeName: meta.TypeName, Message: string(field(object, "message").Bytes), Payload: make(map[string]runtime.Value)}
	d.objects[id] = value
	values, err := d.values(field(object, "payload").Cells)
	if err != nil {
		return nil, err
	}
	if len(values) != len(meta.Names) {
		return nil, unexpected("error payload")
	}
	value.Payload = make(map[string]runtime.Value, len(values))
	for index, name := range meta.Names {
		value.Payload[name] = values[index]
	}
	return value, nil
}

func (d *decoder) boundMethod(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutBoundMethod)
	if err != nil {
		return nil, err
	}
	value := &runtime.BoundMethodValue{}
	d.objects[id] = value
	value.Receiver, err = d.value(field(object, "receiver").Cells[0])
	if err != nil {
		return nil, err
	}
	value.Method, err = d.value(field(object, "method").Cells[0])
	return value, err
}

func (d *decoder) implementationNamespace(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutImplementationNamespace)
	if err != nil {
		return nil, err
	}
	metaRaw, err := d.metadata(field(object, "definition").Scalar)
	if err != nil {
		return nil, err
	}
	meta, ok := metaRaw.(struct {
		Name, Interface *ast.Identifier
		Target          ast.TypeExpression
		Private         bool
		Methods         []string
	})
	if !ok {
		return nil, unexpected("implementation namespace")
	}
	methods, err := d.values(field(object, "methods").Cells)
	if err != nil {
		return nil, err
	}
	value := &runtime.ImplementationNamespaceValue{Name: meta.Name, InterfaceName: meta.Interface, TargetType: meta.Target, IsPrivate: meta.Private, Methods: make(map[string]runtime.Value, len(methods))}
	d.objects[id] = value
	for index, name := range meta.Methods {
		value.Methods[name] = methods[index]
	}
	return value, nil
}

func (d *decoder) partialFunction(id semanticabi.ObjectID) (runtime.Value, error) {
	object, err := d.object(id, semanticabi.LayoutPartialFunction)
	if err != nil {
		return nil, err
	}
	value := &runtime.PartialFunctionValue{}
	d.objects[id] = value
	value.Target, err = d.value(field(object, "target").Cells[0])
	if err != nil {
		return nil, err
	}
	value.BoundArgs, err = d.values(field(object, "bound_arguments").Cells)
	if err != nil {
		return nil, err
	}
	meta, err := d.metadata(field(object, "call").Scalar)
	if err != nil {
		return nil, err
	}
	value.Call, _ = meta.(*ast.FunctionCall)
	return value, nil
}
