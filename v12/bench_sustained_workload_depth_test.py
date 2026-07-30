#!/usr/bin/env python3
"""Contract tests for the sustained multi-feature workload-depth audit."""

from __future__ import annotations

import importlib.machinery
import importlib.util
import json
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
GENERATOR = SCRIPT_DIR / "bench_sustained_workload_depth"
EVIDENCE = (
    SCRIPT_DIR
    / "docs/perf-baselines/2026-07-30-sustained-multi-feature-workload-depth-audit.json"
)


def load_generator():
    loader = importlib.machinery.SourceFileLoader("workload_depth", str(GENERATOR))
    spec = importlib.util.spec_from_loader(loader.name, loader)
    if spec is None:
        raise RuntimeError("cannot load workload-depth generator")
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


class WorkloadDepthTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.generator = load_generator()

    def build(self):
        return self.generator.build_report(
            SCRIPT_DIR / "bench-feature-coverage.json",
            SCRIPT_DIR / "bench-feature-interaction-priorities.json",
            SCRIPT_DIR / "bench-operation-depth.json",
            SCRIPT_DIR / "docs/perf-baselines/external-scoreboard-current.json",
            SCRIPT_DIR
            / "docs/perf-baselines/2026-07-20-cross-mode-performance-frontier.json",
            0.2,
        )

    def test_current_gap_is_exact(self) -> None:
        report = self.build()
        self.assertEqual(report["summary"]["application_count"], 66)
        self.assertEqual(report["summary"]["discriminating_feature_family_count"], 11)
        self.assertEqual(report["policy"]["broad_feature_family_minimum"], 6)
        self.assertEqual(report["summary"]["broad_multi_feature_count"], 24)
        self.assertEqual(report["summary"]["sustained_native_count"], 11)
        self.assertEqual(report["summary"]["intersection_count"], 1)
        self.assertFalse(report["summary"]["material_depth_gap"])
        self.assertEqual(
            report["summary"]["maximum_go_seconds_among_broad"],
            max(
                item["go_seconds"]
                for item in report["applications"]
                if item["broad_multi_feature"]
            ),
        )
        self.assertEqual(report["summary"]["maximum_feature_families_among_sustained"], 10)
        self.assertEqual(
            report["summary"]["admitted_existing_applications"],
            ["versioned_telemetry_pipeline"],
        )

    def test_checked_evidence_is_exact(self) -> None:
        self.assertEqual(
            json.loads(EVIDENCE.read_text(encoding="utf-8")),
            self.build(),
        )

    def test_duration_uses_go_not_able(self) -> None:
        report = self.build()
        target = next(
            item
            for item in report["applications"]
            if item["broad_multi_feature"] and not item["sustained_native_work"]
        )
        scorecard = json.loads(
            (
                SCRIPT_DIR / "docs/perf-baselines/external-scoreboard-current.json"
            ).read_text(encoding="utf-8")
        )
        compiled = next(
            row
            for row in scorecard["rows"]
            if row["mode"] == "compiled" and row["benchmark"] == target["benchmark"]
        )
        compiled["able"]["real_seconds"] = 10.0
        compiled["comparisons"]["go"]["real_seconds"] = 0.001
        with tempfile.TemporaryDirectory() as raw:
            scorecard_path = Path(raw) / "scorecard.json"
            scorecard_path.write_text(json.dumps(scorecard), encoding="utf-8")
            mutated = self.generator.build_report(
                SCRIPT_DIR / "bench-feature-coverage.json",
                SCRIPT_DIR / "bench-feature-interaction-priorities.json",
                SCRIPT_DIR / "bench-operation-depth.json",
                scorecard_path,
                SCRIPT_DIR
                / "docs/perf-baselines/2026-07-20-cross-mode-performance-frontier.json",
                0.2,
            )
        mutated_target = next(
            item
            for item in mutated["applications"]
            if item["benchmark"] == target["benchmark"]
        )
        self.assertEqual(mutated_target["able_compiled_seconds"], 10.0)
        self.assertEqual(mutated_target["go_seconds"], 0.001)
        self.assertTrue(mutated_target["broad_multi_feature"])
        self.assertFalse(mutated_target["sustained_native_work"])

    def test_invalid_threshold_fails(self) -> None:
        with self.assertRaisesRegex(ValueError, "positive and finite"):
            self.generator.build_report(
                SCRIPT_DIR / "bench-feature-coverage.json",
                SCRIPT_DIR / "bench-feature-interaction-priorities.json",
                SCRIPT_DIR / "bench-operation-depth.json",
                SCRIPT_DIR / "docs/perf-baselines/external-scoreboard-current.json",
                SCRIPT_DIR
                / "docs/perf-baselines/2026-07-20-cross-mode-performance-frontier.json",
                0,
            )

    def test_missing_feature_membership_fails_closed(self) -> None:
        with tempfile.TemporaryDirectory() as raw:
            temporary = Path(raw)
            source = SCRIPT_DIR / "bench-feature-coverage.json"
            manifest = json.loads(source.read_text(encoding="utf-8"))
            for family in manifest["families"]:
                family["portable_benchmarks"] = [
                    item
                    for item in family["portable_benchmarks"]
                    if item != "fib"
                ]
            coverage = temporary / "coverage.json"
            coverage.write_text(json.dumps(manifest), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "fib"):
                self.generator.build_report(
                    coverage,
                    SCRIPT_DIR / "bench-feature-interaction-priorities.json",
                    SCRIPT_DIR / "bench-operation-depth.json",
                    SCRIPT_DIR
                    / "docs/perf-baselines/external-scoreboard-current.json",
                    SCRIPT_DIR
                    / "docs/perf-baselines/2026-07-20-cross-mode-performance-frontier.json",
                    0.2,
                )


if __name__ == "__main__":
    unittest.main()
