package typechecker

var (
	lessType    = StructType{StructName: "Less"}
	equalType   = StructType{StructName: "Equal"}
	greaterType = StructType{StructName: "Greater"}
	ordering    = UnionType{
		UnionName: "Ordering",
		Variants: []Type{
			lessType,
			equalType,
			greaterType,
		},
	}
)
