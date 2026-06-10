#!/usr/bin/env python3
"""Validate one complete scorecard's selected repeated-run evidence."""

from __future__ import annotations

import argparse
import importlib.machinery
import importlib.util
import json
import sys
from pathlib import Path

from bench_scorecard_selection import manifest_record


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
PERF_DIR = SCRIPT_DIR / "docs" / "perf-baselines"


def load_object(path: Path) -> dict:
    try:
        value = json.loads(path.read_text(encoding="utf-8"))
    except FileNotFoundError:
        raise ValueError(f"retained evidence file is missing: {path}") from None
    except json.JSONDecodeError as error:
        raise ValueError(f"invalid retained evidence JSON in {path}: {error}") from None
    if not isinstance(value, dict):
        raise ValueError(f"retained evidence root must be an object: {path}")
    return value


def retained_path(raw_path: object, context: str) -> Path:
    if not isinstance(raw_path, str) or not raw_path:
        raise ValueError(f"{context} must be a non-empty path")
    path = Path(raw_path)
    if not path.is_absolute():
        path = REPO_ROOT / path
    path = path.resolve()
    try:
        path.relative_to(PERF_DIR)
    except ValueError:
        raise ValueError(f"{context} is not retained under {PERF_DIR}: {path}") from None
    if not path.is_file():
        raise ValueError(f"{context} is missing: {path}")
    return path


def validate_retained_dependencies(scorecard_path: Path) -> tuple[int, int]:
    scorecard_path = scorecard_path.resolve()
    try:
        scorecard_path.relative_to(PERF_DIR)
    except ValueError:
        return 0, 0
    sources = load_object(scorecard_path).get("sources")
    if not isinstance(sources, list) or not sources:
        raise ValueError(f"{scorecard_path}: sources must be a non-empty array")
    reference_count = 0
    for index, raw_source in enumerate(sources):
        if not isinstance(raw_source, dict):
            raise ValueError(f"{scorecard_path}: sources[{index}] must be an object")
        source_path = retained_path(
            raw_source.get("path"), f"{scorecard_path}: sources[{index}].path"
        )
        source = load_object(source_path)
        for field in ("go_reference_json", "reference_json"):
            if source.get(field) is not None:
                retained_path(source[field], f"{source_path}: {field}")
                reference_count += 1
    return len(sources), reference_count


def load_variance_module():
    loader = importlib.machinery.SourceFileLoader(
        "bench_variance_report_evidence", str(SCRIPT_DIR / "bench_variance_report")
    )
    spec = importlib.util.spec_from_loader(loader.name, loader)
    if spec is None:
        raise RuntimeError("cannot load bench_variance_report")
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


def main() -> int:
    parser = argparse.ArgumentParser(
        description="Require complete selected Able and fresh-reference samples in one scorecard."
    )
    parser.add_argument("--scorecard", required=True, metavar="PATH")
    parser.add_argument("--selection-manifest", required=True, metavar="PATH")
    parser.add_argument("--require-runs", required=True, type=int, metavar="N")
    args = parser.parse_args()
    if args.require_runs < 1:
        parser.error("--require-runs must be at least one")

    variance = load_variance_module()
    try:
        source_count, reference_count = validate_retained_dependencies(Path(args.scorecard))
        rows, metadata, _, full_status = variance.scorecard_cohort(
            Path(args.scorecard),
            args.require_runs,
            manifest_record(Path(args.selection_manifest), REPO_ROOT),
        )
    except ValueError as error:
        print(f"bench_scorecard_evidence_check: {error}", file=sys.stderr)
        return 2
    print(
        "scorecard evidence ok: "
        f"{len(rows)} selected rows, {len(full_status)} full-status rows, "
        f"{args.require_runs} successful Able/reference samples each; "
        f"{source_count} retained sources and {reference_count} retained reference reports; "
        f"selection {metadata['selection_manifest']['selection_sha256']}"
    )
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
