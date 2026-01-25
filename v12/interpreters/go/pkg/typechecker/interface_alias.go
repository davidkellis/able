package typechecker

func resolveInterfaceDecl(decl Type, args []Type) (InterfaceType, []Type, bool) {
	switch val := decl.(type) {
	case InterfaceType:
		return val, args, true
	case AliasType:
		inst, _ := instantiateAlias(val, args)
		return resolveInterfaceDecl(inst, args)
	case AppliedType:
		if iface, ok := val.Base.(InterfaceType); ok {
			return iface, val.Arguments, true
		}
		if alias, ok := val.Base.(AliasType); ok {
			return resolveInterfaceDecl(alias, val.Arguments)
		}
	}
	return InterfaceType{}, nil, false
}
