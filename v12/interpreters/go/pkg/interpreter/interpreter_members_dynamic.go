package interpreter

import (
	"fmt"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) packageMemberAccess(pkg runtime.PackageValue, member ast.Expression) (runtime.Value, error) {
	memberName, err := memberIdentifierName(member, "Package member access expects identifier")
	if err != nil {
		return nil, err
	}
	return i.packageMemberAccessName(pkg, memberName)
}

func (i *Interpreter) dynPackageMemberAccess(pkg runtime.DynPackageValue, member ast.Expression) (runtime.Value, error) {
	memberName, err := memberIdentifierName(member, "Dyn package member access expects identifier")
	if err != nil {
		return nil, err
	}
	return i.dynPackageMemberAccessName(pkg, memberName)
}

func (i *Interpreter) implNamespaceMember(ns runtime.ImplementationNamespaceValue, member ast.Expression) (runtime.Value, error) {
	memberName, err := memberIdentifierName(member, "Impl namespace member access expects identifier")
	if err != nil {
		return nil, err
	}
	return i.implNamespaceMemberName(ns, memberName)
}

func (i *Interpreter) resolveInterfaceMethodCallable(val *runtime.InterfaceValue, memberName string) (runtime.Value, bool, error) {
	if val == nil {
		return nil, false, fmt.Errorf("Interface value is nil")
	}
	if memberName == "" {
		return nil, false, fmt.Errorf("Interface member access expects identifier")
	}
	if method, ok := interfaceValueLookupMethod(val, memberName); ok {
		return method, true, nil
	}
	ifaceName := interfaceDefinitionIdentity(val.Interface)
	if ifaceName == "" {
		return nil, false, fmt.Errorf("Unknown interface for member access")
	}
	var method runtime.Value
	if info, ok := i.getTypeInfoForValue(val.Underlying); ok {
		resolved, err := i.findMethod(info, memberName, ifaceName, val.InterfaceArgs)
		if err != nil {
			return nil, false, err
		}
		method = resolved
	}
	// In compiled no-bootstrap mode, fall back to compiled dispatch for inherent methods
	// that aren't in the interface definition (e.g., Iterator.collect).
	if method == nil && i.compiledInstanceMethodFn != nil {
		if info, ok := i.getTypeInfoForValue(val.Underlying); ok {
			if resolved, found := i.compiledInstanceMethodFn(info.name, memberName); found && resolved != nil {
				method = resolved
			}
		}
	}
	if method == nil && i.interfaceMethodResolver != nil {
		if resolved, found := i.interfaceMethodResolver(val.Underlying, ifaceName, memberName); found && resolved != nil {
			// interfaceMethodResolver returns arity+1 (includes self). Call sites that
			// inject the receiver need the explicit-arg arity instead.
			if native, ok := resolved.(*runtime.NativeFunctionValue); ok && native.Arity > 0 {
				adjusted := *native
				adjusted.Arity = native.Arity - 1
				method = &adjusted
			} else {
				method = resolved
			}
		}
	}
	if method == nil && i.compiledInterfaceMemberFn != nil {
		if resolved, found := i.compiledInterfaceMemberFn(val.Underlying, memberName); found && resolved != nil {
			method = resolved
		}
	}
	// Fall back to IteratorValue native member dispatch (handles next, filter, etc.)
	if method == nil {
		if _, ok := val.Underlying.(*runtime.IteratorValue); ok {
			if memberName == "next" {
				method = iteratorNextNativeMethod()
			} else if memberName == "close" {
				method = iteratorCloseNativeMethod()
			} else if ifaceDef := i.interfaces["Iterator"]; ifaceDef != nil {
				resolved, found, err := i.iteratorInterfaceMethodValue(ifaceDef, memberName)
				if err != nil {
					return nil, false, err
				}
				if found {
					method = resolved
				}
			}
		}
	}
	// Fall back to default interface method implementations (methods with DefaultImpl in the signature).
	if method == nil && val.Interface != nil && val.Interface.Node != nil {
		for _, sig := range val.Interface.Node.Signatures {
			if sig == nil || sig.Name == nil || sig.Name.Name != memberName || sig.DefaultImpl == nil {
				continue
			}
			resolved, found, err := i.interfaceDefaultMethodValue(val.Interface, memberName)
			if err != nil {
				return nil, false, err
			}
			if found {
				method = resolved
			}
			break
		}
	}
	if method == nil {
		return nil, false, fmt.Errorf("No method '%s' for interface %s", memberName, ifaceName)
	}
	if fn := firstFunction(method); fn != nil {
		if fnDef, ok := fn.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, false, fmt.Errorf("Method '%s' on %s is private", memberName, ifaceName)
		}
	}
	return method, true, nil
}

func (i *Interpreter) interfaceMember(val *runtime.InterfaceValue, member ast.Expression) (runtime.Value, error) {
	if val == nil {
		return nil, fmt.Errorf("Interface value is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Interface member access expects identifier")
	}
	ifaceName := interfaceDefinitionIdentity(val.Interface)
	if ifaceName == "" {
		return nil, fmt.Errorf("Unknown interface for member access")
	}
	if method, ok := interfaceValueLookupBoundMethod(val, ident.Name); ok {
		return method, nil
	}
	method, found, err := i.resolveInterfaceMethodCallable(val, ident.Name)
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, fmt.Errorf("No method '%s' for interface %s", ident.Name, ifaceName)
	}
	if interfaceValueMethodIsBound(method) {
		interfaceValueSetBoundMethod(val, ident.Name, method)
		return method, nil
	}
	receiver := interfaceMethodReceiver(i, val, method)
	bound, err := bindResolvedMethodValue(receiver, method, ident.Name)
	if err != nil {
		return nil, err
	}
	interfaceValueSetBoundMethod(val, ident.Name, bound)
	return bound, nil
}

func bindResolvedMethodValue(receiver runtime.Value, method runtime.Value, memberName string) (runtime.Value, error) {
	switch fn := method.(type) {
	case runtime.NativeFunctionValue:
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: fn}, nil
	case *runtime.NativeFunctionValue:
		if fn == nil {
			return nil, fmt.Errorf("native method '%s' is nil", memberName)
		}
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: *fn}, nil
	case runtime.NativeBoundMethodValue:
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: fn.Method}, nil
	case *runtime.NativeBoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("native method '%s' is nil", memberName)
		}
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: fn.Method}, nil
	case runtime.BoundMethodValue:
		return runtime.BoundMethodValue{Receiver: receiver, Method: fn.Method}, nil
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("method '%s' is nil", memberName)
		}
		return runtime.BoundMethodValue{Receiver: receiver, Method: fn.Method}, nil
	default:
		return runtime.BoundMethodValue{Receiver: receiver, Method: method}, nil
	}
}

func interfaceMethodReceiver(i *Interpreter, val *runtime.InterfaceValue, method runtime.Value) runtime.Value {
	_ = i
	if val == nil {
		return nil
	}
	switch bound := method.(type) {
	case runtime.BoundMethodValue:
		receiver := unwrapInterfaceMethodReceiver(bound.Receiver)
		if receiver == nil {
			return runtime.NilValue{}
		}
		return receiver
	case *runtime.BoundMethodValue:
		if bound != nil {
			receiver := unwrapInterfaceMethodReceiver(bound.Receiver)
			if receiver == nil {
				return runtime.NilValue{}
			}
			return receiver
		}
	}
	receiver := unwrapInterfaceMethodReceiver(val.Underlying)
	if receiver == nil {
		return runtime.NilValue{}
	}
	return receiver
}

func unwrapInterfaceMethodReceiver(val runtime.Value) runtime.Value {
	for {
		switch iface := val.(type) {
		case runtime.InterfaceValue:
			val = iface.Underlying
			continue
		case *runtime.InterfaceValue:
			if iface != nil {
				val = iface.Underlying
				continue
			}
		}
		break
	}
	return val
}

func (i *Interpreter) resolveDynRef(ref runtime.DynRefValue) (runtime.Value, error) {
	bucket, ok := i.packageRegistry[ref.Package]
	if !ok {
		return nil, fmt.Errorf("dyn ref '%s.%s' not found", ref.Package, ref.Name)
	}
	val, ok := bucket[ref.Name]
	if !ok {
		return nil, fmt.Errorf("dyn ref '%s.%s' not found", ref.Package, ref.Name)
	}
	if isPrivateSymbol(val) {
		return nil, fmt.Errorf("dyn ref '%s.%s' is private", ref.Package, ref.Name)
	}
	if runtime.IsFunctionLike(val) {
		return val, nil
	}
	return nil, fmt.Errorf("dyn ref '%s.%s' is not callable", ref.Package, ref.Name)
}

func toStructDefinitionValue(val runtime.Value, name string) (*runtime.StructDefinitionValue, error) {
	switch v := val.(type) {
	case *runtime.StructDefinitionValue:
		return v, nil
	case runtime.StructDefinitionValue:
		return &v, nil
	default:
		return nil, fmt.Errorf("'%s' is not a struct type", name)
	}
}
