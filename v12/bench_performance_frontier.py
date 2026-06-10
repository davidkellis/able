#!/usr/bin/env python3
"""Build the complete, evidence-backed cross-mode performance frontier."""

from __future__ import annotations

import argparse
import hashlib
import json
import sys
from collections import Counter
from pathlib import Path
from typing import Any


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
sys.path.insert(0, str(SCRIPT_DIR))

from bench_scorecard_selection import (  # noqa: E402
    manifest_keys,
    semantic_sha256,
)


KIND = "able-performance-frontier"
SCHEMA_VERSION = 2
EVIDENCE_KIND = "able-performance-frontier-evidence"
EVIDENCE_SCHEMA_VERSION = 1
STABILITY_KIND = "able-performance-stability"
STABILITY_SCHEMA_VERSION = 1
STABILITY_CLASSIFICATIONS = {
    "established-meet",
    "variance-sensitive-miss",
    "volatile-crossing",
}
ALLOWED_DISPOSITIONS = {
    "open-candidate",
    "refresh-required",
    "target-guard",
    "closed-no-shared-leaf",
    "closed-insufficient-breadth",
    "closed-rejected-candidate",
    "closed-related-algorithms",
}
ACTION_PRIORITY = {"open-candidate": 0, "refresh-required": 1}

DEFAULT_SCORECARD = (
    REPO_ROOT
    / "v12/docs/perf-baselines/external-scoreboard-current.json"
)
DEFAULT_SELECTION = REPO_ROOT / "v12/bench-selection-manifest.json"
DEFAULT_EVIDENCE = REPO_ROOT / "v12/bench-performance-frontier-evidence.json"
DEFAULT_STABILITY = REPO_ROOT / "v12/bench-performance-stability.json"
DEFAULT_JSON = (
    REPO_ROOT / "v12/docs/perf-baselines/2026-07-20-cross-mode-performance-frontier.json"
)
DEFAULT_MARKDOWN = (
    REPO_ROOT / "v12/docs/perf-baselines/2026-07-20-cross-mode-performance-frontier.md"
)


def load_json(path: Path, label: str) -> dict[str, Any]:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        raise ValueError(f"{label} not found: {path}") from None
    except json.JSONDecodeError as error:
        raise ValueError(f"invalid {label} JSON in {path}: {error}") from None
    if not isinstance(value, dict):
        raise ValueError(f"{label} must be a JSON object")
    return value


def source_record(path: Path) -> dict[str, str]:
    return {
        "path": display_path(path),
        "sha256": hashlib.sha256(path.read_bytes()).hexdigest(),
    }


def display_path(path: Path) -> str:
    try:
        return str(path.resolve().relative_to(REPO_ROOT))
    except ValueError:
        return str(path)


def require_nonempty(group: dict[str, Any], field: str) -> str:
    value = group.get(field)
    if not isinstance(value, str) or not value.strip():
        raise ValueError(f"evidence group {group.get('id', '<unknown>')} needs {field}")
    return value


def require_sha256(value: Any, label: str) -> str:
    if (
        not isinstance(value, str)
        or len(value) != 64
        or any(character not in "0123456789abcdef" for character in value)
    ):
        raise ValueError(f"{label} needs a lowercase SHA-256 fingerprint")
    return value


def parse_row_key(value: Any, group_id: str) -> tuple[str, str]:
    if not isinstance(value, str) or value.count("/") != 1:
        raise ValueError(f"evidence group {group_id} has invalid row key {value!r}")
    benchmark, mode = value.split("/")
    if not benchmark or mode not in {"compiled", "bytecode"}:
        raise ValueError(f"evidence group {group_id} has invalid row key {value!r}")
    return benchmark, mode


def load_evidence(
    path: Path, expected_selection_sha: str, selected_keys: set[tuple[str, str]]
) -> tuple[list[dict[str, Any]], dict[tuple[str, str], dict[str, Any]]]:
    value = load_json(path, "frontier evidence manifest")
    if value.get("kind") != EVIDENCE_KIND:
        raise ValueError("frontier evidence manifest has an invalid kind")
    if value.get("schema_version") != EVIDENCE_SCHEMA_VERSION:
        raise ValueError(
            f"frontier evidence manifest needs schema_version {EVIDENCE_SCHEMA_VERSION}"
        )
    if value.get("selection_sha256") != expected_selection_sha:
        raise ValueError("frontier evidence manifest selection_sha256 is stale")
    raw_groups = value.get("groups")
    if not isinstance(raw_groups, list) or not raw_groups:
        raise ValueError("frontier evidence manifest needs groups")

    groups: list[dict[str, Any]] = []
    by_row: dict[tuple[str, str], dict[str, Any]] = {}
    seen_ids: set[str] = set()
    for raw_group in raw_groups:
        if not isinstance(raw_group, dict):
            raise ValueError("frontier evidence groups must be objects")
        group_id = require_nonempty(raw_group, "id")
        if group_id in seen_ids:
            raise ValueError(f"frontier evidence repeats group id {group_id}")
        seen_ids.add(group_id)
        disposition = require_nonempty(raw_group, "disposition")
        if disposition not in ALLOWED_DISPOSITIONS:
            raise ValueError(f"evidence group {group_id} has invalid disposition")
        breadth = raw_group.get("exact_leaf_breadth")
        if type(breadth) is not int or breadth < 0:
            raise ValueError(f"evidence group {group_id} needs exact_leaf_breadth >= 0")
        raw_rows = raw_group.get("rows")
        if not isinstance(raw_rows, list) or not raw_rows:
            raise ValueError(f"evidence group {group_id} needs rows")
        raw_paths = raw_group.get("evidence")
        if not isinstance(raw_paths, list) or not raw_paths:
            raise ValueError(f"evidence group {group_id} needs evidence paths")
        evidence_paths: list[str] = []
        for raw_path in raw_paths:
            if not isinstance(raw_path, str) or not raw_path:
                raise ValueError(f"evidence group {group_id} has an invalid evidence path")
            evidence_path = REPO_ROOT / raw_path
            if not evidence_path.is_file():
                raise ValueError(f"evidence group {group_id} path not found: {raw_path}")
            evidence_paths.append(raw_path)

        keys = [parse_row_key(row, group_id) for row in raw_rows]
        if len(keys) != len(set(keys)):
            raise ValueError(f"evidence group {group_id} repeats a row")
        group = {
            "id": group_id,
            "row_keys": keys,
            "profile_freshness": require_nonempty(raw_group, "profile_freshness"),
            "artifact_identity": require_nonempty(raw_group, "artifact_identity"),
            "disposition": disposition,
            "exact_leaf_breadth": breadth,
            "owner": require_nonempty(raw_group, "owner"),
            "rationale": require_nonempty(raw_group, "rationale"),
            "evidence": evidence_paths,
        }
        for key in keys:
            if key in by_row:
                old = by_row[key]["id"]
                raise ValueError(
                    f"frontier evidence assigns {key[0]}/{key[1]} to {old} and {group_id}"
                )
            by_row[key] = group
        groups.append(group)

    manifest_keys_seen = set(by_row)
    missing = sorted(selected_keys - manifest_keys_seen)
    extra = sorted(manifest_keys_seen - selected_keys)
    if missing or extra:
        details = []
        if missing:
            details.append("missing " + ", ".join(f"{a}/{b}" for a, b in missing))
        if extra:
            details.append("extra " + ", ".join(f"{a}/{b}" for a, b in extra))
        raise ValueError("frontier evidence row coverage mismatch: " + "; ".join(details))
    return groups, by_row


def validate_scorecard(
    scorecard: dict[str, Any],
    selected_keys: set[tuple[str, str]],
    selection_sha: str,
) -> dict[tuple[str, str], dict[str, Any]]:
    embedded = scorecard.get("selection_manifest")
    if not isinstance(embedded, dict):
        raise ValueError("scorecard needs an embedded selection manifest")
    if embedded.get("selection_sha256") != selection_sha:
        raise ValueError("scorecard selection manifest does not match current selection")
    raw_rows = scorecard.get("rows")
    if not isinstance(raw_rows, list):
        raise ValueError("scorecard needs rows")
    rows: dict[tuple[str, str], dict[str, Any]] = {}
    for row in raw_rows:
        if not isinstance(row, dict):
            raise ValueError("scorecard rows must be objects")
        key = (row.get("benchmark"), row.get("mode"))
        if not all(isinstance(part, str) and part for part in key):
            raise ValueError("scorecard row needs benchmark and mode")
        if key in rows:
            raise ValueError(f"scorecard repeats {key[0]}/{key[1]}")
        rows[key] = row
    missing = sorted(selected_keys - set(rows))
    if missing:
        rendered = ", ".join(f"{a}/{b}" for a, b in missing)
        raise ValueError(f"scorecard is missing selected rows: {rendered}")
    return rows


def load_stability(
    path: Path,
    expected_selection_sha: str,
    selected_keys: set[tuple[str, str]],
    scorecard: dict[str, Any],
    scorecard_rows: dict[tuple[str, str], dict[str, Any]],
) -> dict[tuple[str, str], dict[str, Any]]:
    value = load_json(path, "performance stability manifest")
    if value.get("kind") != STABILITY_KIND:
        raise ValueError("performance stability manifest has an invalid kind")
    if value.get("schema_version") != STABILITY_SCHEMA_VERSION:
        raise ValueError(
            f"performance stability manifest needs schema_version {STABILITY_SCHEMA_VERSION}"
        )
    if value.get("selection_sha256") != expected_selection_sha:
        raise ValueError("performance stability manifest selection_sha256 is stale")

    stdlib_state = scorecard.get("canonical_stdlib_source_state")
    stdlib_sha = (
        stdlib_state.get("source_tree_sha256")
        if isinstance(stdlib_state, dict)
        else None
    )
    expected_stdlib_sha = require_sha256(
        value.get("scorecard_stdlib_source_tree_sha256"),
        "performance stability scorecard stdlib identity",
    )
    if stdlib_sha != expected_stdlib_sha:
        raise ValueError("performance stability manifest scorecard stdlib identity is stale")

    raw_entries = value.get("entries")
    if not isinstance(raw_entries, list) or not raw_entries:
        raise ValueError("performance stability manifest needs entries")
    entries: dict[tuple[str, str], dict[str, Any]] = {}
    for raw in raw_entries:
        if not isinstance(raw, dict):
            raise ValueError("performance stability entries must be objects")
        benchmark = raw.get("benchmark")
        mode = raw.get("mode")
        key = (benchmark, mode)
        if not all(isinstance(part, str) and part for part in key):
            raise ValueError("performance stability entry needs benchmark and mode")
        if key not in selected_keys:
            raise ValueError(f"performance stability entry is not selected: {benchmark}/{mode}")
        if key in entries:
            raise ValueError(f"performance stability repeats {benchmark}/{mode}")
        scorecard_row = scorecard_rows[key]
        if scorecard_row.get("target_status") != "meets":
            raise ValueError(
                f"performance stability entry is not a snapshot meet: {benchmark}/{mode}"
            )

        classification = raw.get("classification")
        if classification not in STABILITY_CLASSIFICATIONS:
            raise ValueError(
                f"performance stability entry {benchmark}/{mode} has invalid classification"
            )
        pooled_ratio = raw.get("pooled_ratio")
        cohort_ratios = raw.get("cohort_ratios")
        if not isinstance(pooled_ratio, (int, float)) or pooled_ratio <= 0:
            raise ValueError(f"performance stability entry {benchmark}/{mode} needs pooled_ratio")
        if (
            not isinstance(cohort_ratios, list)
            or len(cohort_ratios) < 2
            or any(not isinstance(ratio, (int, float)) or ratio <= 0 for ratio in cohort_ratios)
        ):
            raise ValueError(
                f"performance stability entry {benchmark}/{mode} needs two cohort ratios"
            )
        allowed_ratio = target_ratio(scorecard, mode)
        cohort_meets = [float(ratio) <= allowed_ratio for ratio in cohort_ratios]
        if classification == "established-meet":
            if float(pooled_ratio) > allowed_ratio or not all(cohort_meets):
                raise ValueError(
                    f"established stability entry {benchmark}/{mode} does not always meet"
                )
        elif not any(cohort_meets) or all(cohort_meets):
            raise ValueError(
                f"non-established stability entry {benchmark}/{mode} must cross the target"
            )
        if classification == "variance-sensitive-miss" and float(pooled_ratio) <= allowed_ratio:
            raise ValueError(
                f"variance-sensitive miss {benchmark}/{mode} needs a pooled miss"
            )

        able_samples = raw.get("able_samples")
        reference_samples = raw.get("reference_samples")
        if type(able_samples) is not int or able_samples < 10:
            raise ValueError(f"performance stability entry {benchmark}/{mode} needs 10 Able samples")
        if type(reference_samples) is not int or reference_samples < 10:
            raise ValueError(
                f"performance stability entry {benchmark}/{mode} needs 10 reference samples"
            )

        able_source = scorecard_row.get("able_source")
        actual_able_sha = able_source.get("sha256") if isinstance(able_source, dict) else None
        expected_able_sha = require_sha256(
            raw.get("able_source_sha256"),
            f"performance stability entry {benchmark}/{mode} Able source",
        )
        if actual_able_sha != expected_able_sha:
            raise ValueError(f"performance stability Able source is stale: {benchmark}/{mode}")

        comparisons = scorecard_row.get("comparisons")
        raw_reference_shas = raw.get("reference_source_sha256")
        if not isinstance(comparisons, dict) or not isinstance(raw_reference_shas, dict):
            raise ValueError(
                f"performance stability entry {benchmark}/{mode} needs reference fingerprints"
            )
        if set(raw_reference_shas) != set(comparisons):
            raise ValueError(
                f"performance stability reference coverage is stale: {benchmark}/{mode}"
            )
        reference_shas: dict[str, str] = {}
        for language, comparison in comparisons.items():
            expected_sha = require_sha256(
                raw_reference_shas.get(language),
                f"performance stability entry {benchmark}/{mode} {language} source",
            )
            source = comparison.get("source") if isinstance(comparison, dict) else None
            actual_sha = source.get("sha256") if isinstance(source, dict) else None
            if actual_sha != expected_sha:
                raise ValueError(
                    f"performance stability {language} source is stale: {benchmark}/{mode}"
                )
            reference_shas[language] = expected_sha
        limiting_reference = raw.get("limiting_reference")
        if limiting_reference not in comparisons:
            raise ValueError(
                f"performance stability entry {benchmark}/{mode} has invalid limiting reference"
            )

        raw_evidence = raw.get("evidence")
        if not isinstance(raw_evidence, list) or not raw_evidence:
            raise ValueError(f"performance stability entry {benchmark}/{mode} needs evidence")
        evidence = []
        for evidence_path in raw_evidence:
            if not isinstance(evidence_path, str) or not evidence_path:
                raise ValueError(
                    f"performance stability entry {benchmark}/{mode} has invalid evidence"
                )
            resolved = REPO_ROOT / evidence_path
            if not resolved.is_file():
                raise ValueError(f"performance stability evidence not found: {evidence_path}")
            evidence.append(source_record(resolved))

        entries[key] = {
            "classification": classification,
            "established_guard_status": (
                "established" if classification == "established-meet" else "not-established"
            ),
            "pooled_ratio": float(pooled_ratio),
            "cohort_ratios": [float(ratio) for ratio in cohort_ratios],
            "limiting_reference": limiting_reference,
            "able_samples": able_samples,
            "reference_samples": reference_samples,
            "able_source_sha256": expected_able_sha,
            "reference_source_sha256": reference_shas,
            "evidence_stdlib_source_tree_sha256": require_sha256(
                raw.get("evidence_stdlib_source_tree_sha256"),
                f"performance stability entry {benchmark}/{mode} evidence stdlib",
            ),
            "evidence": evidence,
            "rationale": require_nonempty(raw, "rationale"),
        }

    snapshot_meets = {
        key for key in selected_keys if scorecard_rows[key].get("target_status") == "meets"
    }
    missing = sorted(snapshot_meets - set(entries))
    extra = sorted(set(entries) - snapshot_meets)
    if missing or extra:
        details = []
        if missing:
            details.append("missing " + ", ".join(f"{a}/{b}" for a, b in missing))
        if extra:
            details.append("extra " + ", ".join(f"{a}/{b}" for a, b in extra))
        raise ValueError("performance stability snapshot-meet coverage mismatch: " + "; ".join(details))
    return entries


def target_ratio(scorecard: dict[str, Any], mode: str) -> float:
    targets = scorecard.get("targets")
    target = targets.get(mode) if isinstance(targets, dict) else None
    value = target.get("max_able_to_reference_ratio") if isinstance(target, dict) else None
    if not isinstance(value, (int, float)) or value <= 0:
        raise ValueError(f"scorecard needs a positive {mode} target ratio")
    return float(value)


def build_row(
    raw: dict[str, Any],
    group: dict[str, Any],
    stability: dict[str, Any] | None,
    allowed_ratio: float,
) -> dict[str, Any]:
    able = raw.get("able")
    comparisons = raw.get("comparisons")
    if not isinstance(able, dict) or able.get("status") != "verified":
        raise ValueError(
            f"selected row {raw.get('benchmark')}/{raw.get('mode')} is not verified"
        )
    able_seconds = able.get("real_seconds")
    if not isinstance(able_seconds, (int, float)) or able_seconds < 0:
        raise ValueError("selected scorecard row needs non-negative Able real_seconds")
    if not isinstance(comparisons, dict) or not comparisons:
        raise ValueError("selected scorecard row needs comparisons")
    valid_comparisons = []
    for language, comparison in sorted(comparisons.items()):
        if not isinstance(comparison, dict):
            continue
        seconds = comparison.get("real_seconds")
        ratio = comparison.get("ratio")
        if isinstance(seconds, (int, float)) and isinstance(ratio, (int, float)):
            valid_comparisons.append(
                {
                    "language": language,
                    "real_seconds": float(seconds),
                    "ratio": float(ratio),
                }
            )
    if not valid_comparisons:
        raise ValueError("selected scorecard row has no timed comparison")
    fastest = min(valid_comparisons, key=lambda item: item["real_seconds"])
    worst = max(valid_comparisons, key=lambda item: item["ratio"])
    target_budget = fastest["real_seconds"] * allowed_ratio
    target_status = raw.get("target_status")
    if target_status not in {"meets", "miss"}:
        raise ValueError("selected scorecard row must be target-ranked")
    result = {
        "benchmark": raw["benchmark"],
        "mode": raw["mode"],
        "target_status": target_status,
        "established_guard_status": (
            stability["established_guard_status"] if stability else "not-applicable"
        ),
        "stability_classification": stability["classification"] if stability else None,
        "able_real_seconds": float(able_seconds),
        "target_ratio": allowed_ratio,
        "worst_ratio": worst["ratio"],
        "worst_ratio_reference": worst["language"],
        "fastest_reference": fastest["language"],
        "fastest_reference_real_seconds": fastest["real_seconds"],
        "target_budget_seconds": target_budget,
        "excess_seconds": max(0.0, float(able_seconds) - target_budget),
        "comparisons": valid_comparisons,
        "able_source": raw.get("able_source"),
        "group": group["id"],
        "profile_freshness": group["profile_freshness"],
        "artifact_identity": group["artifact_identity"],
        "disposition": group["disposition"],
        "exact_leaf_breadth": group["exact_leaf_breadth"],
        "owner": group["owner"],
        "evidence": group["evidence"],
    }
    if stability:
        result["stability"] = stability
    return result


def build_frontier(
    scorecard_path: Path,
    selection_path: Path,
    evidence_path: Path,
    stability_path: Path,
) -> dict[str, Any]:
    selection = load_json(selection_path, "selection manifest")
    selected_keys = manifest_keys(selection)
    selection_sha = semantic_sha256(selection)
    scorecard = load_json(scorecard_path, "scorecard")
    scorecard_rows = validate_scorecard(scorecard, selected_keys, selection_sha)
    groups, evidence_by_row = load_evidence(
        evidence_path, selection_sha, selected_keys
    )
    stability_by_row = load_stability(
        stability_path, selection_sha, selected_keys, scorecard, scorecard_rows
    )

    rows = []
    for key in sorted(selected_keys, key=lambda item: (item[1], item[0])):
        row = scorecard_rows[key]
        rows.append(
            build_row(
                row,
                evidence_by_row[key],
                stability_by_row.get(key),
                target_ratio(scorecard, key[1]),
            )
        )

    group_summaries = []
    for group in groups:
        group_rows = [row for row in rows if row["group"] == group["id"]]
        group_summaries.append(
            {
                "id": group["id"],
                "row_count": len(group_rows),
                "miss_count": sum(row["target_status"] == "miss" for row in group_rows),
                "total_excess_seconds": sum(row["excess_seconds"] for row in group_rows),
                "maximum_ratio": max(row["worst_ratio"] for row in group_rows),
                "profile_freshness": group["profile_freshness"],
                "artifact_identity": group["artifact_identity"],
                "disposition": group["disposition"],
                "exact_leaf_breadth": group["exact_leaf_breadth"],
                "owner": group["owner"],
                "rationale": group["rationale"],
                "evidence": group["evidence"],
            }
        )
    actionable = [
        group for group in group_summaries if group["disposition"] in ACTION_PRIORITY
    ]
    actionable.sort(
        key=lambda group: (
            ACTION_PRIORITY[group["disposition"]],
            -group["total_excess_seconds"],
            group["id"],
        )
    )
    recommendation = None
    if actionable:
        lead = actionable[0]
        recommendation = {
            "group": lead["id"],
            "action": lead["disposition"],
            "total_excess_seconds": lead["total_excess_seconds"],
            "why": lead["rationale"],
        }

    dispositions = Counter(row["disposition"] for row in rows)
    return {
        "kind": KIND,
        "schema_version": SCHEMA_VERSION,
        "generated_from": scorecard.get("generated_at"),
        "sources": {
            "scorecard": source_record(scorecard_path),
            "selection_manifest": {
                **source_record(selection_path),
                "selection_sha256": selection_sha,
            },
            "evidence_manifest": source_record(evidence_path),
            "stability_manifest": source_record(stability_path),
        },
        "summary": {
            "selected_rows": len(rows),
            "compiled_rows": sum(row["mode"] == "compiled" for row in rows),
            "bytecode_rows": sum(row["mode"] == "bytecode" for row in rows),
            "target_meets": sum(row["target_status"] == "meets" for row in rows),
            "target_misses": sum(row["target_status"] == "miss" for row in rows),
            "established_guards": sum(
                row["established_guard_status"] == "established" for row in rows
            ),
            "unestablished_snapshot_meets": sum(
                row["target_status"] == "meets"
                and row["established_guard_status"] != "established"
                for row in rows
            ),
            "compiled_established_guards": sum(
                row["mode"] == "compiled"
                and row["established_guard_status"] == "established"
                for row in rows
            ),
            "bytecode_established_guards": sum(
                row["mode"] == "bytecode"
                and row["established_guard_status"] == "established"
                for row in rows
            ),
            "total_excess_seconds": sum(row["excess_seconds"] for row in rows),
            "disposition_rows": dict(sorted(dispositions.items())),
            "actionable_groups": len(actionable),
        },
        "recommendation": recommendation,
        "actionable_groups": actionable,
        "groups": group_summaries,
        "rows": rows,
    }


def fmt_number(value: float, digits: int = 3) -> str:
    return f"{value:.{digits}f}"


def md_link(path: str) -> str:
    return f"[{Path(path).name}](../../{path.removeprefix('v12/')})"


def render_markdown(report: dict[str, Any]) -> str:
    summary = report["summary"]
    lines = [
        "# Cross-mode performance frontier",
        "",
        f"Generated from `{report['generated_from']}` scorecard evidence. This ledger joins "
        f"all {summary['selected_rows']} reviewed rows; it excludes unselected status-only rows.",
        "",
        "## Outcome",
        "",
        f"- Selected rows: {summary['compiled_rows']} compiled + "
        f"{summary['bytecode_rows']} bytecode = {summary['selected_rows']}.",
        f"- Product target: {summary['target_meets']} meet, {summary['target_misses']} miss.",
        f"- Established cross-cohort guards: {summary['established_guards']} "
        f"({summary['compiled_established_guards']} compiled + "
        f"{summary['bytecode_established_guards']} bytecode); "
        f"{summary['unestablished_snapshot_meets']} snapshot meets are not established.",
        f"- Aggregate time above the per-row 95%-of-reference budget: "
        f"{fmt_number(summary['total_excess_seconds'])} seconds.",
        f"- Unclosed groups: {summary['actionable_groups']}.",
        "",
    ]
    recommendation = report["recommendation"]
    if recommendation:
        lines.extend(
            [
                "## Recommended next gate",
                "",
                f"Refresh `{recommendation['group']}` first "
                f"({fmt_number(recommendation['total_excess_seconds'])} aggregate excess seconds). "
                f"{recommendation['why']}",
                "",
            ]
        )
    lines.extend(
        [
            "## Actionable groups",
            "",
            "| Rank | Group | Action | Rows | Misses | Excess s | Max ratio | Freshness |",
            "| ---: | --- | --- | ---: | ---: | ---: | ---: | --- |",
        ]
    )
    for rank, group in enumerate(report["actionable_groups"], 1):
        lines.append(
            f"| {rank} | `{group['id']}` | {group['disposition']} | "
            f"{group['row_count']} | {group['miss_count']} | "
            f"{fmt_number(group['total_excess_seconds'])} | "
            f"{fmt_number(group['maximum_ratio'])} | {group['profile_freshness']} |"
        )
    if not report["actionable_groups"]:
        lines.append("| - | None | - | - | - | - | - | - |")

    lines.extend(
        [
            "",
            "## Complete selected-row ledger",
            "",
            "`Excess s` is Able wall time beyond the fastest applicable reference multiplied "
            "by the allowed ratio (1 / 0.95). `Established` is a separate cross-cohort "
            "candidate-admission guard and never rewrites snapshot status.",
            "",
            "| Benchmark | Mode | Snapshot | Established | Stability | Able s | Worst ratio | Excess s | Freshness | Disposition | Group |",
            "| --- | --- | --- | --- | --- | ---: | ---: | ---: | --- | --- | --- |",
        ]
    )
    for row in report["rows"]:
        lines.append(
            f"| {row['benchmark']} | {row['mode']} | {row['target_status']} | "
            f"{row['established_guard_status']} | "
            f"{row['stability_classification'] or '-'} | "
            f"{fmt_number(row['able_real_seconds'])} | {fmt_number(row['worst_ratio'])} | "
            f"{fmt_number(row['excess_seconds'])} | {row['profile_freshness']} | "
            f"{row['disposition']} | `{row['group']}` |"
        )

    lines.extend(
        [
            "",
            "## Cross-cohort stability evidence",
            "",
        ]
    )
    for row in report["rows"]:
        stability = row.get("stability")
        if not stability:
            continue
        cohort_ratios = ", ".join(
            fmt_number(ratio) for ratio in stability["cohort_ratios"]
        )
        evidence = ", ".join(
            md_link(record["path"]) for record in stability["evidence"]
        )
        lines.extend(
            [
                f"### `{row['benchmark']}/{row['mode']}`",
                "",
                f"- Classification: `{stability['classification']}`; established guard: "
                f"`{stability['established_guard_status']}`; pooled limiting ratio: "
                f"{fmt_number(stability['pooled_ratio'])}x "
                f"{stability['limiting_reference']}; cohort ratios: {cohort_ratios}.",
                f"- Samples: {stability['able_samples']} Able and "
                f"{stability['reference_samples']} limiting-reference; Able source: "
                f"`{stability['able_source_sha256']}`; evidence stdlib tree: "
                f"`{stability['evidence_stdlib_source_tree_sha256']}`.",
                f"- Rationale: {stability['rationale']}",
                f"- Evidence: {evidence}.",
                "",
            ]
        )

    lines.extend(
        [
            "## Ownership and disposition evidence",
            "",
        ]
    )
    for group in report["groups"]:
        evidence = ", ".join(md_link(path) for path in group["evidence"])
        lines.extend(
            [
                f"### `{group['id']}`",
                "",
                f"- Disposition: `{group['disposition']}`; exact unlike-application breadth: "
                f"{group['exact_leaf_breadth']}; profile freshness: "
                f"`{group['profile_freshness']}`; artifact identity: "
                f"`{group['artifact_identity']}`.",
                f"- Owner: {group['owner']}",
                f"- Rationale: {group['rationale']}",
                f"- Evidence: {evidence}.",
                "",
            ]
        )
    return "\n".join(lines).rstrip() + "\n"


def write_or_check(path: Path, content: str, check: bool) -> None:
    if check:
        try:
            current = path.read_text(encoding="utf-8")
        except FileNotFoundError:
            raise ValueError(f"generated frontier is missing: {path}") from None
        if current != content:
            raise ValueError(f"generated frontier is stale: {path}")
        return
    path.parent.mkdir(parents=True, exist_ok=True)
    path.write_text(content, encoding="utf-8")


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--scorecard", type=Path, default=DEFAULT_SCORECARD)
    parser.add_argument("--selection-manifest", type=Path, default=DEFAULT_SELECTION)
    parser.add_argument("--evidence-manifest", type=Path, default=DEFAULT_EVIDENCE)
    parser.add_argument("--stability-manifest", type=Path, default=DEFAULT_STABILITY)
    parser.add_argument("--output-json", type=Path, default=DEFAULT_JSON)
    parser.add_argument("--output-markdown", type=Path, default=DEFAULT_MARKDOWN)
    parser.add_argument("--check", action="store_true")
    return parser.parse_args()


def main() -> int:
    args = parse_args()
    try:
        report = build_frontier(
            args.scorecard.resolve(),
            args.selection_manifest.resolve(),
            args.evidence_manifest.resolve(),
            args.stability_manifest.resolve(),
        )
        json_text = json.dumps(report, indent=2, sort_keys=False) + "\n"
        markdown = render_markdown(report)
        write_or_check(args.output_json.resolve(), json_text, args.check)
        write_or_check(args.output_markdown.resolve(), markdown, args.check)
    except ValueError as error:
        print(f"performance frontier error: {error}", file=sys.stderr)
        return 1
    action = "verified" if args.check else "wrote"
    print(
        f"{action} performance frontier: {report['summary']['selected_rows']} rows, "
        f"{report['summary']['actionable_groups']} actionable groups"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
