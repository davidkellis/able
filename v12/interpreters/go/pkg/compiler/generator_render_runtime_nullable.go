package compiler

import (
	"bytes"
	"fmt"
)

func (g *generator) renderRuntimeNullableHelpers(buf *bytes.Buffer) {
	fmt.Fprintf(buf, "func __able_ptr[T any](v T) *T {\n")
	fmt.Fprintf(buf, "\treturn &v\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "type __able_nullable[T any] struct {\n")
	fmt.Fprintf(buf, "\tvalue T\n")
	fmt.Fprintf(buf, "\tvalid bool\n")
	fmt.Fprintf(buf, "}\n\n")
	fmt.Fprintf(buf, "func __able_some[T any](v T) __able_nullable[T] {\n")
	fmt.Fprintf(buf, "\treturn __able_nullable[T]{value: v, valid: true}\n")
	fmt.Fprintf(buf, "}\n\n")
	for _, spec := range nativeNullableSpecs {
		g.renderRuntimeNullableFromHelper(buf, spec)
		g.renderRuntimeNullableFromPanicHelper(buf, spec)
		g.renderRuntimeNullableToHelper(buf, spec)
	}
}

func (g *generator) renderRuntimeNullableFromHelper(buf *bytes.Buffer, spec nativeNullableSpec) {
	fmt.Fprintf(buf, "func __able_nullable_%s_from_value(value runtime.Value) (%s, error) {\n", spec.HelperStem, spec.CarrierType)
	fmt.Fprintf(buf, "\tif __able_is_nil(value) {\n")
	if spec.ValueCarrier {
		fmt.Fprintf(buf, "\t\treturn %s{}, nil\n", spec.CarrierType)
	} else {
		fmt.Fprintf(buf, "\t\treturn nil, nil\n")
	}
	fmt.Fprintf(buf, "\t}\n")
	switch spec.InnerType {
	case "bool":
		fmt.Fprintf(buf, "\tconverted, err := bridge.AsBool(value)\n")
	case "runtime.ErrorValue":
		fmt.Fprintf(buf, "\tconverted, ok, nilPtr := __able_runtime_error_value(value)\n")
		fmt.Fprintf(buf, "\tif !ok || nilPtr {\n")
		fmt.Fprintf(buf, "\t\tconverted = bridge.ErrorValue(__able_runtime, value)\n")
		fmt.Fprintf(buf, "\t}\n")
		fmt.Fprintf(buf, "\tvar err error\n")
	case "string":
		fmt.Fprintf(buf, "\tconverted, err := bridge.AsString(value)\n")
	case "rune":
		fmt.Fprintf(buf, "\tconverted, err := bridge.AsRune(value)\n")
	case "float32":
		fmt.Fprintf(buf, "\traw, err := bridge.AsFloat(value)\n")
		fmt.Fprintf(buf, "\tconverted := float32(raw)\n")
	case "float64":
		fmt.Fprintf(buf, "\tconverted, err := bridge.AsFloat(value)\n")
	case "int":
		fmt.Fprintf(buf, "\traw, err := bridge.AsInt(value, bridge.NativeIntBits)\n")
		fmt.Fprintf(buf, "\tconverted := int(raw)\n")
	case "uint":
		fmt.Fprintf(buf, "\traw, err := bridge.AsUint(value, bridge.NativeIntBits)\n")
		fmt.Fprintf(buf, "\tconverted := uint(raw)\n")
	case "int8", "int16", "int32", "int64":
		fmt.Fprintf(buf, "\traw, err := bridge.AsInt(value, %d)\n", g.intBits(spec.InnerType))
		fmt.Fprintf(buf, "\tconverted := %s(raw)\n", spec.InnerType)
	case "uint8", "uint16", "uint32", "uint64":
		fmt.Fprintf(buf, "\traw, err := bridge.AsUint(value, %d)\n", g.intBits(spec.InnerType))
		fmt.Fprintf(buf, "\tconverted := %s(raw)\n", spec.InnerType)
	case "runtime.Int128":
		fmt.Fprintf(buf, "\tconverted, ok := runtime.Int128FromValue(value)\n")
		fmt.Fprintf(buf, "\tif !ok { return %s{}, fmt.Errorf(\"type mismatch: expected i128\") }\n", spec.CarrierType)
		fmt.Fprintf(buf, "\tvar err error\n")
	case "runtime.Uint128":
		fmt.Fprintf(buf, "\tconverted, ok := runtime.Uint128FromValue(value)\n")
		fmt.Fprintf(buf, "\tif !ok { return %s{}, fmt.Errorf(\"type mismatch: expected u128\") }\n", spec.CarrierType)
		fmt.Fprintf(buf, "\tvar err error\n")
	default:
		if spec.ValueCarrier {
			fmt.Fprintf(buf, "\treturn %s{}, fmt.Errorf(\"unsupported native nullable type %s\")\n", spec.CarrierType, spec.InnerType)
		} else {
			fmt.Fprintf(buf, "\treturn nil, fmt.Errorf(\"unsupported native nullable type %s\")\n", spec.InnerType)
		}
		fmt.Fprintf(buf, "}\n\n")
		return
	}
	fmt.Fprintf(buf, "\tif err != nil {\n")
	if spec.ValueCarrier {
		fmt.Fprintf(buf, "\t\treturn %s{}, err\n", spec.CarrierType)
	} else {
		fmt.Fprintf(buf, "\t\treturn nil, err\n")
	}
	fmt.Fprintf(buf, "\t}\n")
	if spec.ValueCarrier {
		fmt.Fprintf(buf, "\treturn __able_some(converted), nil\n")
	} else {
		fmt.Fprintf(buf, "\treturn __able_ptr(converted), nil\n")
	}
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderRuntimeNullableFromPanicHelper(buf *bytes.Buffer, spec nativeNullableSpec) {
	fmt.Fprintf(buf, "func __able_nullable_%s_from_value_or_panic(value runtime.Value) %s {\n", spec.HelperStem, spec.CarrierType)
	fmt.Fprintf(buf, "\tconverted, err := __able_nullable_%s_from_value(value)\n", spec.HelperStem)
	fmt.Fprintf(buf, "\tif err != nil {\n")
	fmt.Fprintf(buf, "\t\tpanic(err)\n")
	fmt.Fprintf(buf, "\t}\n")
	fmt.Fprintf(buf, "\treturn converted\n")
	fmt.Fprintf(buf, "}\n\n")
}

func (g *generator) renderRuntimeNullableToHelper(buf *bytes.Buffer, spec nativeNullableSpec) {
	fmt.Fprintf(buf, "func __able_nullable_%s_to_value(value %s) runtime.Value {\n", spec.HelperStem, spec.CarrierType)
	if spec.ValueCarrier {
		fmt.Fprintf(buf, "\tif !value.valid {\n")
	} else {
		fmt.Fprintf(buf, "\tif value == nil {\n")
	}
	fmt.Fprintf(buf, "\t\treturn runtime.NilValue{}\n")
	fmt.Fprintf(buf, "\t}\n")
	valueExpr := "*value"
	if spec.ValueCarrier {
		valueExpr = "value.value"
	}
	switch spec.InnerType {
	case "bool":
		fmt.Fprintf(buf, "\treturn bridge.ToBool(%s)\n", valueExpr)
	case "runtime.ErrorValue":
		fmt.Fprintf(buf, "\treturn %s\n", valueExpr)
	case "string":
		fmt.Fprintf(buf, "\treturn bridge.ToString(%s)\n", valueExpr)
	case "rune":
		fmt.Fprintf(buf, "\treturn bridge.ToRune(%s)\n", valueExpr)
	case "float32":
		fmt.Fprintf(buf, "\treturn bridge.ToFloat32(%s)\n", valueExpr)
	case "float64":
		fmt.Fprintf(buf, "\treturn bridge.ToFloat64(%s)\n", valueExpr)
	case "int", "int8", "int16", "int32", "int64":
		suffix, _ := g.integerTypeSuffix(spec.InnerType)
		fmt.Fprintf(buf, "\treturn bridge.ToInt(int64(%s), runtime.IntegerType(%q))\n", valueExpr, suffix)
	case "uint", "uint8", "uint16", "uint32", "uint64":
		suffix, _ := g.integerTypeSuffix(spec.InnerType)
		fmt.Fprintf(buf, "\treturn bridge.ToUint(uint64(%s), runtime.IntegerType(%q))\n", valueExpr, suffix)
	case "runtime.Int128", "runtime.Uint128":
		fmt.Fprintf(buf, "\treturn %s.IntegerValue()\n", valueExpr)
	default:
		fmt.Fprintf(buf, "\tpanic(fmt.Errorf(\"unsupported native nullable type %s\"))\n", spec.InnerType)
	}
	fmt.Fprintf(buf, "}\n\n")
}
