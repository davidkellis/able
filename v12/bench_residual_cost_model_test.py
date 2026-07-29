#!/usr/bin/env python3
"""Fast contract tests for the cross-mode residual cost model."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
GENERATOR = SCRIPT_DIR / "bench_residual_cost_model"
EVIDENCE = SCRIPT_DIR / "bench-residual-cost-evidence.json"


class ResidualCostModelTests(unittest.TestCase):
    def generate(self, evidence: Path = EVIDENCE) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [str(GENERATOR), "--evidence", str(evidence)],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    def test_current_model_reconciles_and_selects_no_leaf(self) -> None:
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        summary = report["summary"]
        self.assertEqual(summary["application_count"], 5)
        self.assertEqual(summary["unlike_family_count"], 5)
        self.assertEqual(summary["selected_mode_rows"], 10)
        self.assertAlmostEqual(summary["selected_excess_seconds"], 66.363684, places=6)
        self.assertAlmostEqual(
            summary["frontier_total_excess_seconds"], 226.856947, places=6
        )
        self.assertAlmostEqual(
            summary["selected_excess_share_percent"], 29.253538, places=6
        )
        self.assertEqual(summary["eligible_mechanisms"], [])
        self.assertEqual(summary["decision"], "no-candidate")
        self.assertEqual(
            summary["dynamic_stats_status_counts"],
            {"guard-capped": 1, "guard-skipped": 1, "measured": 3},
        )
        self.assertEqual(
            summary["compiled_telemetry_totals"]["generic_union_fallback"], 0
        )

    def test_unknown_mechanism_application_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            path = Path(raw_dir) / "evidence.json"
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["mechanisms"][0]["applications"].append("missing")
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate(path)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("unknown application", result.stderr)

    def test_open_three_family_mechanism_becomes_eligible(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            path = Path(raw_dir) / "evidence.json"
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["mechanisms"][0]["status"] = "open"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate(path)
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(
            report["summary"]["eligible_mechanisms"],
            ["bytecode-stack-slot-transport"],
        )
        self.assertEqual(report["summary"]["decision"], "candidate-eligible")


if __name__ == "__main__":
    unittest.main()
