#!/usr/bin/env python3
"""Fast contract tests for the cross-engine architecture budget."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO_ROOT = ROOT.parent
GENERATOR = ROOT / "bench_cross_engine_architecture_budget"
EVIDENCE = ROOT / "bench-cross-engine-architecture-budget.json"


class CrossEngineArchitectureBudgetTests(unittest.TestCase):
    def generate(self, *args: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["python3", str(GENERATOR), *args],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    def test_current_budget_selects_no_local_candidate(self) -> None:
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        summary = report["summary"]
        self.assertEqual(
            summary["decision"], "no-go-current-cross-engine-local-mechanism"
        )
        self.assertFalse(summary["candidate_eligible"])
        self.assertEqual(summary["selected_rows"], 126)
        self.assertEqual(summary["target_meets"], 10)
        self.assertEqual(summary["target_misses"], 116)
        self.assertAlmostEqual(summary["total_excess_seconds"], 227.17905263157894)
        self.assertGreater(summary["bytecode_excess_share_percent"], 80)
        bounds = report["architecture_bounds"]
        self.assertGreater(bounds["bytecode"]["minimum_remaining_target_speedup"], 7)
        self.assertGreater(bounds["compiled"]["minimum_remaining_target_speedup"], 3)
        self.assertEqual(
            report["group_budgets"][0]["id"],
            "bytecode-portable-workload-admission",
        )
        self.assertEqual(
            report["next_architecture_lane"]["id"],
            "cross-engine-structural-strategy-reconciliation",
        )

    def test_group_and_mode_excess_reconcile(self) -> None:
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        group_total = sum(row["total_excess_seconds"] for row in report["group_budgets"])
        mode_total = sum(row["total_excess_seconds"] for row in report["modes"])
        self.assertAlmostEqual(group_total, mode_total)
        self.assertAlmostEqual(mode_total, report["summary"]["total_excess_seconds"])

    def test_checked_artifacts_are_current(self) -> None:
        result = self.generate("--check")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_source_fingerprint_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["sources"]["frontier"]["sha256"] = "0" * 64
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("fingerprint drifted", result.stderr)


if __name__ == "__main__":
    unittest.main()
