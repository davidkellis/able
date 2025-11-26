package typechecker

func registerBuiltins(env *Environment) {
	if env == nil {
		return
	}

	nilType := PrimitiveType{Kind: PrimitiveNil}
	boolType := PrimitiveType{Kind: PrimitiveBool}
	i32Type := IntegerType{Suffix: "i32"}
	i64Type := IntegerType{Suffix: "i64"}
	u64Type := IntegerType{Suffix: "u64"}
	anyType := UnknownType{}
	stringType := PrimitiveType{Kind: PrimitiveString}
	charType := PrimitiveType{Kind: PrimitiveChar}
	byteArrayType := ArrayType{Element: i32Type}

	procYield := FunctionType{
		Params: nil,
		Return: nilType,
	}
	procCancelled := FunctionType{
		Params: nil,
		Return: boolType,
	}
	procFlush := FunctionType{
		Params: nil,
		Return: nilType,
	}
	procPendingTasks := FunctionType{
		Params: nil,
		Return: i32Type,
	}
	printFn := FunctionType{
		Params: []Type{UnknownType{}},
		Return: nilType,
	}

	env.Define("proc_yield", procYield)
	env.Define("proc_cancelled", procCancelled)
	env.Define("proc_flush", procFlush)
	env.Define("proc_pending_tasks", procPendingTasks)
	env.Define("print", printFn)

	env.Define("__able_channel_new", FunctionType{
		Params: []Type{i32Type},
		Return: i64Type,
	})
	env.Define("__able_channel_send", FunctionType{
		Params: []Type{anyType, anyType},
		Return: nilType,
	})
	env.Define("__able_channel_receive", FunctionType{
		Params: []Type{anyType},
		Return: anyType,
	})
	env.Define("__able_channel_try_send", FunctionType{
		Params: []Type{anyType, anyType},
		Return: boolType,
	})
	env.Define("__able_channel_try_receive", FunctionType{
		Params: []Type{anyType},
		Return: anyType,
	})
	env.Define("__able_channel_close", FunctionType{
		Params: []Type{anyType},
		Return: nilType,
	})
	env.Define("__able_channel_is_closed", FunctionType{
		Params: []Type{anyType},
		Return: boolType,
	})

	env.Define("__able_mutex_new", FunctionType{
		Params: nil,
		Return: i64Type,
	})
	env.Define("__able_mutex_lock", FunctionType{
		Params: []Type{i64Type},
		Return: nilType,
	})
	env.Define("__able_mutex_unlock", FunctionType{
		Params: []Type{i64Type},
		Return: nilType,
	})

	env.Define("__able_array_new", FunctionType{
		Params: nil,
		Return: i64Type,
	})
	env.Define("__able_array_with_capacity", FunctionType{
		Params: []Type{i32Type},
		Return: i64Type,
	})
	env.Define("__able_array_size", FunctionType{
		Params: []Type{i64Type},
		Return: u64Type,
	})
	env.Define("__able_array_capacity", FunctionType{
		Params: []Type{i64Type},
		Return: u64Type,
	})
	env.Define("__able_array_set_len", FunctionType{
		Params: []Type{i64Type, i32Type},
		Return: nilType,
	})
	env.Define("__able_array_read", FunctionType{
		Params: []Type{i64Type, i32Type},
		Return: anyType,
	})
	env.Define("__able_array_write", FunctionType{
		Params: []Type{i64Type, i32Type, anyType},
		Return: nilType,
	})
	env.Define("__able_array_reserve", FunctionType{
		Params: []Type{i64Type, i32Type},
		Return: u64Type,
	})
	env.Define("__able_array_clone", FunctionType{
		Params: []Type{i64Type},
		Return: i64Type,
	})

	env.Define("__able_string_from_builtin", FunctionType{
		Params: []Type{stringType},
		Return: byteArrayType,
	})
	env.Define("__able_string_to_builtin", FunctionType{
		Params: []Type{byteArrayType},
		Return: stringType,
	})
	env.Define("__able_char_from_codepoint", FunctionType{
		Params: []Type{i32Type},
		Return: charType,
	})

	env.Define("__able_hasher_create", FunctionType{
		Params: nil,
		Return: i64Type,
	})
	env.Define("__able_hasher_write", FunctionType{
		Params: []Type{i64Type, stringType},
		Return: nilType,
	})
	env.Define("__able_hasher_finish", FunctionType{
		Params: []Type{i64Type},
		Return: i64Type,
	})

	env.Define("Less", lessType)
	env.Define("Equal", equalType)
	env.Define("Greater", greaterType)
	env.Define("Ordering", ordering)

	ordIface := InterfaceType{
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
	env.Define("Ord", ordIface)

	env.Define("Display", InterfaceType{
		InterfaceName: "Display",
	})
	env.Define("Clone", InterfaceType{
		InterfaceName: "Clone",
	})
	procErrorFields := map[string]Type{
		"details": stringType,
	}
	env.Define("ProcError", StructType{
		StructName: "ProcError",
		Fields:     procErrorFields,
	})
}
