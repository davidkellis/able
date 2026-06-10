#!/usr/bin/env python3
"""Contract tests for bytecode semantic-region feasibility reconciliation."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO_ROOT = ROOT.parent
GENERATOR = ROOT / "bench_bytecode_semantic_region_feasibility"
EVIDENCE = ROOT / "bench-bytecode-semantic-region-feasibility.json"


class BytecodeSemanticRegionFeasibilityTests(unittest.TestCase):
    def generate(self, *args: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["python3", str(GENERATOR), *args],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    def test_current_region_tier_is_closed(self) -> None:
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        summary = report["summary"]
        self.assertEqual(summary["decision"], "no-go-current-semantic-region-tier")
        self.assertFalse(summary["candidate_eligible"])
        self.assertEqual(summary["material_region_family_count"], 5)
        self.assertEqual(summary["rankable_current_region_rows"], 4)
        self.assertEqual(summary["uniform_cost_model_target_closures"], 1)
        self.assertEqual(summary["prototype_application_count"], 3)
        self.assertEqual(summary["prototype_regression_count"], 3)
        self.assertEqual(summary["current_semantic_candidate_count"], 0)
        self.assertGreater(summary["minimum_remaining_speedup_after_free_transport"], 7)
        bounds = {row["benchmark"]: row for row in report["typed_region_bounds"]}
        self.assertTrue(bounds["monte_carlo_pi"]["uniform_cost_model_reaches_target"])
        self.assertGreater(
            bounds["fixed_width_128"]["remaining_speedup_in_uniform_cost_model"],
            20,
        )
        self.assertEqual(
            report["next_architecture_lane"]["id"],
            "bytecode-native-hot-code-tier-design",
        )

    def test_source_fingerprint_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["sources"]["typed_region_gate"]["sha256"] = "0" * 64
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("fingerprint drifted", result.stderr)

    def test_rejected_prototype_cannot_be_relabelled_as_kept(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["typed_region_census"]["prototype"]["kept"] = True
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("rejected result", result.stderr)

    def test_checked_artifacts_are_current(self) -> None:
        result = self.generate("--check")
        self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == "__main__":
    unittest.main()
