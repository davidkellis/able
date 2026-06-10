#!/usr/bin/env python3
"""Build a Go overlay that keeps closure-owned kernel callables receiver-free."""

from __future__ import annotations

import argparse
import json
from pathlib import Path


METHODS = (
    "isReady",
    "register",
    "commit",
    "isDefault",
    "wakeFn",
    "cancelMethod",
)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--module-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--overlay", type=Path, required=True)
    parser.add_argument("--report", type=Path)
    args = parser.parse_args()

    module_dir = args.module_dir.resolve()
    output_dir = args.output_dir.resolve()
    output_dir.mkdir(parents=True, exist_ok=True)

    compiled_path = module_dir / "compiled.go"
    source = compiled_path.read_text(encoding="utf-8")
    counts: dict[str, int] = {}
    for method in METHODS:
        old = (
            "&runtime.NativeBoundMethodValue"
            f"{{Receiver: inst, Method: {method}}}"
        )
        count = source.count(old)
        counts[method] = count
        source = source.replace(old, method)

    if sum(counts.values()) == 0:
        raise ValueError("generated source has no closure-owned bound methods")

    replacement = output_dir / "compiled.go"
    replacement.write_text(source, encoding="utf-8")
    overlay = {"Replace": {str(compiled_path): str(replacement)}}
    args.overlay.parent.mkdir(parents=True, exist_ok=True)
    args.overlay.write_text(
        json.dumps(overlay, sort_keys=True, indent=2) + "\n",
        encoding="utf-8",
    )
    if args.report is not None:
        args.report.parent.mkdir(parents=True, exist_ok=True)
        args.report.write_text(
            json.dumps(
                {
                    "schema_version": 1,
                    "replacement_counts": counts,
                    "total_replacements": sum(counts.values()),
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
