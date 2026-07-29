#!/usr/bin/env python3
"""Build the source-identity CPU/allocation ledger for bytecode target misses."""

from __future__ import annotations

import argparse
import hashlib
import json
from pathlib import Path
from typing import Any

from bench_scorecard_selection import semantic_sha256


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
KIND = "able-bytecode-profile-coverage"
SCHEMA_VERSION = 1
MANIFEST_KIND = "able-bytecode-profile-coverage-manifest"
MANIFEST_SCHEMA_VERSION = 1


def load_object(path: Path, label: str) -> dict[str, Any]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        raise ValueError(f"{label} not found: {path}") from None
    except json.JSONDecodeError as error:
        raise ValueError(f"invalid {label} JSON in {path}: {error}") from None
    if not isinstance(value, dict):
        raise ValueError(f"{label} must be a JSON object")
    return value


def display_path(path: Path) -> str:
    try:
        return str(path.resolve().relative_to(REPO_ROOT))
    except ValueError:
        return str(path.resolve())


def fingerprint(path: Path) -> dict[str, str]:
    return {
        "path": display_path(path),
        "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
    }


def require_string(value: Any, label: str) -> str:
    if not isinstance(value, str) or not value:
        raise ValueError(f"{label} must be a non-empty string")
    return value


def evidence_records(paths: Any, group_id: str) -> list[dict[str, str]]:
    if not isinstance(paths, list) or not paths:
        raise ValueError(f"coverage group {group_id} needs evidence")
    records = []
    for raw_path in paths:
        path = REPO_ROOT / require_string(
            raw_path, f"coverage group {group_id} evidence path"
        )
        if not path.is_file():
            raise ValueError(f"coverage group {group_id} evidence not found: {raw_path}")
        records.append(fingerprint(path))
    return records


def selected_source(row: dict[str, Any], key: str) -> dict[str, str]:
    source = row.get("able_source")
    if not isinstance(source, dict):
        raise ValueError(f"{key} needs able_source")
    raw_path = require_string(source.get("path"), f"{key} source path")
    expected_sha = require_string(source.get("sha256"), f"{key} source SHA-256")
    path = REPO_ROOT / raw_path
    if not path.is_file():
        raise ValueError(f"{key} source not found: {raw_path}")
    actual_sha = hashlib.sha256(path.read_bytes()).hexdigest()
    if actual_sha != expected_sha:
        raise ValueError(f"{key} source identity is stale")
    return {"path": raw_path, "sha256": expected_sha}


def excess_seconds(scorecard: dict[str, Any], row: dict[str, Any]) -> float:
    targets = scorecard.get("targets")
    target = targets.get("bytecode") if isinstance(targets, dict) else None
    allowed = target.get("max_able_to_reference_ratio") if isinstance(target, dict) else None
    able = row.get("able")
    comparisons = row.get("comparisons")
    if (
        not isinstance(allowed, (int, float))
        or not isinstance(able, dict)
        or not isinstance(able.get("real_seconds"), (int, float))
        or not isinstance(comparisons, dict)
    ):
        raise ValueError("scorecard has an invalid bytecode timing row")
    reference_seconds = [
        value.get("real_seconds")
        for value in comparisons.values()
        if isinstance(value, dict) and isinstance(value.get("real_seconds"), (int, float))
    ]
    if not reference_seconds:
        raise ValueError("scorecard bytecode row has no timed reference")
    return max(0.0, float(able["real_seconds"]) - min(reference_seconds) * float(allowed))


def build_report(
    scorecard_path: Path,
    selection_path: Path,
    frontier_evidence_path: Path,
    closure_ledger_path: Path,
    coverage_manifest_path: Path,
) -> dict[str, Any]:
    selection = load_object(selection_path, "selection manifest")
    selection_sha = semantic_sha256(selection)
    scorecard = load_object(scorecard_path, "scorecard")
    embedded_selection = scorecard.get("selection_manifest")
    if (
        not isinstance(embedded_selection, dict)
        or embedded_selection.get("selection_sha256") != selection_sha
    ):
        raise ValueError("scorecard selection identity is stale")

    frontier = load_object(frontier_evidence_path, "frontier evidence")
    if frontier.get("selection_sha256") != selection_sha:
        raise ValueError("frontier evidence selection identity is stale")
    frontier_groups = {
        group.get("id"): group
        for group in frontier.get("groups", [])
        if isinstance(group, dict) and isinstance(group.get("id"), str)
    }

    closure_ledger = load_object(closure_ledger_path, "closure ledger")
    if closure_ledger.get("selection_sha256") != selection_sha:
        raise ValueError("closure ledger selection identity is stale")
    closures = closure_ledger.get("closures")
    if not isinstance(closures, list):
        raise ValueError("closure ledger needs closures")

    manifest = load_object(coverage_manifest_path, "coverage manifest")
    if manifest.get("kind") != MANIFEST_KIND:
        raise ValueError("coverage manifest has an invalid kind")
    if manifest.get("schema_version") != MANIFEST_SCHEMA_VERSION:
        raise ValueError(
            f"coverage manifest needs schema_version {MANIFEST_SCHEMA_VERSION}"
        )
    if manifest.get("selection_sha256") != selection_sha:
        raise ValueError("coverage manifest selection identity is stale")
    raw_coverage_groups = manifest.get("groups")
    if not isinstance(raw_coverage_groups, list) or not raw_coverage_groups:
        raise ValueError("coverage manifest needs groups")

    scorecard_rows = scorecard.get("rows")
    if not isinstance(scorecard_rows, list):
        raise ValueError("scorecard needs rows")
    misses = {
        f"{row.get('benchmark')}/bytecode": row
        for row in scorecard_rows
        if isinstance(row, dict)
        and row.get("mode") == "bytecode"
        and row.get("target_status") == "miss"
    }

    rows: list[dict[str, Any]] = []
    groups: list[dict[str, Any]] = []
    covered_keys: set[str] = set()
    seen_group_ids: set[str] = set()
    for raw_group in raw_coverage_groups:
        if not isinstance(raw_group, dict):
            raise ValueError("coverage groups must be objects")
        group_id = require_string(raw_group.get("id"), "coverage group id")
        if group_id in seen_group_ids:
            raise ValueError(f"coverage manifest repeats group {group_id}")
        seen_group_ids.add(group_id)
        if raw_group.get("cpu_status") != "current":
            raise ValueError(f"coverage group {group_id} CPU status is not current")
        if raw_group.get("allocation_status") != "current":
            raise ValueError(f"coverage group {group_id} allocation status is not current")
        frontier_group = frontier_groups.get(group_id)
        if not isinstance(frontier_group, dict):
            raise ValueError(f"coverage group {group_id} is missing from frontier evidence")
        disposition = require_string(
            frontier_group.get("disposition"), f"frontier group {group_id} disposition"
        )
        if not disposition.startswith("closed-"):
            raise ValueError(f"coverage group {group_id} is not closure-reconciled")

        group_rows = []
        for raw_key in frontier_group.get("rows", []):
            if not isinstance(raw_key, str) or not raw_key.endswith("/bytecode"):
                continue
            if raw_key not in misses:
                raise ValueError(f"coverage group {group_id} contains non-miss {raw_key}")
            if raw_key in covered_keys:
                raise ValueError(f"coverage manifest repeats row {raw_key}")
            covered_keys.add(raw_key)
            scorecard_row = misses[raw_key]
            benchmark = raw_key.removesuffix("/bytecode")
            report_row = {
                "benchmark": benchmark,
                "mode": "bytecode",
                "group": group_id,
                "source": selected_source(scorecard_row, raw_key),
                "cpu_status": "current",
                "allocation_status": "current",
                "excess_seconds": excess_seconds(scorecard, scorecard_row),
                "disposition": disposition,
            }
            rows.append(report_row)
            group_rows.append(report_row)

        evidence = evidence_records(raw_group.get("evidence"), group_id)
        groups.append(
            {
                "id": group_id,
                "row_count": len(group_rows),
                "total_excess_seconds": sum(row["excess_seconds"] for row in group_rows),
                "cpu_status": "current",
                "allocation_status": "current",
                "disposition": disposition,
                "evidence": evidence,
            }
        )

    missing = sorted(set(misses) - covered_keys)
    extra = sorted(covered_keys - set(misses))
    if missing or extra:
        details = []
        if missing:
            details.append("missing " + ", ".join(missing))
        if extra:
            details.append("extra " + ", ".join(extra))
        raise ValueError("bytecode miss coverage mismatch: " + "; ".join(details))

    rows.sort(key=lambda row: (-row["excess_seconds"], row["benchmark"]))
    groups.sort(key=lambda group: (-group["total_excess_seconds"], group["id"]))
    return {
        "kind": KIND,
        "schema_version": SCHEMA_VERSION,
        "selection_sha256": selection_sha,
        "sources": {
            "scorecard": fingerprint(scorecard_path),
            "selection_manifest": fingerprint(selection_path),
            "frontier_evidence": fingerprint(frontier_evidence_path),
            "closure_ledger": fingerprint(closure_ledger_path),
            "coverage_manifest": fingerprint(coverage_manifest_path),
        },
        "summary": {
            "bytecode_target_misses": len(misses),
            "cpu_current_rows": len(rows),
            "allocation_current_rows": len(rows),
            "uncovered_rows": 0,
            "closure_ledger_entries": len(closures),
            "total_excess_seconds": sum(row["excess_seconds"] for row in rows),
            "production_optimization_admitted": False,
        },
        "groups": groups,
        "rows": rows,
    }


def render_markdown(report: dict[str, Any]) -> str:
    summary = report["summary"]
    lines = [
        "# Bytecode current profile coverage",
        "",
        f"All {summary['bytecode_target_misses']} current bytecode target misses have "
        "source-identity-checked CPU and allocation evidence. The four bytecode "
        "target guards are intentionally excluded.",
        "",
        f"Total current target excess is {summary['total_excess_seconds']:.6f} seconds. "
        f"The retained closure ledger has {summary['closure_ledger_entries']} entries; "
        "no production optimization was admitted.",
        "",
        "## Evidence groups",
        "",
        "| Group | Rows | Target excess | CPU | Allocation | Disposition |",
        "| --- | ---: | ---: | --- | --- | --- |",
    ]
    for group in report["groups"]:
        lines.append(
            f"| `{group['id']}` | {group['row_count']} | "
            f"{group['total_excess_seconds']:.6f} s | {group['cpu_status']} | "
            f"{group['allocation_status']} | `{group['disposition']}` |"
        )
    lines.extend(
        [
            "",
            "## Source identity map",
            "",
            "| Application | Group | Source SHA-256 | Target excess |",
            "| --- | --- | --- | ---: |",
        ]
    )
    for row in report["rows"]:
        lines.append(
            f"| `{row['benchmark']}` | `{row['group']}` | "
            f"`{row['source']['sha256']}` | {row['excess_seconds']:.6f} s |"
        )
    lines.extend(
        [
            "",
            "Every row is `current` for both CPU and allocation coverage. Source files "
            "are rehashed when this ledger is generated; stale scorecard identities, "
            "missing evidence, incomplete miss coverage, and non-closed frontier groups "
            "are errors.",
            "",
        ]
    )
    return "\n".join(lines)


def main() -> int:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument(
        "--scorecard",
        type=Path,
        default=REPO_ROOT / "v12/docs/perf-baselines/external-scoreboard-current.json",
    )
    parser.add_argument(
        "--selection",
        type=Path,
        default=REPO_ROOT / "v12/bench-selection-manifest.json",
    )
    parser.add_argument(
        "--frontier-evidence",
        type=Path,
        default=REPO_ROOT / "v12/bench-performance-frontier-evidence.json",
    )
    parser.add_argument(
        "--closure-ledger",
        type=Path,
        default=REPO_ROOT / "v12/bench-performance-closure-ledger.json",
    )
    parser.add_argument(
        "--coverage-manifest",
        type=Path,
        default=REPO_ROOT / "v12/bench-bytecode-profile-coverage.json",
    )
    parser.add_argument("--output-json", type=Path)
    parser.add_argument("--output-markdown", type=Path)
    args = parser.parse_args()
    try:
        report = build_report(
            args.scorecard,
            args.selection,
            args.frontier_evidence,
            args.closure_ledger,
            args.coverage_manifest,
        )
    except ValueError as error:
        parser.error(str(error))
    rendered_json = json.dumps(report, indent=2) + "\n"
    rendered_markdown = render_markdown(report)
    if args.output_json:
        args.output_json.write_text(rendered_json, encoding="utf-8")
    if args.output_markdown:
        args.output_markdown.write_text(rendered_markdown, encoding="utf-8")
    if not args.output_json and not args.output_markdown:
        print(rendered_json, end="")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
