package interpreter

import (
	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func interfaceSignatureReturnsSelf(signature *ast.FunctionSignature) bool {
	if signature == nil || signature.ReturnType == nil {
		return false
	}
	simple, ok := signature.ReturnType.(*ast.SimpleTypeExpression)
	return ok && simple != nil && simple.Name != nil && simple.Name.Name == "Self"
}

func (i *Interpreter) interfaceMethodReturnsSelf(value *runtime.InterfaceValue, methodName string) bool {
	if i == nil || value == nil || methodName == "" {
		return false
	}
	if value.Interface == nil || value.Interface.Node == nil {
		return false
	}
	return i.interfaceDefinitionMethodReturnsSelf(value.Interface, methodName, nil)
}

func (i *Interpreter) interfaceDefinitionMethodReturnsSelf(definition *runtime.InterfaceDefinitionValue, methodName string, visited map[string]struct{}) bool {
	if i == nil || definition == nil || definition.Node == nil || methodName == "" {
		return false
	}
	for _, signature := range definition.Node.Signatures {
		if signature != nil && signature.Name != nil && signature.Name.Name == methodName {
			return interfaceSignatureReturnsSelf(signature)
		}
	}
	if len(definition.Node.BaseInterfaces) == 0 {
		return false
	}
	if visited == nil {
		visited = make(map[string]struct{}, len(definition.Node.BaseInterfaces)+1)
	}
	identity := interfaceDefinitionIdentity(definition)
	if _, seen := visited[identity]; seen {
		return false
	}
	visited[identity] = struct{}{}
	for _, base := range definition.Node.BaseInterfaces {
		info, ok := parseTypeExpression(base)
		if !ok || info.name == "" {
			continue
		}
		baseDefinition := i.interfaces[i.canonicalInterfaceName(info.name)]
		if i.interfaceDefinitionMethodReturnsSelf(baseDefinition, methodName, visited) {
			return true
		}
	}
	return false
}

func cloneInterfaceMethodOverlay(methods map[string]runtime.Value) map[string]runtime.Value {
	if len(methods) == 0 {
		return nil
	}
	cloned := make(map[string]runtime.Value, len(methods))
	for name, method := range methods {
		cloned[name] = method
	}
	return cloned
}

func preserveInterfaceSelfReturn(value *runtime.InterfaceValue, result runtime.Value) runtime.Value {
	if value == nil {
		return result
	}
	result = unwrapInterfaceMethodReceiver(result)
	return &runtime.InterfaceValue{
		Interface:     value.Interface,
		Underlying:    result,
		Methods:       cloneInterfaceMethodOverlay(value.Methods),
		SharedMethods: value.SharedMethods,
		InterfaceArgs: value.InterfaceArgs,
	}
}

func (i *Interpreter) preserveInterfaceMethodSelfReturn(receiver runtime.Value, methodName string, result runtime.Value, err error) (runtime.Value, error) {
	if err != nil {
		return nil, err
	}
	value, ok := receiver.(*runtime.InterfaceValue)
	if !ok || !i.interfaceMethodReturnsSelf(value, methodName) {
		return result, nil
	}
	return preserveInterfaceSelfReturn(value, result), nil
}
