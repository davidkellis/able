#!/usr/bin/env python3
"""Fast contract tests for the compiled architecture target budget."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO_ROOT = ROOT.parent
GENERATOR = ROOT / "bench_compiled_architecture_budget"
EVIDENCE = ROOT / "bench-compiled-architecture-budget.json"
ALLOCATION = (
    ROOT
    / "docs/perf-baselines/2026-07-21-compiled-generated-allocation-shape.json"
)


class CompiledArchitectureBudgetTests(unittest.TestCase):
    def generate(self, *args: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["python3", str(GENERATOR), *args],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    def test_current_budget_rejects_a_candidate(self) -> None:
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        summary = report["summary"]
        self.assertEqual(summary["application_count"], 5)
        self.assertEqual(summary["unlike_family_count"], 5)
        self.assertEqual(summary["eligible_mechanisms"], [])
        self.assertEqual(
            summary["decision"], "no-go-current-compiled-architecture-mechanism"
        )
        self.assertGreater(
            summary["minimum_remaining_speedup_after_perfect_local_owner_removal"],
            1,
        )
        applications = {row["benchmark"]: row for row in report["applications"]}
        self.assertEqual(applications["k_nucleotide"]["able_samples"], 5)
        self.assertEqual(applications["fixed_width_128"]["go_samples"], 5)
        self.assertEqual(applications["distance_field"]["able_samples"], 5)
        self.assertEqual(applications["policy_record_dispatch"]["go_samples"], 5)
        self.assertEqual(applications["concurrent_event_routing"]["able_samples"], 5)
        self.assertAlmostEqual(summary["minimum_required_speedup"], 6.416609, places=6)
        self.assertAlmostEqual(summary["maximum_required_speedup"], 492.565312, places=6)

    def test_source_fingerprint_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["sources"]["architecture_audit"]["sha256"] = "0" * 64
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("fingerprint drifted", result.stderr)

    def test_current_timing_fingerprint_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["current_timing_sources"]["k_nucleotide"]["able"][0][
                "sha256"
            ] = "0" * 64
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("fingerprint drifted", result.stderr)

    def test_three_material_families_admit_a_mechanism(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            allocation = json.loads(ALLOCATION.read_text(encoding="utf-8"))
            mechanism = next(
                item
                for item in allocation["mechanisms"]
                if item["id"] == "builtin-string-conversion"
            )
            event = mechanism["applications"].index("concurrent_event_routing")
            mechanism["cpu_cumulative_percent"][event] = 1.0
            path = Path(raw_dir) / "allocation.json"
            path.write_text(json.dumps(allocation), encoding="utf-8")
            result = self.generate("--allocation-report", str(path))
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(
            report["summary"]["eligible_mechanisms"],
            ["builtin-string-conversion"],
        )
        self.assertEqual(report["summary"]["decision"], "candidate-eligible")


if __name__ == "__main__":
    unittest.main()
