#!/usr/bin/env python3
"""Build a Go overlay that attributes callable dispatch and currentGID reach."""

from __future__ import annotations

import argparse
import json
from pathlib import Path


def replace_once(source: str, old: str, new: str, label: str) -> str:
    count = source.count(old)
    if count != 1:
        raise ValueError(f"{label}: expected one replacement anchor, found {count}")
    return source.replace(old, new)


def rewrite_function(
    source: str,
    start: str,
    end: str,
    replacements: list[tuple[str, str, str]],
    label: str,
) -> str:
    start_index = source.find(start)
    if start_index < 0:
        raise ValueError(f"{label}: missing function start")
    end_index = source.find(end, start_index + len(start))
    if end_index < 0:
        raise ValueError(f"{label}: missing function end")
    body = source[start_index:end_index]
    for old, new, replacement_label in replacements:
        body = replace_once(body, old, new, f"{label} {replacement_label}")
    return source[:start_index] + body + source[end_index:]


COUNTERS = (
    "ctx_total",
    "ctx_native_function",
    "ctx_native_function_ptr",
    "ctx_native_bound_method",
    "ctx_native_bound_method_ptr",
    "ctx_compatibility_default",
    "ctx_default_with_env",
    "ctx_default_env_swapped",
    "legacy_total",
    "legacy_native_function",
    "legacy_native_function_ptr",
    "legacy_native_bound_method",
    "legacy_native_bound_method_ptr",
    "legacy_compatibility_default",
    "method_ctx_total",
    "method_node_ctx_total",
    "method_lookup_success",
    "method_lookup_error",
    "method_lookup_direct",
    "method_lookup_without_direct",
    "method_callable_success",
    "method_callable_error",
    "method_nil_result",
    "method_value_result",
)


def counter_declarations() -> str:
    fields = "".join(f"\t{counter} atomic.Uint64\n" for counter in COUNTERS)
    entries = "".join(
        f'\t\t"{counter}": __able_callable_dispatch_diag.{counter}.Load(),\n'
        for counter in COUNTERS
    )
    return (
        "var __able_callable_dispatch_diag struct {\n"
        f"{fields}"
        "}\n\n"
        "var __able_callable_dispatch_bound_name_mu sync.Mutex\n"
        "var __able_callable_dispatch_bound_names = map[string]uint64{}\n\n"
        "func __able_callable_dispatch_record_bound(fn runtime.NativeFunctionValue) {\n"
        '\tkey := fn.Name + "|borrow=" + strconv.FormatBool(fn.BorrowArgs)\n'
        "\t__able_callable_dispatch_bound_name_mu.Lock()\n"
        "\t__able_callable_dispatch_bound_names[key]++\n"
        "\t__able_callable_dispatch_bound_name_mu.Unlock()\n"
        "}\n\n"
        "func __able_callable_dispatch_bound_name_snapshot() map[string]uint64 {\n"
        "\t__able_callable_dispatch_bound_name_mu.Lock()\n"
        "\tdefer __able_callable_dispatch_bound_name_mu.Unlock()\n"
        "\tout := make(map[string]uint64, len(__able_callable_dispatch_bound_names))\n"
        "\tfor key, count := range __able_callable_dispatch_bound_names {\n"
        "\t\tout[key] = count\n"
        "\t}\n"
        "\treturn out\n"
        "}\n\n"
        "func __able_callable_dispatch_snapshot() map[string]uint64 {\n"
        "\treturn map[string]uint64{\n"
        f"{entries}"
        "\t}\n"
        "}\n\n"
    )


def instrument_compiled(source: str) -> str:
    legacy_header = (
        "func __able_call_value_fast(fn runtime.Value, "
        "args []runtime.Value) (runtime.Value, error) {\n"
    )
    ctx_header = (
        "func __able_call_value_fast_ctx(fn runtime.Value, args []runtime.Value, "
        "__able_exec_ctx *__able_execution_context) (runtime.Value, error) {\n"
    )
    source = replace_once(
        source,
        legacy_header,
        counter_declarations()
        + legacy_header
        + "\t__able_callable_dispatch_diag.legacy_total.Add(1)\n",
        "diagnostic declarations and legacy total",
    )
    source = rewrite_function(
        source,
        legacy_header,
        ctx_header,
        [
            (
                "\tcase runtime.NativeFunctionValue:\n",
                "\tcase runtime.NativeFunctionValue:\n"
                "\t\t__able_callable_dispatch_diag.legacy_native_function.Add(1)\n",
                "native function",
            ),
            (
                "\tcase *runtime.NativeFunctionValue:\n",
                "\tcase *runtime.NativeFunctionValue:\n"
                "\t\t__able_callable_dispatch_diag.legacy_native_function_ptr.Add(1)\n",
                "native function pointer",
            ),
            (
                "\tcase runtime.NativeBoundMethodValue:\n",
                "\tcase runtime.NativeBoundMethodValue:\n"
                "\t\t__able_callable_dispatch_diag.legacy_native_bound_method.Add(1)\n",
                "native bound method",
            ),
            (
                "\t\t__able_callable_dispatch_diag.legacy_native_bound_method.Add(1)\n"
                "\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n",
                "\t\t__able_callable_dispatch_diag.legacy_native_bound_method.Add(1)\n"
                "\t\t__able_callable_dispatch_record_bound(v.Method)\n"
                "\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n",
                "bound method name",
            ),
            (
                "\tcase *runtime.NativeBoundMethodValue:\n",
                "\tcase *runtime.NativeBoundMethodValue:\n"
                "\t\t__able_callable_dispatch_diag.legacy_native_bound_method_ptr.Add(1)\n",
                "native bound method pointer",
            ),
            (
                '\t\t\treturn nil, fmt.Errorf("native bound method is nil")\n'
                "\t\t}\n"
                "\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n",
                '\t\t\treturn nil, fmt.Errorf("native bound method is nil")\n'
                "\t\t}\n"
                "\t\t__able_callable_dispatch_record_bound(v.Method)\n"
                "\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n",
                "bound method pointer name",
            ),
            (
                "\tdefault:\n",
                "\tdefault:\n"
                "\t\t__able_callable_dispatch_diag.legacy_compatibility_default.Add(1)\n",
                "compatibility default",
            ),
        ],
        "legacy callable helper",
    )
    source = replace_once(
        source,
        ctx_header,
        ctx_header + "\t__able_callable_dispatch_diag.ctx_total.Add(1)\n",
        "context total",
    )
    method_header = (
        "func __able_method_call_ctx(obj runtime.Value, methodName string, "
        "args []runtime.Value, __able_exec_ctx *__able_execution_context) "
        "(runtime.Value, *__ableControl) {\n"
    )
    source = rewrite_function(
        source,
        ctx_header,
        method_header,
        [
            (
                "\tcase runtime.NativeFunctionValue:\n",
                "\tcase runtime.NativeFunctionValue:\n"
                "\t\t__able_callable_dispatch_diag.ctx_native_function.Add(1)\n",
                "native function",
            ),
            (
                "\tcase *runtime.NativeFunctionValue:\n",
                "\tcase *runtime.NativeFunctionValue:\n"
                "\t\t__able_callable_dispatch_diag.ctx_native_function_ptr.Add(1)\n",
                "native function pointer",
            ),
            (
                "\tcase runtime.NativeBoundMethodValue:\n",
                "\tcase runtime.NativeBoundMethodValue:\n"
                "\t\t__able_callable_dispatch_diag.ctx_native_bound_method.Add(1)\n",
                "native bound method",
            ),
            (
                "\t\t__able_callable_dispatch_diag.ctx_native_bound_method.Add(1)\n"
                "\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n",
                "\t\t__able_callable_dispatch_diag.ctx_native_bound_method.Add(1)\n"
                "\t\t__able_callable_dispatch_record_bound(v.Method)\n"
                "\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n",
                "bound method name",
            ),
            (
                "\tcase *runtime.NativeBoundMethodValue:\n",
                "\tcase *runtime.NativeBoundMethodValue:\n"
                "\t\t__able_callable_dispatch_diag.ctx_native_bound_method_ptr.Add(1)\n",
                "native bound method pointer",
            ),
            (
                '\t\t\treturn nil, fmt.Errorf("native bound method is nil")\n'
                "\t\t}\n"
                "\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n",
                '\t\t\treturn nil, fmt.Errorf("native bound method is nil")\n'
                "\t\t}\n"
                "\t\t__able_callable_dispatch_record_bound(v.Method)\n"
                "\t\tinjected := append([]runtime.Value{v.Receiver}, args...)\n",
                "bound method pointer name",
            ),
            (
                "\tdefault:\n",
                "\tdefault:\n"
                "\t\t__able_callable_dispatch_diag.ctx_compatibility_default.Add(1)\n",
                "compatibility default",
            ),
            (
                "\t\tif ctx.Env != nil {\n",
                "\t\tif ctx.Env != nil {\n"
                "\t\t\t__able_callable_dispatch_diag.ctx_default_with_env.Add(1)\n",
                "default environment",
            ),
            (
                "\t\t\tif previous, swapped := bridge.SwapEnvIfNeeded"
                "(__able_runtime, ctx.Env); swapped {\n",
                "\t\t\tif previous, swapped := bridge.SwapEnvIfNeeded"
                "(__able_runtime, ctx.Env); swapped {\n"
                "\t\t\t\t__able_callable_dispatch_diag.ctx_default_env_swapped.Add(1)\n",
                "default environment swap",
            ),
        ],
        "context callable helper",
    )
    node_header = (
        "func __able_method_call_node_ctx(obj runtime.Value, methodName string, "
        "args []runtime.Value, call *ast.FunctionCall, "
        "__able_exec_ctx *__able_execution_context) "
        "(runtime.Value, *__ableControl) {\n"
    )
    source = instrument_method(source, method_header, node_header, "method_ctx_total")
    source = instrument_method(
        source,
        node_header,
        "func __able_await_value_ctx(",
        "method_node_ctx_total",
    )
    return source


def instrument_method(source: str, start: str, end: str, total: str) -> str:
    return rewrite_function(
        source,
        start,
        end,
        [
            (
                start,
                start + f"\t__able_callable_dispatch_diag.{total}.Add(1)\n",
                "total",
            ),
            (
                "\tmethod, err := __able_try_member_get_method"
                "(obj, runtime.StringValue{Val: methodName})\n"
                "\tif err != nil {\n",
                "\tmethod, err := __able_try_member_get_method"
                "(obj, runtime.StringValue{Val: methodName})\n"
                "\tif err != nil {\n"
                "\t\t__able_callable_dispatch_diag.method_lookup_error.Add(1)\n",
                "lookup error",
            ),
            (
                "\tval, err := __able_call_value_fast_ctx"
                "(method, args, __able_exec_ctx)\n"
                "\tif err != nil {\n",
                "\t__able_callable_dispatch_diag.method_lookup_success.Add(1)\n"
                "\tif nativeBound, ok, _ := "
                "__able_callable_native_bound_method_value(method); ok {\n"
                "\t\ttypeName := "
                "__able_runtime_value_type_name(nativeBound.Receiver)\n"
                "\t\tentry := __able_lookup_compiled_method"
                "(typeName, methodName, true)\n"
                "\t\tif entry != nil && entry.direct != nil {\n"
                "\t\t\t__able_callable_dispatch_diag."
                "method_lookup_direct.Add(1)\n"
                "\t\t} else {\n"
                "\t\t\t__able_callable_dispatch_diag."
                "method_lookup_without_direct.Add(1)\n"
                "\t\t}\n"
                "\t}\n"
                "\tval, err := __able_call_value_fast_ctx"
                "(method, args, __able_exec_ctx)\n"
                "\tif err != nil {\n"
                "\t\t__able_callable_dispatch_diag.method_callable_error.Add(1)\n",
                "lookup success and callable error",
            ),
            (
                "\tif val == nil {\n",
                "\t__able_callable_dispatch_diag.method_callable_success.Add(1)\n"
                "\tif val == nil {\n"
                "\t\t__able_callable_dispatch_diag.method_nil_result.Add(1)\n",
                "callable success and nil result",
            ),
            (
                "\treturn val, nil\n",
                "\t__able_callable_dispatch_diag.method_value_result.Add(1)\n"
                "\treturn val, nil\n",
                "value result",
            ),
        ],
        total,
    )


def instrument_bridge(source: str) -> str:
    source = replace_once(
        source,
        "func currentGID() uint64 {\n",
        (
            "var callableDispatchCurrentGIDMu sync.Mutex\n"
            "var callableDispatchCurrentGIDStacks = map[string]uint64{}\n\n"
            "// CallableDispatchCurrentGIDStacks exposes diagnostic-only overlay data.\n"
            "func CallableDispatchCurrentGIDStacks() map[string]uint64 {\n"
            "\tcallableDispatchCurrentGIDMu.Lock()\n"
            "\tdefer callableDispatchCurrentGIDMu.Unlock()\n"
            "\tout := make(map[string]uint64, len(callableDispatchCurrentGIDStacks))\n"
            "\tfor key, count := range callableDispatchCurrentGIDStacks {\n"
            "\t\tout[key] = count\n"
            "\t}\n"
            "\treturn out\n"
            "}\n\n"
            "func currentGID() uint64 {\n"
            "\tvar pcs [8]uintptr\n"
            "\tcount := goruntime.Callers(2, pcs[:])\n"
            "\tframes := goruntime.CallersFrames(pcs[:count])\n"
            "\tparts := make([]string, 0, 4)\n"
            "\tfor len(parts) < 4 {\n"
            "\t\tframe, more := frames.Next()\n"
            "\t\tname := frame.Function\n"
            '\t\tname = strings.TrimPrefix(name, "able/interpreter-go/")\n'
            '\t\tname = strings.TrimPrefix(name, "main.")\n'
            "\t\tparts = append(parts, name)\n"
            "\t\tif !more {\n"
            "\t\t\tbreak\n"
            "\t\t}\n"
            "\t}\n"
            '\tstack := strings.Join(parts, " <- ")\n'
            "\tcallableDispatchCurrentGIDMu.Lock()\n"
            "\tcallableDispatchCurrentGIDStacks[stack]++\n"
            "\tcallableDispatchCurrentGIDMu.Unlock()\n"
        ),
        "currentGID stack attribution",
    )
    return source


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--module-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--overlay", type=Path, required=True)
    args = parser.parse_args()

    module_dir = args.module_dir.resolve()
    output_dir = args.output_dir.resolve()
    output_dir.mkdir(parents=True, exist_ok=True)

    compiled_path = module_dir / "compiled.go"
    bridge_path = (
        module_dir / "v12/interpreters/go/pkg/compiler/bridge/bridge.go"
    )
    main_path = module_dir / "main.go"

    compiled = instrument_compiled(compiled_path.read_text(encoding="utf-8"))
    bridge = instrument_bridge(bridge_path.read_text(encoding="utf-8"))
    main_source = main_path.read_text(encoding="utf-8")
    main_source = replace_once(
        main_source,
        '\t"fmt"\n',
        '\t"encoding/json"\n\t"fmt"\n',
        "main json import",
    )
    main_source = replace_once(
        main_source,
        "\tos.Exit(exitCode)\n",
        (
            "\tdiagnostic, _ := json.Marshal(map[string]any{\n"
            '\t\t"dispatch": __able_callable_dispatch_snapshot(),\n'
            '\t\t"bound_names": __able_callable_dispatch_bound_name_snapshot(),\n'
            '\t\t"current_gid_stacks": bridge.CallableDispatchCurrentGIDStacks(),\n'
            "\t})\n"
            '\tfmt.Fprintf(os.Stderr, "__ABLE_CALLABLE_DISPATCH__=%s\\n", diagnostic)\n'
            "\tos.Exit(exitCode)\n"
        ),
        "generated main report",
    )

    replacements = {}
    for path, contents in (
        (compiled_path, compiled),
        (bridge_path, bridge),
        (main_path, main_source),
    ):
        replacement = output_dir / path.name
        replacement.write_text(contents, encoding="utf-8")
        replacements[str(path)] = str(replacement)

    args.overlay.parent.mkdir(parents=True, exist_ok=True)
    args.overlay.write_text(
        json.dumps({"Replace": replacements}, sort_keys=True, indent=2) + "\n",
        encoding="utf-8",
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
