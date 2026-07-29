#!/usr/bin/env python3
"""Fast tests for full-status and strict-selection scorecard separation."""

from __future__ import annotations

import importlib.machinery
import importlib.util
import json
import subprocess
import sys
import tempfile
import unittest
from unittest import mock
from pathlib import Path
from types import ModuleType
from typing import Any


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
sys.path.insert(0, str(SCRIPT_DIR))

from bench_scorecard_selection import manifest_keys, manifest_record  # noqa: E402


def load_script(name: str) -> ModuleType:
    loader = importlib.machinery.SourceFileLoader(name, str(SCRIPT_DIR / name))
    spec = importlib.util.spec_from_loader(name, loader)
    if spec is None:
        raise RuntimeError(f"cannot load {name}")
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


variance = load_script("bench_variance_report")
scoreboard_tool = load_script("bench_external_scoreboard")


def successful_samples(seconds: float = 1.0) -> list[dict[str, Any]]:
    return [
        {
            "run": run,
            "status": "ok",
            "real_seconds": seconds + run / 1000,
            "stdout_sha256": "output-hash",
        }
        for run in range(1, 6)
    ]


def good_row(benchmark: str, mode: str = "compiled") -> dict[str, Any]:
    languages = ("go",) if mode == "compiled" else ("python", "ruby")
    return {
        "benchmark": benchmark,
        "mode": mode,
        "able": {
            "ok_runs": 5,
            "timeouts": 0,
            "failures": 0,
            "avg_real_seconds": 1.003,
            "validation": {"status": "verified", "verified_runs": 5, "failed_runs": 0},
            "samples": successful_samples(),
        },
        "comparisons": {
            language: {
                "ratio": 2.0,
                "reference_real_seconds": 0.5,
                "reference_source": {"provenance": "measured"},
                "reference_samples": successful_samples(0.5),
            }
            for language in languages
        },
    }


def timeout_row(
    benchmark: str, mode: str = "bytecode", timeouts: int = 5
) -> dict[str, Any]:
    return {
        "benchmark": benchmark,
        "mode": mode,
        "able": {
            "ok_runs": 0,
            "timeouts": timeouts,
            "failures": 0,
            "avg_real_seconds": None,
            "validation": {"status": "not-run", "verified_runs": 0, "failed_runs": 0},
            "samples": [],
        },
        "comparisons": {"go": {"ratio": None}},
    }


STDLIB_STATE = {
    "kind": "canonical-stdlib-source-state",
    "root": "../able-stdlib",
    "source_file_count": 1,
    "source_tree_sha256": "a" * 64,
    "git_head": "b" * 40,
    "git_dirty": True,
}


class SelectionVarianceTests(unittest.TestCase):
    def write_manifest(self, root: Path, bytecode: list[str] | None = None) -> Path:
        path = root / "selection.json"
        path.write_text(
            json.dumps(
                {
                    "kind": "external-benchmark-selection-manifest",
                    "schema_version": 1,
                    "modes": {
                        "compiled": ["fast"],
                        "bytecode": bytecode or ["fast_bytecode"],
                    },
                }
            ),
            encoding="utf-8",
        )
        return path

    def write_cohort(
        self,
        root: Path,
        name: str,
        selection: dict[str, Any],
        rows: list[dict[str, Any]],
        stdlib_state: dict[str, Any] = STDLIB_STATE,
    ) -> Path:
        source_path = root / f"{name}-source.json"
        source_path.write_text(
            json.dumps(
                {
                    "generated_at": name,
                    "suite": "test",
                    "modes": ["compiled", "bytecode"],
                    "languages": ["go", "python", "ruby"],
                    "rows": rows,
                }
            ),
            encoding="utf-8",
        )
        cohort_path = root / f"{name}-cohort.json"
        cohort_path.write_text(
            json.dumps(
                {
                    "kind": "external-benchmark-scoreboard",
                    "generated_at": name,
                    "suite": "test",
                    "modes": ["compiled", "bytecode"],
                    "languages": ["go", "python", "ruby"],
                    "canonical_stdlib_source_state": stdlib_state,
                    "selection_manifest": selection,
                    "sources": [{"path": str(source_path)}],
                    "rows": rows,
                }
            ),
            encoding="utf-8",
        )
        return cohort_path

    def test_timeout_stays_in_full_status_but_not_selected_variance(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            manifest_path = self.write_manifest(root)
            selection = manifest_record(manifest_path, REPO_ROOT)
            rows = [
                good_row("fast"),
                good_row("fast_bytecode", "bytecode"),
                timeout_row("slow", timeouts=1),
            ]
            cohorts = [
                self.write_cohort(root, "a", selection, rows),
                self.write_cohort(root, "b", selection, rows),
            ]
            report = variance.build_scorecard_report(cohorts, 5, selection)
            self.assertEqual(
                {(row["benchmark"], row["mode"]) for row in report["rows"]},
                {("fast", "compiled"), ("fast_bytecode", "bytecode")},
            )
            self.assertEqual([source["full_status_row_count"] for source in report["sources"]], [3, 3])
            self.assertEqual([source["selected_row_count"] for source in report["sources"]], [2, 2])

    def test_selected_timeout_fails_strict_evidence(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            manifest_path = self.write_manifest(root, ["slow"])
            selection = manifest_record(manifest_path, REPO_ROOT)
            rows = [good_row("fast"), timeout_row("slow")]
            cohorts = [
                self.write_cohort(root, "a", selection, rows),
                self.write_cohort(root, "b", selection, rows),
            ]
            with self.assertRaisesRegex(ValueError, "slow/bytecode Able result is not verifier-backed"):
                variance.build_scorecard_report(cohorts, 5, selection)

    def test_manifest_identity_and_full_coverage_are_enforced(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            manifest_path = self.write_manifest(root)
            selection = manifest_record(manifest_path, REPO_ROOT)
            rows = [
                good_row("fast"),
                good_row("fast_bytecode", "bytecode"),
                timeout_row("slow", timeouts=1),
            ]
            first = self.write_cohort(root, "a", selection, rows)

            changed_path = root / "changed.json"
            changed_path.write_text(
                json.dumps(
                    {
                        "kind": "external-benchmark-selection-manifest",
                        "schema_version": 1,
                        "modes": {"compiled": ["fast"], "bytecode": ["slow"]},
                    }
                ),
                encoding="utf-8",
            )
            changed = manifest_record(changed_path, REPO_ROOT)
            second = self.write_cohort(root, "b", changed, rows)
            with self.assertRaisesRegex(ValueError, "embedded selection manifest differs"):
                variance.build_scorecard_report([first, second], 5, selection)

            second = self.write_cohort(root, "c", selection, rows[:-1])
            with self.assertRaisesRegex(ValueError, "full status coverage differs"):
                variance.build_scorecard_report([first, second], 5, selection)

    def test_stdlib_state_identity_is_enforced(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            manifest_path = self.write_manifest(root)
            selection = manifest_record(manifest_path, REPO_ROOT)
            rows = [good_row("fast"), good_row("fast_bytecode", "bytecode")]
            first = self.write_cohort(root, "a", selection, rows)
            changed_state = {**STDLIB_STATE, "source_tree_sha256": "c" * 64}
            second = self.write_cohort(root, "b", selection, rows, changed_state)
            with self.assertRaisesRegex(ValueError, "stdlib source state differs"):
                variance.build_scorecard_report([first, second], 5, selection)

    def test_scoreboard_embeds_selection_without_dropping_timeout_status(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            manifest_path = self.write_manifest(root)
            selection = manifest_record(manifest_path, REPO_ROOT)
            rows = [
                good_row("fast"),
                good_row("fast_bytecode", "bytecode"),
                timeout_row("slow", timeouts=1),
            ]
            for row in rows:
                row["benchmark_contract"] = {}
            source = root / "source.json"
            source.write_text(
                json.dumps(
                    {
                        "generated_at": "test",
                        "suite": "test",
                        "benchmarks": ["fast", "fast_bytecode", "slow"],
                        "modes": ["compiled", "bytecode"],
                        "languages": ["go"],
                        "rows": rows,
                    }
                ),
                encoding="utf-8",
            )

            def compact(row: dict[str, Any], *_: Any) -> dict[str, Any]:
                return {
                    "benchmark": row["benchmark"],
                    "mode": row["mode"],
                    "able": row["able"],
                    "comparisons": row["comparisons"],
                }

            targets = {row["benchmark"]: root / f"{row['benchmark']}.able" for row in rows}
            with (
                mock.patch.object(scoreboard_tool, "canonical_benchmark_targets", return_value=targets),
                mock.patch.object(scoreboard_tool, "scorecard_source_fingerprint", return_value={}),
                mock.patch.object(scoreboard_tool, "compact_row", side_effect=compact),
            ):
                scoreboard = scoreboard_tool.build_scoreboard([source], STDLIB_STATE, selection)
            self.assertEqual(scoreboard["selection_manifest"], selection)
            self.assertEqual(len(scoreboard["rows"]), 3)
            slow = next(row for row in scoreboard["rows"] if row["benchmark"] == "slow")
            self.assertEqual(slow["able"]["timeouts"], 1)


class SelectionStatusTests(unittest.TestCase):
    def test_current_markdown_headline_counts_only_selected_rows(self) -> None:
        scoreboard = json.loads(
            (SCRIPT_DIR / "docs/perf-baselines/external-scoreboard-current.json").read_text()
        )
        selected = manifest_keys(scoreboard["selection_manifest"])
        rendered = scoreboard_tool.render_markdown(scoreboard)
        for mode in ("compiled", "bytecode"):
            ranked = [
                row for row in scoreboard["rows"]
                if (row["benchmark"], mode) in selected and row["mode"] == mode
                and row["target_status"] != "unranked"
            ]
            meets = sum(row["target_status"] == "meets" for row in ranked)
            self.assertIn(f"`{meets}/{len(ranked)}` selected rankable rows", rendered)

    def test_completed_excluded_row_is_reported_for_re_admission(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            manifest = json.loads((SCRIPT_DIR / "bench-selection-manifest.json").read_text())
            manifest["modes"]["bytecode"].remove("binarytrees")
            manifest_path = root / "selection.json"
            manifest_path.write_text(json.dumps(manifest), encoding="utf-8")
            rows = [
                {"benchmark": benchmark, "mode": mode, "able": {"status": "verified"}}
                for mode, benchmarks in manifest["modes"].items()
                for benchmark in benchmarks
            ]
            rows.append(
                {"benchmark": "binarytrees", "mode": "bytecode", "able": {"status": "verified"}}
            )
            scoreboard = root / "status.json"
            scoreboard.write_text(
                json.dumps({"kind": "external-benchmark-scoreboard", "rows": rows}),
                encoding="utf-8",
            )
            result = subprocess.run(
                [
                    str(SCRIPT_DIR / "bench_selection_manifest_check"),
                    "--manifest",
                    str(manifest_path),
                    "--status-scorecard",
                    str(scoreboard),
                ],
                cwd=REPO_ROOT,
                text=True,
                capture_output=True,
                check=False,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            self.assertIn("review for re-admission: binarytrees/bytecode", result.stdout)


if __name__ == "__main__":
    unittest.main()
