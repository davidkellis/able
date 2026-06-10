package compiler

import (
	"fmt"
	"strings"

	"able/interpreter-go/pkg/ast"
)

type primitiveRuntimeHelperInfo struct {
	DirectName          string
	ParamGoTypes        []string
	ReturnGoType        string
	CanFail             bool
	RequireNativeParams bool
}

func primitiveRuntimeHelper(name string) (primitiveRuntimeHelperInfo, bool) {
	switch strings.TrimSpace(name) {
	case "__able_String_from_builtin":
		return primitiveRuntimeHelperInfo{
			DirectName:          "__able_string_from_builtin_native",
			ParamGoTypes:        []string{"string"},
			ReturnGoType:        "*__able_array_u8",
			RequireNativeParams: true,
		}, true
	case "__able_String_to_builtin":
		return primitiveRuntimeHelperInfo{
			DirectName:          "__able_string_to_builtin_native",
			ParamGoTypes:        []string{"*__able_array_u8"},
			ReturnGoType:        "string",
			CanFail:             true,
			RequireNativeParams: true,
		}, true
	case "__able_char_from_codepoint":
		return primitiveRuntimeHelperInfo{
			DirectName:   "__able_char_from_codepoint_native",
			ParamGoTypes: []string{"int32"},
			ReturnGoType: "rune",
			CanFail:      true,
		}, true
	case "__able_char_to_codepoint":
		return primitiveRuntimeHelperInfo{
			DirectName:   "__able_char_to_codepoint_native",
			ParamGoTypes: []string{"rune"},
			ReturnGoType: "int32",
		}, true
	case "__able_char_simple_fold_next":
		return primitiveRuntimeHelperInfo{
			DirectName:   "__able_char_simple_fold_next_native",
			ParamGoTypes: []string{"rune"},
			ReturnGoType: "rune",
		}, true
	case "__able_f32_bits":
		return primitiveRuntimeHelperInfo{
			DirectName:   "__able_f32_bits_native",
			ParamGoTypes: []string{"float32"},
			ReturnGoType: "uint32",
		}, true
	case "__able_f64_bits":
		return primitiveRuntimeHelperInfo{
			DirectName:   "__able_f64_bits_native",
			ParamGoTypes: []string{"float64"},
			ReturnGoType: "uint64",
		}, true
	case "__able_f64_sqrt":
		return primitiveRuntimeHelperInfo{
			DirectName:   "__able_f64_sqrt_native",
			ParamGoTypes: []string{"float64"},
			ReturnGoType: "float64",
		}, true
	case "__able_u64_mul":
		return primitiveRuntimeHelperInfo{
			DirectName:   "__able_u64_mul_native",
			ParamGoTypes: []string{"uint64", "uint64"},
			ReturnGoType: "uint64",
		}, true
	default:
		return primitiveRuntimeHelperInfo{}, false
	}
}

func (g *generator) compilePrimitiveRuntimeHelperCall(ctx *compileContext, call *ast.FunctionCall, expected string, name string, callNode string) ([]string, string, string, bool) {
	if g == nil || ctx == nil || call == nil {
		return nil, "", "", false
	}
	helper, ok := primitiveRuntimeHelper(name)
	if !ok || len(call.Arguments) != len(helper.ParamGoTypes) {
		return nil, "", "", false
	}

	lines := make([]string, 0, len(call.Arguments)*2+4)
	args := make([]string, 0, len(call.Arguments))
	for idx, arg := range call.Arguments {
		paramType := helper.ParamGoTypes[idx]
		compileExpected := paramType
		if helper.RequireNativeParams {
			compileExpected = ""
		}
		argLines, argExpr, argType, ok := g.compileExprLines(ctx, arg, compileExpected)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, argLines...)
		if helper.RequireNativeParams && !g.typeMatches(paramType, argType) {
			return nil, "", "", false
		}
		if !g.typeMatches(paramType, argType) {
			coerceLines, coercedExpr, coercedType, ok := g.prepareStaticCallArg(ctx, argExpr, argType, paramType)
			if !ok || !g.typeMatches(paramType, coercedType) {
				return nil, "", "", false
			}
			lines = append(lines, coerceLines...)
			argExpr = coercedExpr
		}
		args = append(args, argExpr)
	}

	callExpr := fmt.Sprintf("%s(%s)", helper.DirectName, strings.Join(args, ", "))
	resultExpr := ""
	if helper.CanFail {
		var ok bool
		lines, resultExpr, ok = g.appendRuntimeHelperErrorLines(ctx, lines, callExpr, callNode)
		if !ok {
			return nil, "", "", false
		}
	} else {
		resultExpr = ctx.newTemp()
		lines = append(lines, fmt.Sprintf("%s := %s", resultExpr, callExpr))
	}

	resultType := helper.ReturnGoType
	switch {
	case expected == "", expected == "any", g.typeMatches(expected, resultType):
		return lines, resultExpr, resultType, true
	case expected == "runtime.Value":
		convLines, converted, ok := g.lowerRuntimeValue(ctx, resultExpr, resultType)
		if !ok {
			return nil, "", "", false
		}
		lines = append(lines, convLines...)
		return lines, converted, "runtime.Value", true
	case expected != "" && g.canCoerceStaticExpr(expected, resultType):
		return g.lowerCoerceExpectedStaticExpr(ctx, lines, resultExpr, resultType, expected)
	default:
		return nil, "", "", false
	}
}

func (g *generator) runtimeHelperImpl(name string) (string, bool) {
	switch name {
	case "__able_array_new":
		return "__able_array_new_impl", true
	case "__able_array_with_capacity":
		return "__able_array_with_capacity_impl", true
	case "__able_array_size":
		return "__able_array_size_impl", true
	case "__able_array_capacity":
		return "__able_array_capacity_impl", true
	case "__able_array_set_len":
		return "__able_array_set_len_impl", true
	case "__able_array_read":
		return "__able_array_read_impl", true
	case "__able_array_write":
		return "__able_array_write_impl", true
	case "__able_array_reserve":
		return "__able_array_reserve_impl", true
	case "__able_array_clone":
		return "__able_array_clone_impl", true
	case "__able_hash_map_new":
		return "__able_hash_map_new_impl", true
	case "__able_hash_map_with_capacity":
		return "__able_hash_map_with_capacity_impl", true
	case "__able_hash_map_get":
		return "__able_hash_map_get_impl", true
	case "__able_hash_map_set":
		return "__able_hash_map_set_impl", true
	case "__able_hash_map_remove":
		return "__able_hash_map_remove_impl", true
	case "__able_hash_map_contains":
		return "__able_hash_map_contains_impl", true
	case "__able_hash_map_size":
		return "__able_hash_map_size_impl", true
	case "__able_hash_map_clear":
		return "__able_hash_map_clear_impl", true
	case "__able_hash_map_for_each":
		return "__able_hash_map_for_each_impl", true
	case "__able_hash_map_clone":
		return "__able_hash_map_clone_impl", true
	case "__able_String_from_builtin":
		return "__able_string_from_builtin_impl", true
	case "__able_String_to_builtin":
		return "__able_string_to_builtin_impl", true
	case "__able_char_from_codepoint":
		return "__able_char_from_codepoint_impl", true
	case "__able_char_to_codepoint":
		return "__able_char_to_codepoint_impl", true
	case "__able_char_simple_fold_next":
		return "__able_char_simple_fold_next_impl", true
	case "__able_ratio_from_float":
		return "__able_ratio_from_float_impl", true
	case "__able_f32_bits":
		return "__able_f32_bits_impl", true
	case "__able_f64_bits":
		return "__able_f64_bits_impl", true
	case "__able_f64_sqrt":
		return "__able_f64_sqrt_impl", true
	case "__able_u64_mul":
		return "__able_u64_mul_impl", true
	case "__able_channel_new":
		return "__able_channel_new_impl", true
	case "__able_channel_send":
		return "__able_channel_send_impl", true
	case "__able_channel_receive":
		return "__able_channel_receive_impl", true
	case "__able_channel_try_send":
		return "__able_channel_try_send_impl", true
	case "__able_channel_try_receive":
		return "__able_channel_try_receive_impl", true
	case "__able_channel_await_try_recv":
		return "__able_channel_await_try_recv_impl", true
	case "__able_channel_await_try_send":
		return "__able_channel_await_try_send_impl", true
	case "__able_channel_close":
		return "__able_channel_close_impl", true
	case "__able_channel_is_closed":
		return "__able_channel_is_closed_impl", true
	case "__able_mutex_new":
		return "__able_mutex_new_impl", true
	case "__able_mutex_lock":
		return "__able_mutex_lock_impl", true
	case "__able_mutex_unlock":
		return "__able_mutex_unlock_impl", true
	case "__able_mutex_await_lock":
		return "__able_mutex_await_lock_impl", true
	default:
		return "", false
	}
}
