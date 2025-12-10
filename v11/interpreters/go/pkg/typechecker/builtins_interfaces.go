package typechecker

func (c *Checker) initBuiltinInterfaces() {
	stringType := PrimitiveType{Kind: PrimitiveString}
	display := InterfaceType{
		InterfaceName: "Display",
		Methods: map[string]FunctionType{
			"to_String": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
				},
				Return: stringType,
			},
		},
	}
	clone := InterfaceType{
		InterfaceName: "Clone",
		Methods: map[string]FunctionType{
			"clone": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
				},
				Return: TypeParameterType{ParameterName: "Self"},
			},
		},
	}
	errorIface := InterfaceType{
		InterfaceName: "Error",
		Methods: map[string]FunctionType{
			"message": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
				},
				Return: stringType,
			},
			"cause": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
				},
				Return: NullableType{Inner: InterfaceType{InterfaceName: "Error"}},
			},
		},
	}
	ord := InterfaceType{
		InterfaceName: "Ord",
		Methods: map[string]FunctionType{
			"cmp": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
					TypeParameterType{ParameterName: "Self"},
				},
				Return: ordering,
			},
		},
	}
	c.global.Define("Display", display)
	c.global.Define("Clone", clone)
	c.global.Define("Error", errorIface)
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
		{"Error", StructType{StructName: "ProcError"}},
	} {
		methods := map[string]FunctionType{}
		switch entry.iface {
		case "Ord":
			methods["cmp"] = FunctionType{
				Params: []Type{entry.typ, entry.typ},
				Return: ordering,
			}
		case "Display":
			methods["to_String"] = FunctionType{
				Params: []Type{entry.typ},
				Return: stringType,
			}
		case "Clone":
			methods["clone"] = FunctionType{
				Params: []Type{entry.typ},
				Return: entry.typ,
			}
		case "Error":
			methods["message"] = FunctionType{
				Params: []Type{entry.typ},
				Return: stringType,
			}
			methods["cause"] = FunctionType{
				Params: []Type{entry.typ},
				Return: NullableType{Inner: errorIface},
			}
		}
		impls = append(impls, ImplementationSpec{
			InterfaceName: entry.iface,
			Target:        entry.typ,
			Methods:       methods,
		})
	}
	c.builtinImplementations = impls
}
