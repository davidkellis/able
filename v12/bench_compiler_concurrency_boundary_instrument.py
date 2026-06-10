#!/usr/bin/env python3
"""Build a Go overlay that counts goroutine-identity boundary lookups."""

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
    parser.add_argument(
        "--main-path",
        type=Path,
        help="generated main.go to instrument (defaults to MODULE_DIR/main.go)",
    )
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--overlay", type=Path, required=True)
    args = parser.parse_args()

    module_dir = args.module_dir.resolve()
    output_dir = args.output_dir.resolve()
    output_dir.mkdir(parents=True, exist_ok=True)

    bridge_path = (
        module_dir
        / "v12/interpreters/go/pkg/compiler/bridge/bridge.go"
    )
    bridge = bridge_path.read_text(encoding="utf-8")
    bridge = replace_once(
        bridge,
        "func currentGID() uint64 {\n",
        (
            "var concurrencyBoundaryCurrentGIDCalls atomic.Uint64\n\n"
            "// ConcurrencyBoundaryCurrentGIDCalls exposes diagnostic-only "
            "overlay reach.\n"
            "func ConcurrencyBoundaryCurrentGIDCalls() uint64 {\n"
            "\treturn concurrencyBoundaryCurrentGIDCalls.Load()\n"
            "}\n\n"
            "func currentGID() uint64 {\n"
            "\tconcurrencyBoundaryCurrentGIDCalls.Add(1)\n"
        ),
        "currentGID counter",
    )

    main_path = (
        args.main_path.resolve()
        if args.main_path is not None
        else module_dir / "main.go"
    )
    main_source = main_path.read_text(encoding="utf-8")
    main_source = replace_once(
        main_source,
        "\tos.Exit(exitCode)\n",
        (
            '\tfmt.Fprintf(os.Stderr, "__ABLE_CONCURRENCY_BOUNDARY__='
            '{\\"current_gid_calls\\":%d}\\n", '
            "bridge.ConcurrencyBoundaryCurrentGIDCalls())\n"
            "\tos.Exit(exitCode)\n"
        ),
        "generated main counter report",
    )

    bridge_replacement = output_dir / "bridge.go"
    main_replacement = output_dir / "main.go"
    bridge_replacement.write_text(bridge, encoding="utf-8")
    main_replacement.write_text(main_source, encoding="utf-8")
    overlay = {
        "Replace": {
            str(bridge_path): str(bridge_replacement),
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
