#!/usr/bin/env python3
"""Create a Go overlay that marks corrected generated struct-field reads."""

from __future__ import annotations

import argparse
import json
import re
from pathlib import Path


MARKER_PREFIX = "__ABLE_STRUCT_FIELD_REACH__:"
BRANCHES = [
    "string_bytes",
    "array_values",
    "array_apply",
    "array_member_set",
    "array_member_get",
    "generic_member_get",
    "callable_member_get",
    "future_error_details",
    "main_array_shape",
    "main_array_format",
    "awaitable_shape",
    "runtime_functional_update",
    "ir_functional_update",
]


def function_span(source: str, name: str) -> tuple[int, int]:
    start = source.find(f"func {name}(")
    if start < 0:
        raise ValueError(f"generated function not found: {name}")
    next_func = source.find("\nfunc ", start + 1)
    return start, len(source) if next_func < 0 else next_func


def instrument_function(
    source: str,
    name: str,
    needle: str,
    branch: str,
    expected: int,
) -> tuple[str, int]:
    start, end = function_span(source, name)
    body = source[start:end]
    count = body.count(needle)
    if count != expected:
        raise ValueError(
            f"{name}: expected {expected} occurrences of {needle!r}, found {count}"
        )
    indent = needle[: len(needle) - len(needle.lstrip())]
    marker = (
        f'{indent}__able_struct_field_reach_mark('
        f'&__able_struct_field_reach_{branch}, "{branch}")\n'
    )
    body = body.replace(needle, marker + needle)
    return source[:start] + body + source[end:], count


def instrument_generated(module_dir: Path, output_dir: Path) -> tuple[dict[str, str], dict[str, int]]:
    compiled_path = module_dir / "compiled.go"
    main_path = module_dir / "main.go"
    compiled = compiled_path.read_text(encoding="utf-8")
    main = main_path.read_text(encoding="utf-8")
    counts = {branch: 0 for branch in BRANCHES}

    declaration_anchor = "var __able_runtime *bridge.Runtime\n"
    if compiled.count(declaration_anchor) != 1:
        raise ValueError("compiled.go: runtime declaration anchor is not unique")
    declarations = ["", "func __able_struct_field_reach_mark(once *sync.Once, name string) {"]
    declarations += [
        f'\tonce.Do(func() {{ println("{MARKER_PREFIX}" + name) }})',
        "}",
        "",
    ]
    declarations += [
        f"var __able_struct_field_reach_{branch} sync.Once"
        for branch in BRANCHES
        if branch != "awaitable_shape"
    ]
    declarations.append("")
    compiled = compiled.replace(
        declaration_anchor,
        declaration_anchor + "\n".join(declarations) + "\n",
    )

    generated_specs = [
        (
            "__able_string_bytes_from_struct",
            '\tbytesVal, _ := __able_struct_named_field_value(inst, "bytes")\n',
            "string_bytes",
            1,
        ),
        (
            "__able_array_values",
            '\t\thandle, ok := __able_struct_named_field_value(inst, "storage_handle")\n',
            "array_values",
            1,
        ),
        (
            "__able_struct_Array_apply",
            '\tif handleVal, ok := __able_struct_named_field_value(inst, "storage_handle"); ok {\n',
            "array_apply",
            1,
        ),
        (
            "__able_try_member_set",
            '\t\t\t\t\tif handleVal, ok := __able_struct_named_field_value(inst, "storage_handle"); ok {\n',
            "array_member_set",
            2,
        ),
        (
            "__able_try_member_get",
            '\t\t\t\thandleVal, ok := __able_struct_named_field_value(inst, "storage_handle")\n',
            "array_member_get",
            1,
        ),
        (
            "__able_try_member_get",
            "\t\t\tif val, ok := __able_struct_named_field_value(inst, name); ok {\n",
            "generic_member_get",
            1,
        ),
        (
            "__able_try_member_get_method",
            "\t\t\t\tif val, ok := __able_struct_named_field_value(inst, name); ok && __able_is_callable_value(val) {\n",
            "callable_member_get",
            1,
        ),
        (
            "__able_future_error_details",
            '\t\t\tif detail, ok := __able_struct_named_field_value(v, "details"); ok {\n',
            "future_error_details",
            1,
        ),
    ]
    for name, needle, branch, expected in generated_specs:
        compiled, count = instrument_function(
            compiled, name, needle, branch, expected
        )
        counts[branch] += count

    main_specs = [
        (
            "isArrayStructInstance",
            '\t_, hasHandle := __able_struct_named_field_value(v, "storage_handle")\n',
            "main_array_shape",
            1,
        ),
        (
            "isArrayStructInstance",
            '\t_, hasLength := __able_struct_named_field_value(v, "length")\n',
            "main_array_shape",
            1,
        ),
        (
            "isArrayStructInstance",
            '\t_, hasCapacity := __able_struct_named_field_value(v, "capacity")\n',
            "main_array_shape",
            1,
        ),
        (
            "formatRuntimeValue",
            '\t\t\tif h, ok := __able_struct_named_field_value(v, "storage_handle"); ok {\n',
            "main_array_format",
            1,
        ),
    ]
    for name, needle, branch, expected in main_specs:
        main, count = instrument_function(main, name, needle, branch, expected)
        counts[branch] += count

    functional_pattern = re.compile(
        r"^(?P<indent>\s*)for _, defField := range .*Definition\.Node\.Fields "
        r"\{.*(?P<message>[Ff]unctional update source missing field).*$",
        re.MULTILINE,
    )
    replacements: dict[Path, str] = {
        compiled_path: compiled,
        main_path: main,
    }
    for path in sorted(module_dir.glob("*.go")):
        source = replacements.get(path)
        if source is None:
            source = path.read_text(encoding="utf-8")
        matched = False

        def replace_functional(match: re.Match[str]) -> str:
            nonlocal matched
            matched = True
            message = match.group("message")
            branch = (
                "runtime_functional_update"
                if message.startswith("Functional")
                else "ir_functional_update"
            )
            counts[branch] += 1
            indent = match.group("indent")
            marker = (
                f'{indent}__able_struct_field_reach_mark('
                f'&__able_struct_field_reach_{branch}, "{branch}")\n'
            )
            return marker + match.group(0)

        updated = functional_pattern.sub(replace_functional, source)
        if matched:
            replacements[path] = updated

    bridge_path = (
        module_dir
        / "v12/interpreters/go/pkg/compiler/bridge/bridge_static_types.go"
    )
    bridge = bridge_path.read_text(encoding="utf-8")
    bridge_anchor = "func matchTypeWithoutInterpreter("
    if bridge.count(bridge_anchor) != 1:
        raise ValueError("bridge static-types function anchor is not unique")
    bridge = bridge.replace(
        bridge_anchor,
        "var __able_struct_field_reach_awaitable_shape bool\n\n" + bridge_anchor,
    )
    start, end = function_span(bridge, "isRuntimeAwaitable")
    body = bridge[start:end]
    return_anchor = "\treturn true\n}"
    if body.count(return_anchor) != 1:
        raise ValueError("isRuntimeAwaitable: successful return anchor is not unique")
    reach = (
        '\tif !__able_struct_field_reach_awaitable_shape {\n'
        "\t\t__able_struct_field_reach_awaitable_shape = true\n"
        f'\t\tprintln("{MARKER_PREFIX}awaitable_shape")\n'
        "\t}\n"
    )
    body = body.replace(return_anchor, reach + return_anchor)
    bridge = bridge[:start] + body + bridge[end:]
    counts["awaitable_shape"] = 1
    replacements[bridge_path] = bridge

    output_dir.mkdir(parents=True, exist_ok=True)
    overlay: dict[str, str] = {}
    for index, (original, contents) in enumerate(sorted(replacements.items())):
        replacement = output_dir / f"{index:02d}-{original.name}"
        replacement.write_text(contents, encoding="utf-8")
        overlay[str(original.resolve())] = str(replacement.resolve())

    return overlay, counts


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--module-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--overlay", type=Path, required=True)
    parser.add_argument("--metadata", type=Path, required=True)
    args = parser.parse_args()

    module_dir = args.module_dir.resolve()
    output_dir = args.output_dir.resolve()
    overlay, counts = instrument_generated(module_dir, output_dir)
    args.overlay.parent.mkdir(parents=True, exist_ok=True)
    args.metadata.parent.mkdir(parents=True, exist_ok=True)
    args.overlay.write_text(
        json.dumps({"Replace": overlay}, sort_keys=True, indent=2) + "\n",
        encoding="utf-8",
    )
    args.metadata.write_text(
        json.dumps(
            {
                "kind": "able-compiler-struct-field-reach-instrumentation",
                "schema_version": 1,
                "marker_prefix": MARKER_PREFIX,
                "branches": counts,
            },
            sort_keys=True,
            indent=2,
        )
        + "\n",
        encoding="utf-8",
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
