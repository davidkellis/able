#!/usr/bin/env python3
"""Build a Go overlay that counts the generated callable-context path."""

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
    call_header = (
        "func __able_call_value_fast_ctx(fn runtime.Value, args []runtime.Value, "
        "__able_exec_ctx *__able_execution_context) (runtime.Value, error) {\n"
    )
    compiled = replace_once(
        compiled,
        call_header,
        (
            "var __able_callable_context_call_value_fast_calls atomic.Uint64\n"
            "var __able_callable_context_await_calls atomic.Uint64\n\n"
            + call_header
            + "\t__able_callable_context_call_value_fast_calls.Add(1)\n"
        ),
        "context-aware callable helper",
    )
    await_header = (
        "func __able_await_ctx(expr *ast.AwaitExpression, iterable runtime.Value, "
        "__able_exec_ctx *__able_execution_context) runtime.Value {\n"
    )
    compiled = replace_once(
        compiled,
        await_header,
        await_header + "\t__able_callable_context_await_calls.Add(1)\n",
        "context-aware await helper",
    )

    main_path = module_dir / "main.go"
    main_source = main_path.read_text(encoding="utf-8")
    main_source = replace_once(
        main_source,
        "\tos.Exit(exitCode)\n",
        (
            '\tfmt.Fprintf(os.Stderr, "__ABLE_CALLABLE_CONTEXT_REACH__='
            '{\\"call_value_fast\\":%d,\\"await\\":%d}\\n", '
            "__able_callable_context_call_value_fast_calls.Load(), "
            "__able_callable_context_await_calls.Load())\n"
            "\tos.Exit(exitCode)\n"
        ),
        "generated main report",
    )

    compiled_replacement = output_dir / "compiled.go"
    main_replacement = output_dir / "main.go"
    compiled_replacement.write_text(compiled, encoding="utf-8")
    main_replacement.write_text(main_source, encoding="utf-8")
    overlay = {
        "Replace": {
            str(compiled_path): str(compiled_replacement),
            str(main_path): str(main_replacement),
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
