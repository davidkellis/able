package typechecker

func stringMemberType(memberName string) (Type, bool) {
	u64 := IntegerType{Suffix: "u64"}
	switch memberName {
	case "len_bytes", "len_chars", "len_graphemes":
		return FunctionType{Params: nil, Return: u64}, true
	case "bytes":
		return FunctionType{Params: nil, Return: IteratorType{Element: IntegerType{Suffix: "u8"}}}, true
	case "chars":
		return FunctionType{Params: nil, Return: IteratorType{Element: PrimitiveType{Kind: PrimitiveChar}}}, true
	case "graphemes":
		return FunctionType{Params: nil, Return: IteratorType{Element: PrimitiveType{Kind: PrimitiveString}}}, true
	case "substring":
		errorOrString := UnionLiteralType{
			Members: []Type{
				StructType{StructName: "Error"},
				PrimitiveType{Kind: PrimitiveString},
			},
		}
		return FunctionType{
			Params: []Type{u64, NullableType{Inner: u64}},
			Return: errorOrString,
		}, true
	case "split":
		return FunctionType{
			Params: []Type{PrimitiveType{Kind: PrimitiveString}},
			Return: ArrayType{Element: PrimitiveType{Kind: PrimitiveString}},
		}, true
	case "replace":
		return FunctionType{
			Params: []Type{
				PrimitiveType{Kind: PrimitiveString},
				PrimitiveType{Kind: PrimitiveString},
			},
			Return: PrimitiveType{Kind: PrimitiveString},
		}, true
	case "starts_with", "ends_with":
		return FunctionType{
			Params: []Type{PrimitiveType{Kind: PrimitiveString}},
			Return: PrimitiveType{Kind: PrimitiveBool},
		}, true
	default:
		return nil, false
	}
}
