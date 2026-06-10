#!/usr/bin/env python3
"""Compose evidence for the build-first, preserved-binary benchmark lane."""

from __future__ import annotations

import argparse
import hashlib
import json
import statistics
from datetime import datetime, timezone
from pathlib import Path
from typing import Any


def file_fingerprint(path: Path) -> dict[str, Any]:
    data = path.read_bytes()
    return {"sha256": hashlib.sha256(data).hexdigest(), "size_bytes": len(data)}


def tree_fingerprint(root: Path) -> dict[str, Any]:
    digest = hashlib.sha256()
    count = size = 0
    for path in sorted(item for item in root.rglob("*") if item.is_file()):
        relative = path.relative_to(root).as_posix().encode()
        data = path.read_bytes()
        digest.update(len(relative).to_bytes(8, "big"))
        digest.update(relative)
        digest.update(len(data).to_bytes(8, "big"))
        digest.update(data)
        count += 1
        size += len(data)
    return {"sha256": digest.hexdigest(), "file_count": count, "size_bytes": size}


def cohort_summary(payload: dict[str, Any], requested_runs: int, index: int) -> dict[str, Any]:
    rows = [row for row in payload.get("results", []) if row.get("mode") == "compiled"]
    if len(rows) != 1:
        raise ValueError(f"cohort {index}: expected exactly one compiled result")
    row = rows[0]
    samples = row.get("samples", [])
    times = [sample["real_seconds"] for sample in samples if sample.get("status") == "ok"]
    validation = row.get("validation", {})
    complete = (
        len(times) == requested_runs
        and row.get("ok_runs") == requested_runs
        and row.get("timeouts") == 0
        and row.get("failures") == 0
        and validation.get("status") == "verified"
        and validation.get("verified_runs") == requested_runs
        and validation.get("failed_runs") == 0
    )
    return {
        "cohort": index,
        "order_index": payload["preserved_order_index"],
        "requested_runs": requested_runs,
        "ok_runs": row.get("ok_runs", 0),
        "timeouts": row.get("timeouts", 0),
        "failures": row.get("failures", 0),
        "mean_real_seconds": statistics.fmean(times) if times else None,
        "samples": samples,
        "validation": validation,
        "complete_and_verified": complete,
    }


def aggregate_cohorts(cohorts: list[dict[str, Any]], max_spread_percent: float) -> dict[str, Any]:
    means = [row["mean_real_seconds"] for row in cohorts if row["mean_real_seconds"] is not None]
    samples = [
        sample["real_seconds"]
        for cohort in cohorts
        for sample in cohort["samples"]
        if sample.get("status") == "ok" and sample.get("real_seconds") is not None
    ]
    spread = None
    if len(means) == len(cohorts) and means and min(means) > 0:
        spread = (max(means) - min(means)) / min(means) * 100.0
    consistent = spread is not None and spread <= max_spread_percent
    complete = all(row["complete_and_verified"] for row in cohorts)
    return {
        "mean_real_seconds": statistics.fmean(samples) if samples else None,
        "sample_count": len(samples),
        "cohort_spread_percent": spread,
        "max_cohort_spread_percent": max_spread_percent,
        "cohorts_consistent": consistent,
        "complete_and_verified": complete,
        "promotion_eligible": complete and consistent,
    }


def go_rows(path: str | None) -> tuple[dict[str, Any], dict[str, dict[str, Any]]]:
    if not path:
        return {}, {}
    payload = json.loads(Path(path).read_text())
    return payload, {row["benchmark"]: row for row in payload.get("rows", [])}


def compose(manifest: dict[str, Any]) -> dict[str, Any]:
    reference_payload, references = go_rows(manifest.get("go_reference_json"))
    rows = []
    all_eligible = True
    for entry in manifest["benchmarks"]:
        cohorts = []
        for index, result_path in enumerate(entry["cohort_results"], 1):
            payload = json.loads(Path(result_path).read_text())
            cohorts.append(cohort_summary(payload, manifest["runs_per_cohort"], index))
        aggregate = aggregate_cohorts(cohorts, manifest["max_cohort_spread_percent"])
        binary = Path(entry["binary"])
        generated = Path(entry["generated_dir"])
        source = Path(entry["target"])
        rejection_reasons = []
        if not aggregate["complete_and_verified"]:
            rejection_reasons.append("incomplete_or_unverified_cohort")
        if not aggregate["cohorts_consistent"]:
            rejection_reasons.append("cohort_spread_exceeds_limit")
        comparison = None
        reference = references.get(entry["benchmark"])
        if reference:
            measured_contract = entry["execution_contract"]
            reference_contract = reference.get("execution_contract")
            compatible = all(
                measured_contract.get(field) == (reference_contract or {}).get(field)
                for field in ("logical_cpu_budget", "cpu_affinity", "executor_policy")
            )
            go_seconds = reference.get("avg_real_seconds")
            able_seconds = aggregate["mean_real_seconds"]
            comparison = {
                "go_version": reference_payload.get("go_version"),
                "real_seconds": go_seconds,
                "able_to_go_ratio": able_seconds / go_seconds if able_seconds and go_seconds else None,
                "execution_contract": reference_contract,
                "contract_compatible": compatible,
                "source": {
                    "path": reference.get("source"),
                    "sha256": reference.get("source_sha256"),
                },
            }
            if not compatible:
                rejection_reasons.append("go_reference_contract_mismatch")
        else:
            rejection_reasons.append("missing_fresh_go_reference")
        eligible = not rejection_reasons
        aggregate["promotion_eligible"] = eligible
        all_eligible = all_eligible and eligible
        rows.append(
            {
                "benchmark": entry["benchmark"],
                "source": {"path": str(source), **file_fingerprint(source)},
                "artifact": {
                    "workspace_retained": manifest.get("artifact_workspace_retained", False),
                    "binary": {"path": str(binary), **file_fingerprint(binary)},
                    "generated_tree": {"path": str(generated), **tree_fingerprint(generated)},
                    "build_args": entry.get("compiled_build_args", []),
                },
                "benchmark_contract": entry["benchmark_contract"],
                "execution_contract": entry["execution_contract"],
                "cohorts": cohorts,
                "aggregate": aggregate,
                "go_reference": comparison,
                "rejection_reasons": rejection_reasons,
            }
        )
    return {
        "schema_version": 1,
        "kind": "able-preserved-compiled-cohorts",
        "generated_at": datetime.now(timezone.utc).isoformat().replace("+00:00", "Z"),
        "suite": manifest["suite"],
        "benchmark_repo": manifest["benchmark_repo"],
        "cpu_affinity_pool": manifest["cpu_affinity_pool"],
        "runs_per_cohort": manifest["runs_per_cohort"],
        "cohort_count": manifest["cohort_count"],
        "build_phase_completed_before_timing": True,
        "artifact_workspace_retained": manifest.get("artifact_workspace_retained", False),
        "ordering_policy": "forward_then_reverse",
        "max_cohort_spread_percent": manifest["max_cohort_spread_percent"],
        "promotion_eligible": all_eligible,
        "rows": rows,
    }


def markdown(payload: dict[str, Any]) -> str:
    lines = [
        "# Preserved compiled cohort comparison",
        "",
        f"Two-phase protocol: build all binaries first, then time {payload['cohort_count']} cohorts "
        f"with {payload['runs_per_cohort']} runs each. Cohort spread limit: "
        f"{payload['max_cohort_spread_percent']:.1f}%.",
        "",
        "| Benchmark | Able mean | Cohort spread | Go mean | Able / Go | Eligible |",
        "|---|---:|---:|---:|---:|:---:|",
    ]
    for row in payload["rows"]:
        aggregate = row["aggregate"]
        reference = row["go_reference"] or {}
        def number(value: Any, suffix: str = "") -> str:
            return "n/a" if value is None else f"{value:.4f}{suffix}"
        lines.append(
            "| {name} | {able} s | {spread} | {go} s | {ratio}x | {eligible} |".format(
                name=row["benchmark"], able=number(aggregate["mean_real_seconds"]),
                spread=number(aggregate["cohort_spread_percent"], "%"),
                go=number(reference.get("real_seconds")), ratio=number(reference.get("able_to_go_ratio")),
                eligible="yes" if aggregate["promotion_eligible"] else "no",
            )
        )
    lines.extend(["", f"Overall promotion eligible: **{'yes' if payload['promotion_eligible'] else 'no'}**", ""])
    return "\n".join(lines)


def main() -> None:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", required=True)
    parser.add_argument("--output-json", required=True)
    parser.add_argument("--output-md")
    args = parser.parse_args()
    payload = compose(json.loads(Path(args.manifest).read_text()))
    Path(args.output_json).write_text(json.dumps(payload, indent=2) + "\n")
    if args.output_md:
        Path(args.output_md).write_text(markdown(payload))


if __name__ == "__main__":
    main()
