package typechecker

func (c *Checker) initBuiltinInterfaces() {
	display := InterfaceType{
		InterfaceName: "Display",
	}
	clone := InterfaceType{
		InterfaceName: "Clone",
	}
	ord := InterfaceType{
		InterfaceName: "Ord",
		TypeParams: []GenericParamSpec{
			{Name: "Rhs"},
		},
		Methods: map[string]FunctionType{
			"cmp": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
					TypeParameterType{ParameterName: "Rhs"},
				},
				Return: ordering,
			},
		},
	}
	c.global.Define("Display", display)
	c.global.Define("Clone", clone)
	c.global.Define("Ord", ord)

	var impls []ImplementationSpec
	for _, entry := range []struct {
		iface string
		typ   Type
	}{
		{"Display", PrimitiveType{Kind: PrimitiveString}},
		{"Display", PrimitiveType{Kind: PrimitiveBool}},
		{"Display", PrimitiveType{Kind: PrimitiveChar}},
		{"Display", IntegerType{Suffix: "i32"}},
		{"Display", FloatType{Suffix: "f64"}},
		{"Clone", PrimitiveType{Kind: PrimitiveString}},
		{"Clone", PrimitiveType{Kind: PrimitiveBool}},
		{"Clone", PrimitiveType{Kind: PrimitiveChar}},
		{"Clone", IntegerType{Suffix: "i32"}},
		{"Clone", FloatType{Suffix: "f64"}},
		{"Ord", IntegerType{Suffix: "i32"}},
		{"Ord", PrimitiveType{Kind: PrimitiveString}},
	} {
		methods := map[string]FunctionType{}
		var ifaceArgs []Type
		if entry.iface == "Ord" {
			methods["cmp"] = FunctionType{
				Params: []Type{entry.typ, entry.typ},
				Return: ordering,
			}
			ifaceArgs = []Type{entry.typ}
		}
		impls = append(impls, ImplementationSpec{
			InterfaceName: entry.iface,
			Target:        entry.typ,
			Methods:       methods,
			InterfaceArgs: ifaceArgs,
		})
	}
	c.builtinImplementations = impls
}
