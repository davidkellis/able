package interpreter

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
	"able/interpreter-go/pkg/runtime"
)

func (i *Interpreter) packageMemberAccess(pkg runtime.PackageValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Package member access expects identifier")
	}
	if pkg.Public == nil {
		return nil, fmt.Errorf("Package has no public members")
	}
	val, ok := pkg.Public[ident.Name]
	if !ok {
		pkgName := pkg.Name
		if pkgName == "" {
			pkgName = strings.Join(pkg.NamePath, ".")
		}
		if pkgName == "" {
			pkgName = "<package>"
		}
		return nil, fmt.Errorf("No public member '%s' on package %s", ident.Name, pkgName)
	}
	return val, nil
}

func (i *Interpreter) dynPackageMemberAccess(pkg runtime.DynPackageValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Dyn package member access expects identifier")
	}
	if ident.Name == "def" {
		return runtime.NativeBoundMethodValue{Receiver: pkg, Method: i.dynPackageDefMethod}, nil
	}
	if ident.Name == "eval" {
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
	sym, ok := bucket[ident.Name]
	if !ok {
		return nil, fmt.Errorf("dyn package '%s' has no member '%s'", pkgName, ident.Name)
	}
	if isPrivateSymbol(sym) {
		return nil, fmt.Errorf("dyn package '%s' member '%s' is private", pkgName, ident.Name)
	}
	return runtime.DynRefValue{Package: pkgName, Name: ident.Name}, nil
}

func (i *Interpreter) implNamespaceMember(ns runtime.ImplementationNamespaceValue, member ast.Expression) (runtime.Value, error) {
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Impl namespace member access expects identifier")
	}
	if ns.Methods == nil {
		return nil, fmt.Errorf("Impl namespace has no methods")
	}
	method, ok := ns.Methods[ident.Name]
	if !ok {
		name := "<impl>"
		if ns.Name != nil {
			name = ns.Name.Name
		}
		return nil, fmt.Errorf("No method '%s' on impl %s", ident.Name, name)
	}
	return method, nil
}

func (i *Interpreter) interfaceMember(val *runtime.InterfaceValue, member ast.Expression) (runtime.Value, error) {
	if val == nil {
		return nil, fmt.Errorf("Interface value is nil")
	}
	ident, ok := member.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("Interface member access expects identifier")
	}
	ifaceName := ""
	if val.Interface != nil && val.Interface.Node != nil && val.Interface.Node.ID != nil {
		ifaceName = val.Interface.Node.ID.Name
	}
	if ifaceName == "" {
		return nil, fmt.Errorf("Unknown interface for member access")
	}
	var method runtime.Value
	if val.Methods != nil {
		method = val.Methods[ident.Name]
	}
	if method == nil {
		if info, ok := i.getTypeInfoForValue(val.Underlying); ok {
			resolved, err := i.findMethod(info, ident.Name, ifaceName, val.InterfaceArgs)
			if err != nil {
				return nil, err
			}
			method = resolved
			if method != nil {
				if val.Methods == nil {
					val.Methods = make(map[string]runtime.Value)
				}
				val.Methods[ident.Name] = method
			}
		}
	}
	// In compiled no-bootstrap mode, fall back to compiled dispatch for inherent methods
	// that aren't in the interface definition (e.g., Iterator.collect).
	if method == nil && i.compiledInstanceMethodFn != nil {
		if info, ok := i.getTypeInfoForValue(val.Underlying); ok {
			if resolved, found := i.compiledInstanceMethodFn(info.name, ident.Name); found && resolved != nil {
				method = resolved
			}
		}
	}
	if method == nil && i.interfaceMethodResolver != nil {
		if resolved, found := i.interfaceMethodResolver(val.Underlying, ifaceName, ident.Name); found && resolved != nil {
			// interfaceMethodResolver returns arity+1; interfaceMember wraps in NativeBoundMethodValue
			// which also injects receiver. Adjust arity down by 1.
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
		if resolved, found := i.compiledInterfaceMemberFn(val.Underlying, ident.Name); found && resolved != nil {
			method = resolved
		}
	}
	// Fall back to IteratorValue native member dispatch (handles next, filter, etc.)
	if method == nil {
		if iter, ok := val.Underlying.(*runtime.IteratorValue); ok {
			if resolved, err := i.iteratorMember(iter, member); err == nil && resolved != nil {
				return resolved, nil
			}
		}
	}
	// Fall back to default interface method implementations (methods with DefaultImpl in the signature).
	if method == nil && val.Interface != nil && val.Interface.Node != nil {
		for _, sig := range val.Interface.Node.Signatures {
			if sig == nil || sig.Name == nil || sig.Name.Name != ident.Name || sig.DefaultImpl == nil {
				continue
			}
			defaultDef := ast.NewFunctionDefinition(sig.Name, sig.Params, sig.DefaultImpl, sig.ReturnType, sig.GenericParams, sig.WhereClause, false, false)
			method = &runtime.FunctionValue{Declaration: defaultDef, Closure: val.Interface.Env, MethodPriority: -1}
			break
		}
	}
	if method == nil {
		return nil, fmt.Errorf("No method '%s' for interface %s", ident.Name, ifaceName)
	}
	if fn := firstFunction(method); fn != nil {
		if fnDef, ok := fn.Declaration.(*ast.FunctionDefinition); ok && fnDef.IsPrivate {
			return nil, fmt.Errorf("Method '%s' on %s is private", ident.Name, ifaceName)
		}
	}
	receiver := interfaceMethodReceiver(i, val, method)
	switch fn := method.(type) {
	case runtime.NativeFunctionValue:
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: fn}, nil
	case *runtime.NativeFunctionValue:
		if fn == nil {
			return nil, fmt.Errorf("native method '%s' is nil", ident.Name)
		}
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: *fn}, nil
	case runtime.NativeBoundMethodValue:
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: fn.Method}, nil
	case *runtime.NativeBoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("native method '%s' is nil", ident.Name)
		}
		return runtime.NativeBoundMethodValue{Receiver: receiver, Method: fn.Method}, nil
	case runtime.BoundMethodValue:
		return runtime.BoundMethodValue{Receiver: receiver, Method: fn.Method}, nil
	case *runtime.BoundMethodValue:
		if fn == nil {
			return nil, fmt.Errorf("method '%s' is nil", ident.Name)
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
