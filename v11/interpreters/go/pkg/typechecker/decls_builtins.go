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
	awaitableIface := InterfaceType{
		InterfaceName: "Awaitable",
		TypeParams: []GenericParamSpec{
			{Name: "Output"},
		},
		Methods: map[string]FunctionType{
			"is_ready": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
				},
				Return: boolType,
			},
			"register": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
					StructType{StructName: "AwaitWaker"},
				},
				Return: StructType{StructName: "AwaitRegistration"},
			},
			"commit": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
				},
				Return: TypeParameterType{ParameterName: "Output"},
			},
			"is_default": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
				},
				Return: boolType,
			},
		},
	}
	awaitableUnknown := AppliedType{
		Base:      awaitableIface,
		Arguments: []Type{UnknownType{}},
	}
	callbackType := FunctionType{
		Params: nil,
		Return: UnknownType{},
	}

	env.Define("proc_yield", procYield)
	env.Define("proc_cancelled", procCancelled)
	env.Define("proc_flush", procFlush)
	env.Define("proc_pending_tasks", procPendingTasks)
	env.Define("print", printFn)
	env.Define("AwaitWaker", StructType{StructName: "AwaitWaker"})
	env.Define("AwaitRegistration", StructType{StructName: "AwaitRegistration"})
	env.Define("Awaitable", awaitableIface)
	env.Define("__able_await_default", FunctionType{
		Params: []Type{callbackType},
		Return: awaitableUnknown,
	})
	env.Define("__able_await_sleep_ms", FunctionType{
		Params: []Type{i64Type, NullableType{Inner: callbackType}},
		Return: awaitableUnknown,
	})

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
	env.Define("__able_channel_await_try_recv", FunctionType{
		Params: []Type{anyType, anyType},
		Return: anyType,
	})
	env.Define("__able_channel_await_try_send", FunctionType{
		Params: []Type{anyType, anyType, anyType},
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
	env.Define("__able_mutex_await_lock", FunctionType{
		Params: []Type{i64Type, anyType},
		Return: anyType,
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

	displayIface := InterfaceType{
		InterfaceName: "Display",
		Methods: map[string]FunctionType{
			"to_string": {
				Params: []Type{
					TypeParameterType{ParameterName: "Self"},
				},
				Return: stringType,
			},
		},
	}
	cloneIface := InterfaceType{
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
	ordIface := InterfaceType{
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
	env.Define("Ord", ordIface)

	env.Define("Display", displayIface)
	env.Define("Clone", cloneIface)
	env.Define("Error", errorIface)
	divModFields := map[string]Type{
		"quotient":  TypeParameterType{ParameterName: "T"},
		"remainder": TypeParameterType{ParameterName: "T"},
	}
	env.Define("DivMod", StructType{
		StructName: "DivMod",
		TypeParams: []GenericParamSpec{{Name: "T"}},
		Fields:     divModFields,
	})
	procErrorFields := map[string]Type{
		"details": stringType,
	}
	env.Define("ProcError", StructType{
		StructName: "ProcError",
		Fields:     procErrorFields,
	})
}
