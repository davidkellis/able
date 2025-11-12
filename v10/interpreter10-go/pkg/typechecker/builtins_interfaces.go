package typechecker

func (c *Checker) initBuiltinInterfaces() {
	display := InterfaceType{
		InterfaceName: "Display",
	}
	clone := InterfaceType{
		InterfaceName: "Clone",
	}
	c.global.Define("Display", display)
	c.global.Define("Clone", clone)

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
	} {
		impls = append(impls, ImplementationSpec{
			InterfaceName: entry.iface,
			Target:        entry.typ,
		})
	}
	c.builtinImplementations = impls
}
