package interpreter

import (
	"errors"
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func memberIdentifierName(member ast.Expression, errMsg string) (string, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return "", errors.New(errMsg)
	}
	return ident.Name, nil
}

func (i *Interpreter) structDefinitionMemberName(def *runtime.StructDefinitionValue, memberName string) (runtime.Value, error) {
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("struct definition missing identifier")
	}
	typeName := def.Node.ID.Name
	bucket := i.inherentMethods[typeName]
	var method runtime.Value
	var found bool
	if bucket != nil {
		method, found = bucket[memberName]
	}
	if !found {
		candidate, err := i.findMethodCached(typeInfo{name: typeName}, memberName, "")
		if err != nil {
			return nil, err
		}
		method = candidate
	}
	if method == nil {
		return nil, fmt.Errorf("No static method '%s' for %s", memberName, typeName)
	}
	if fn := firstFunction(method); fn != nil {
		if fnDef, ok := fn.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, fmt.Errorf("Method '%s' on %s is private", memberName, typeName)
		}
	}
	return method, nil
}

func (i *Interpreter) interfaceDefinitionMemberName(def *runtime.InterfaceDefinitionValue, memberName string) (runtime.Value, error) {
	if def == nil || def.Node == nil || def.Node.ID == nil {
		return nil, fmt.Errorf("interface definition missing identifier")
	}
	ifaceName := def.Node.ID.Name
	var sig *ast.FunctionSignature
	for _, candidate := range def.Node.Signatures {
		if candidate == nil || candidate.Name == nil || candidate.Name.Name != memberName {
			continue
		}
		sig = candidate
		break
	}
	if sig == nil {
		return nil, fmt.Errorf("No method '%s' for interface %s", memberName, ifaceName)
	}
	arity := len(sig.Params)
	fn := runtime.NativeFunctionValue{
		Name:  fmt.Sprintf("%s.%s", ifaceName, memberName),
		Arity: arity,
		Impl: func(ctx *runtime.NativeCallContext, args []runtime.Value) (runtime.Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("%s.%s requires a receiver", ifaceName, memberName)
			}
			receiver := unwrapInterfaceValue(args[0])
			method, err := i.resolveInterfaceMethod(receiver, ifaceName, memberName)
			if err != nil {
				return nil, err
			}
			if method == nil {
				return nil, fmt.Errorf("No method '%s' for interface %s", memberName, ifaceName)
			}
			callArgs := append([]runtime.Value{receiver}, args[1:]...)
			return i.callCallableValue(method, callArgs, ctx.Env, nil)
		},
	}
	return fn, nil
}

func (i *Interpreter) typeRefMemberName(ref runtime.TypeRefValue, memberName string) (runtime.Value, error) {
	typeName := ref.TypeName
	if typeName == "" {
		return nil, fmt.Errorf("type reference missing name")
	}
	bucket := i.inherentMethods[typeName]
	var method runtime.Value
	var found bool
	if bucket != nil {
		method, found = bucket[memberName]
	}
	if !found {
		candidate, err := i.findMethod(typeInfo{name: typeName, typeArgs: ref.TypeArgs}, memberName, "", nil)
		if err != nil {
			return nil, err
		}
		method = candidate
	}
	if method == nil {
		return nil, fmt.Errorf("No static method '%s' for %s", memberName, typeName)
	}
	if fn := firstFunction(method); fn != nil {
		if fnDef, ok := fn.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, fmt.Errorf("Method '%s' on %s is private", memberName, typeName)
		}
	}
	return method, nil
}

func (i *Interpreter) packageMemberAccessName(pkg runtime.PackageValue, memberName string) (runtime.Value, error) {
	if pkg.Public == nil {
		return nil, fmt.Errorf("Package has no public members")
	}
	val, ok := pkg.Public[memberName]
	if !ok {
		pkgName := pkg.Name
		if pkgName == "" {
			pkgName = strings.Join(pkg.NamePath, ".")
		}
		if pkgName == "" {
			pkgName = "<package>"
		}
		return nil, fmt.Errorf("No public member '%s' on package %s", memberName, pkgName)
	}
	return val, nil
}

func (i *Interpreter) dynPackageMemberAccessName(pkg runtime.DynPackageValue, memberName string) (runtime.Value, error) {
	if memberName == "def" {
		return runtime.NativeBoundMethodValue{Receiver: pkg, Method: i.dynPackageDefMethod}, nil
	}
	if memberName == "eval" {
		return runtime.NativeBoundMethodValue{Receiver: pkg, Method: i.dynPackageEvalMethod}, nil
	}
	pkgName := pkg.Name
	if pkgName == "" {
		pkgName = strings.Join(pkg.NamePath, ".")
	}
	if pkgName == "" {
		return nil, fmt.Errorf("Dyn package missing name")
	}
	bucket, ok := i.packageRegistry[pkgName]
	if !ok {
		return nil, fmt.Errorf("dyn package '%s' not found", pkgName)
	}
	sym, ok := bucket[memberName]
	if !ok {
		return nil, fmt.Errorf("dyn package '%s' has no member '%s'", pkgName, memberName)
	}
	if isPrivateSymbol(sym) {
		return nil, fmt.Errorf("dyn package '%s' member '%s' is private", pkgName, memberName)
	}
	return runtime.DynRefValue{Package: pkgName, Name: memberName}, nil
}

func (i *Interpreter) implNamespaceMemberName(ns runtime.ImplementationNamespaceValue, memberName string) (runtime.Value, error) {
	if ns.Methods == nil {
		return nil, fmt.Errorf("Impl namespace has no methods")
	}
	method, ok := ns.Methods[memberName]
	if !ok {
		name := "<impl>"
		if ns.Name != nil {
			name = ns.Name.Name
		}
		return nil, fmt.Errorf("No method '%s' on impl %s", memberName, name)
	}
	return method, nil
}
