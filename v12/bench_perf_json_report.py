#!/usr/bin/env python3
"""Render bench_perf's tab-separated process records as JSON."""

from __future__ import annotations

import json
import sys
from datetime import datetime, timezone

from bench_go_toolchain_contract import current_go_toolchain


def decode_metric(value: str) -> float | None:
    return None if value == "n/a" else float(value)


def main() -> None:
    (
        rows_path,
        validation_path,
        samples_path,
        output_path,
        target,
        run_from,
        runs,
        timeout_seconds,
        build_timeout_seconds,
        modes_csv,
        bytecode_runtime_calls,
        compiled_build_args,
        program_args,
        stdlib_root,
        executor,
        cpu_affinity,
        experimental_execution_context,
        source_root_only,
    ) = sys.argv[1:]
    validations = {}
    with open(validation_path, "r", encoding="utf-8") as source:
        for line in source:
            line = line.rstrip("\n")
            if not line:
                continue
            mode, status, verifier, verified, failed, hashes = line.split("\t")
            validations[mode] = {
                "status": status,
                "verifier": None if verifier == "n/a" else verifier,
                "verified_runs": int(verified),
                "failed_runs": int(failed),
                "stdout_sha256": [] if hashes == "n/a" else hashes.split(","),
            }
    samples_by_mode = {}
    with open(samples_path, "r", encoding="utf-8") as source:
        for line in source:
            line = line.rstrip("\n")
            if not line:
                continue
            (
                mode,
                run_index,
                status,
                real,
                user,
                sys_time,
                gc,
                ns,
                bytes_per_op,
                allocs,
                stdout_hash,
            ) = line.split("\t")
            samples_by_mode.setdefault(mode, []).append(
                {
                    "run": int(run_index),
                    "status": status,
                    "real_seconds": decode_metric(real),
                    "user_seconds": decode_metric(user),
                    "sys_seconds": decode_metric(sys_time),
                    "gc_count": decode_metric(gc),
                    "ns_per_op": decode_metric(ns),
                    "bytes_per_op": decode_metric(bytes_per_op),
                    "allocs_per_op": decode_metric(allocs),
                    "stdout_sha256": None if stdout_hash == "n/a" else stdout_hash,
                }
            )
    results = []
    with open(rows_path, "r", encoding="utf-8") as source:
        for line in source:
            line = line.rstrip("\n")
            if not line:
                continue
            columns = line.split("\t")
            if len(columns) == 8:
                mode, ok, timeouts, failed, avg_real, avg_user, avg_sys, avg_gc = columns
                avg_ns = avg_bytes = avg_allocs = "n/a"
            elif len(columns) == 11:
                (
                    mode,
                    ok,
                    timeouts,
                    failed,
                    avg_real,
                    avg_user,
                    avg_sys,
                    avg_gc,
                    avg_ns,
                    avg_bytes,
                    avg_allocs,
                ) = columns
            else:
                raise SystemExit(f"unexpected row shape: {columns!r}")
            results.append(
                {
                    "mode": mode,
                    "ok_runs": int(ok),
                    "timeouts": int(timeouts),
                    "failures": int(failed),
                    "avg_real_seconds": decode_metric(avg_real),
                    "avg_user_seconds": decode_metric(avg_user),
                    "avg_sys_seconds": decode_metric(avg_sys),
                    "avg_gc": decode_metric(avg_gc),
                    "avg_ns_per_op": decode_metric(avg_ns),
                    "avg_bytes_per_op": decode_metric(avg_bytes),
                    "avg_allocs_per_op": decode_metric(avg_allocs),
                    "samples": samples_by_mode.get(mode, []),
                    "validation": validations.get(
                        mode,
                        {
                            "status": "unavailable",
                            "verifier": None,
                            "verified_runs": 0,
                            "failed_runs": 0,
                            "stdout_sha256": [],
                        },
                    ),
                }
            )
    try:
        go_toolchain = current_go_toolchain()
    except ValueError as error:
        raise SystemExit(f"bench_perf_json_report: {error}") from None
    payload = {
        "generated_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "target": target,
        "run_from": run_from,
        "runs": int(runs),
        "timeout_seconds": int(timeout_seconds),
        "build_timeout_seconds": int(build_timeout_seconds),
        "modes": [mode for mode in modes_csv.split(",") if mode],
        "bytecode_prechecked": "bytecode-prechecked" in modes_csv.split(","),
        "bytecode_runtime_calls": int(bytecode_runtime_calls),
        "compiled_build_args": [arg for arg in compiled_build_args.split(" ") if arg],
        "experimental_execution_context": experimental_execution_context == "true",
        "source_root_only": source_root_only == "true",
        "program_args": [arg for arg in program_args.split(" ") if arg],
        "stdlib_root": stdlib_root or None,
        "executor": executor or None,
        "cpu_affinity": cpu_affinity or None,
        "go_toolchain": go_toolchain,
        "results": results,
    }
    with open(output_path, "w", encoding="utf-8") as destination:
        json.dump(payload, destination, indent=2)
        destination.write("\n")


if __name__ == "__main__":
    main()
