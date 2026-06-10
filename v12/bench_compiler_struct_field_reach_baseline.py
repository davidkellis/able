#!/usr/bin/env python3
"""Reconstruct the pre-audit behavior for the two census-reached branches."""

from __future__ import annotations

import argparse
import json
from pathlib import Path


def replace_once(source: str, old: str, new: str, label: str) -> str:
    count = source.count(old)
    if count != 1:
        raise ValueError(f"{label}: expected one replacement anchor, found {count}")
    return source.replace(old, new)


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
    compiled = compiled_path.read_text(encoding="utf-8")
    candidate_callable = (
        "\t\t\t\tif val, ok := __able_struct_named_field_value(inst, name); "
        "ok && __able_is_callable_value(val) {\n"
        "\t\t\t\t\treturn val, nil\n"
        "\t\t\t\t}\n"
    )
    baseline_callable = (
        "\t\t\t\tif inst.Fields != nil {\n"
        "\t\t\t\t\tif val, ok := inst.Fields[name]; ok && "
        "__able_is_callable_value(val) {\n"
        "\t\t\t\t\t\treturn val, nil\n"
        "\t\t\t\t\t}\n"
        "\t\t\t\t}\n"
    )
    compiled = replace_once(
        compiled,
        candidate_callable,
        baseline_callable,
        "callable-member baseline",
    )

    bridge_path = (
        module_dir
        / "v12/interpreters/go/pkg/compiler/bridge/bridge_static_types.go"
    )
    bridge = bridge_path.read_text(encoding="utf-8")
    candidate_awaitable = (
        '\tfor _, name := range []string{"is_ready", "register", "commit", '
        '"is_default"} {\n'
        "\t\tfield, ok := structNamedFieldValue(inst, name)\n"
        "\t\tif !ok || !isRuntimeCallableValue(field) {\n"
        "\t\t\treturn false\n"
        "\t\t}\n"
        "\t}\n"
    )
    baseline_awaitable = (
        '\tfor _, name := range []string{"is_ready", "register", "commit", '
        '"is_default"} {\n'
        "\t\tif !isRuntimeCallableValue(inst.Fields[name]) {\n"
        "\t\t\treturn false\n"
        "\t\t}\n"
        "\t}\n"
    )
    bridge = replace_once(
        bridge,
        candidate_awaitable,
        baseline_awaitable,
        "Awaitable baseline",
    )

    compiled_replacement = output_dir / "compiled.go"
    bridge_replacement = output_dir / "bridge_static_types.go"
    compiled_replacement.write_text(compiled, encoding="utf-8")
    bridge_replacement.write_text(bridge, encoding="utf-8")
    overlay = {
        "Replace": {
            str(compiled_path.resolve()): str(compiled_replacement.resolve()),
            str(bridge_path.resolve()): str(bridge_replacement.resolve()),
        }
    }
    args.overlay.parent.mkdir(parents=True, exist_ok=True)
    args.overlay.write_text(
        json.dumps(overlay, sort_keys=True, indent=2) + "\n",
        encoding="utf-8",
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
