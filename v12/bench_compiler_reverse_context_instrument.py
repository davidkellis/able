#!/usr/bin/env python3
"""Build a Go overlay that attributes generated execution-context rebuilding."""

from __future__ import annotations

import argparse
import json
from pathlib import Path


def replace_once(source: str, old: str, new: str, label: str) -> str:
    count = source.count(old)
    if count != 1:
        raise ValueError(f"{label}: expected one anchor, found {count}")
    return source.replace(old, new)


def declarations() -> str:
    return """
var __able_reverse_context_diag struct {
	fromEnvironment atomic.Uint64
	fromNative      atomic.Uint64
	nativeNil       atomic.Uint64
	nativeState     atomic.Uint64
	spawn           atomic.Uint64
}

var __able_reverse_context_callers_mu sync.Mutex
var __able_reverse_context_callers = map[string]uint64{}

func __able_reverse_context_record_caller(kind string) {
	var pcs [1]uintptr
	if goruntime.Callers(3, pcs[:]) != 1 {
		return
	}
	frame, _ := goruntime.CallersFrames(pcs[:]).Next()
	name := strings.TrimPrefix(frame.Function, "main.")
	__able_reverse_context_callers_mu.Lock()
	__able_reverse_context_callers[kind+"|"+name]++
	__able_reverse_context_callers_mu.Unlock()
}

func __able_reverse_context_snapshot() map[string]any {
	__able_reverse_context_callers_mu.Lock()
	callers := make(map[string]uint64, len(__able_reverse_context_callers))
	for name, count := range __able_reverse_context_callers {
		callers[name] = count
	}
	__able_reverse_context_callers_mu.Unlock()
	return map[string]any{
		"from_environment": __able_reverse_context_diag.fromEnvironment.Load(),
		"from_native": __able_reverse_context_diag.fromNative.Load(),
		"native_nil": __able_reverse_context_diag.nativeNil.Load(),
		"native_state_payload": __able_reverse_context_diag.nativeState.Load(),
		"spawn": __able_reverse_context_diag.spawn.Load(),
		"callers": callers,
	}
}

"""


def instrument_compiled(source: str) -> str:
    source = replace_once(
        source,
        '\t"math"\n',
        '\t"math"\n\tgoruntime "runtime"\n',
        "Go runtime import",
    )
    environment_header = (
        "func __able_context_from_environment(env *runtime.Environment) "
        "*__able_execution_context {\n"
    )
    source = replace_once(
        source,
        environment_header,
        declarations()
        + environment_header
        + "\t__able_reverse_context_diag.fromEnvironment.Add(1)\n"
        + '\t__able_reverse_context_record_caller("environment")\n',
        "environment counter",
    )
    native_header = (
        "func __able_context_from_native(native *runtime.NativeCallContext) "
        "*__able_execution_context {\n"
    )
    source = replace_once(
        source,
        native_header,
        native_header
        + "\t__able_reverse_context_diag.fromNative.Add(1)\n"
        + '\t__able_reverse_context_record_caller("native")\n',
        "native counter",
    )
    source = replace_once(
        source,
        "\tif native == nil {\n",
        "\tif native == nil {\n"
        "\t\t__able_reverse_context_diag.nativeNil.Add(1)\n",
        "nil native counter",
    )
    source = replace_once(
        source,
        "\tctx := __able_context_from_environment(native.Env)\n",
        (
            "\tif _, ok := native.State.(*__able_async_payload); ok {\n"
            "\t\t__able_reverse_context_diag.nativeState.Add(1)\n"
            "\t}\n"
            "\tctx := __able_context_from_environment(native.Env)\n"
        ),
        "native payload counter",
    )
    spawn_anchor = "\t\tchild := __able_context_from_environment(taskEnv)\n"
    source = replace_once(
        source,
        spawn_anchor,
        "\t\t__able_reverse_context_diag.spawn.Add(1)\n" + spawn_anchor,
        "spawn counter",
    )
    return source


def instrument_main(source: str) -> str:
    source = replace_once(
        source,
        '\t"fmt"\n',
        '\t"encoding/json"\n\t"fmt"\n',
        "JSON import",
    )
    return replace_once(
        source,
        "\tos.Exit(exitCode)\n",
        (
            "\tdiagnostic, _ := json.Marshal(__able_reverse_context_snapshot())\n"
            '\tfmt.Fprintf(os.Stderr, "__ABLE_REVERSE_CONTEXT__=%s\\n", '
            "diagnostic)\n"
            "\tos.Exit(exitCode)\n"
        ),
        "main report",
    )


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--module-dir", type=Path, required=True)
    parser.add_argument("--output-dir", type=Path, required=True)
    parser.add_argument("--overlay", type=Path, required=True)
    args = parser.parse_args()

    module_dir = args.module_dir.resolve()
    output_dir = args.output_dir.resolve()
    output_dir.mkdir(parents=True, exist_ok=True)
    paths = {
        module_dir / "compiled.go": instrument_compiled(
            (module_dir / "compiled.go").read_text(encoding="utf-8")
        ),
        module_dir / "main.go": instrument_main(
            (module_dir / "main.go").read_text(encoding="utf-8")
        ),
    }
    replacements: dict[str, str] = {}
    for path, contents in paths.items():
        replacement = output_dir / path.name
        replacement.write_text(contents, encoding="utf-8")
        replacements[str(path)] = str(replacement)
    args.overlay.write_text(
        json.dumps({"Replace": replacements}, sort_keys=True, indent=2) + "\n",
        encoding="utf-8",
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
