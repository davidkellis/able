#!/usr/bin/env python3
"""Build a Go overlay that counts generated Awaitable carrier round trips."""

from __future__ import annotations

import argparse
import json
from pathlib import Path


def replace_once(source: str, old: str, new: str, label: str) -> str:
    count = source.count(old)
    if count != 1:
        raise ValueError(f"{label}: expected one anchor, found {count}")
    return source.replace(old, new)


def rewrite_function(
    source: str,
    start: str,
    end: str,
    old: str,
    new: str,
    label: str,
) -> str:
    start_index = source.find(start)
    if start_index < 0:
        raise ValueError(f"{label}: missing function start")
    end_index = source.find(end, start_index + len(start))
    if end_index < 0:
        raise ValueError(f"{label}: missing function end")
    body = replace_once(source[start_index:end_index], old, new, label)
    return source[:start_index] + body + source[end_index:]


COUNTERS = (
    "channel_produced",
    "channel_materialized",
    "mutex_produced",
    "mutex_materialized",
    "timer_produced",
    "timer_materialized",
    "interface_from_i64",
    "interface_to_i64",
    "array_to_runtime",
    "array_to_runtime_elements",
    "await_collect_array",
    "await_collect_elements",
    "await_collect_native_array",
    "await_collect_native_elements",
    "await_method_calls",
    "await_native_is_default",
    "await_native_is_ready",
    "await_native_register",
    "await_native_commit",
    "await_wakers_constructed",
    "await_registrations_constructed",
)


def declarations() -> str:
    fields = "".join(f"\t{name} atomic.Uint64\n" for name in COUNTERS)
    entries = "".join(
        f'\t\t"{name}": __able_awaitable_carrier_diag.{name}.Load(),\n'
        for name in COUNTERS
    )
    return (
        "var __able_awaitable_carrier_diag struct {\n"
        f"{fields}"
        "}\n\n"
        "func __able_awaitable_carrier_snapshot() map[string]uint64 {\n"
        "\treturn map[string]uint64{\n"
        f"{entries}"
        "\t}\n"
        "}\n\n"
    )


def instrument(source: str) -> str:
    channel_materialize = (
        "func (a *__able_channel_awaitable) toStruct() "
        "*runtime.StructInstanceValue {\n"
    )
    source = replace_once(
        source,
        channel_materialize,
        declarations()
        + channel_materialize
        + "\t__able_awaitable_carrier_diag.channel_materialized.Add(1)\n",
        "declarations and channel materialization",
    )
    array_header = (
        "func __able_array_iface_Awaitable_i64_to("
        "rt *bridge.Runtime, value *__able_array_iface_Awaitable_i64) "
        "(runtime.Value, error) {\n"
    )
    hooks = (
        (
            "func __able_channel_await_try_recv_impl("
            "args []runtime.Value, "
            "__able_exec_ctxs ...*__able_execution_context) "
            "(runtime.Value, error) {\n",
            "channel_produced",
        ),
        (
            "func __able_mutex_await_lock_impl("
            "args []runtime.Value, "
            "__able_exec_ctxs ...*__able_execution_context) "
            "(runtime.Value, error) {\n",
            "mutex_produced",
        ),
        (
            "func (a *__able_mutex_awaitable) toStruct() "
            "*runtime.StructInstanceValue {\n",
            "mutex_materialized",
        ),
        (
            "func __able_await_sleep_ms_impl("
            "args []runtime.Value) (runtime.Value, error) {\n",
            "timer_produced",
        ),
        (
            "func (a *__able_timer_awaitable) toStruct() "
            "*runtime.StructInstanceValue {\n",
            "timer_materialized",
        ),
        (
            "func __able_iface_Awaitable_i64_try_from_value("
            "rt *bridge.Runtime, value runtime.Value) "
            "(__able_iface_Awaitable_i64, bool, error) {\n",
            "interface_from_i64",
        ),
        (
            "func __able_iface_Awaitable_i64_to_runtime_value("
            "rt *bridge.Runtime, value __able_iface_Awaitable_i64) "
            "(runtime.Value, error) {\n",
            "interface_to_i64",
        ),
        (array_header, "array_to_runtime"),
        (
            "func __able_invoke_awaitable_method_ctx("
            "awaitable runtime.Value, method string, args []runtime.Value, "
            "__able_exec_ctx *__able_execution_context) "
            "(runtime.Value, error) {\n",
            "await_method_calls",
        ),
        (
            "func __able_make_await_waker_ctx("
            "payload *__able_async_payload, state *__able_await_state, "
            "__able_exec_ctx *__able_execution_context) "
            "(runtime.Value, error) {\n",
            "await_wakers_constructed",
        ),
        (
            "func __able_make_await_registration_value_ctx("
            "cancelFn func(), __able_exec_ctx *__able_execution_context) "
            "runtime.Value {\n",
            "await_registrations_constructed",
        ),
    )
    for header, counter in hooks:
        source = replace_once(
            source,
            header,
            header + f"\t__able_awaitable_carrier_diag.{counter}.Add(1)\n",
            counter,
        )
    source = rewrite_function(
        source,
        array_header,
        "\n}\n\ntype __able_iface_",
        "\telems := make([]runtime.Value, len(value.Elements), "
        "cap(value.Elements))\n",
        (
            "\t__able_awaitable_carrier_diag.array_to_runtime_elements."
            "Add(uint64(len(value.Elements)))\n"
            "\telems := make([]runtime.Value, len(value.Elements), "
            "cap(value.Elements))\n"
        ),
        "array elements",
    )
    pointer_collect_header = (
        "func __able_collect_await_arms_ctx("
        "iterable runtime.Value, "
        "__able_exec_ctx *__able_execution_context) "
        "([]*__able_await_arm_state, error) {\n"
    )
    value_collect_header = pointer_collect_header.replace(
        "([]*__able_await_arm_state, error)",
        "([]__able_await_arm_state, error)",
    )
    collect_header = next(
        (
            header
            for header in (pointer_collect_header, value_collect_header)
            if header in source
        ),
        "",
    )
    if not collect_header:
        raise ValueError("await collected array elements: missing function start")
    source = rewrite_function(
        source,
        collect_header,
        "func __able_await_arm_is_default_ctx(",
        "\tif values, ok := __able_array_values(iterable); ok {\n",
        (
            "\tif values, ok := __able_array_values(iterable); ok {\n"
            "\t\t__able_awaitable_carrier_diag.await_collect_array.Add(1)\n"
            "\t\t__able_awaitable_carrier_diag.await_collect_elements."
            "Add(uint64(len(values)))\n"
        ),
        "await collected array elements",
    )
    pointer_native_collect_header = (
        "func __able_collect_await_arms___able_array_iface_Awaitable_i64_ctx("
        "iterable *__able_array_iface_Awaitable_i64, "
        "__able_exec_ctx *__able_execution_context) "
        "([]*__able_await_arm_state, error) {\n"
    )
    value_native_collect_header = pointer_native_collect_header.replace(
        "([]*__able_await_arm_state, error)",
        "([]__able_await_arm_state, error)",
    )
    native_collect_header = next(
        (
            header
            for header in (
                pointer_native_collect_header,
                value_native_collect_header,
            )
            if header in source
        ),
        "",
    )
    if native_collect_header:
        source = replace_once(
            source,
            native_collect_header,
            (
                native_collect_header
                + "\t__able_awaitable_carrier_diag.await_collect_native_array."
                "Add(1)\n"
            ),
            "native await collected array",
        )
        pointer_arms = (
            "\tarms := make([]*__able_await_arm_state, 0, "
            "len(iterable.Elements))\n"
        )
        value_arms = pointer_arms.replace(
            "make([]*__able_await_arm_state",
            "make([]__able_await_arm_state",
        )
        arms_anchor = next(
            (anchor for anchor in (pointer_arms, value_arms) if anchor in source),
            "",
        )
        if not arms_anchor:
            raise ValueError("native await collected elements: missing arms slice")
        source = rewrite_function(
            source,
            native_collect_header,
            "func __able_await_value___able_array_iface_Awaitable_i64_ctx(",
            arms_anchor,
            (
                "\t__able_awaitable_carrier_diag.await_collect_native_elements."
                "Add(uint64(len(iterable.Elements)))\n"
                + arms_anchor
            ),
            "native await collected elements",
        )
    native_protocol_hooks = (
        (
            "func __able_await_arm_is_default_ctx(",
            "func __able_await_arm_is_ready_ctx(",
            "\tif arm.native != nil {\n",
            "await_native_is_default",
        ),
        (
            "func __able_await_arm_is_ready_ctx(",
            "func __able_register_await_arm_ctx(",
            "\tif arm != nil && arm.native != nil {\n",
            "await_native_is_ready",
        ),
        (
            "func __able_register_await_arm_ctx(",
            "func __able_commit_await_arm_ctx(",
            "\tif arm != nil && arm.native != nil {\n",
            "await_native_register",
        ),
        (
            "func __able_commit_await_arm_ctx(",
            "func __able_select_ready_await_arm_ctx(",
            "\tif arm != nil && arm.native != nil {\n",
            "await_native_commit",
        ),
    )
    for start, end, branch, counter in native_protocol_hooks:
        start_index = source.find(start)
        end_index = source.find(end, start_index + len(start))
        if (
            start_index < 0
            or end_index < 0
            or branch not in source[start_index:end_index]
        ):
            continue
        source = rewrite_function(
            source,
            start,
            end,
            branch,
            (
                branch
                + f"\t\t__able_awaitable_carrier_diag.{counter}.Add(1)\n"
            ),
            counter,
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
    main_path = module_dir / "main.go"

    compiled = instrument(compiled_path.read_text(encoding="utf-8"))
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
            "\tdiagnostic, _ := json.Marshal("
            "__able_awaitable_carrier_snapshot())\n"
            '\tfmt.Fprintf(os.Stderr, "__ABLE_AWAITABLE_CARRIER__=%s\\n", '
            "diagnostic)\n"
            "\tos.Exit(exitCode)\n"
        ),
        "main report",
    )

    replacements: dict[str, str] = {}
    for path, contents in (
        (compiled_path, compiled),
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
