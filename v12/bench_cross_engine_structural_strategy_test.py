#!/usr/bin/env python3
"""Fast contract tests for cross-engine structural-strategy reconciliation."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO_ROOT = ROOT.parent
GENERATOR = ROOT / "bench_cross_engine_structural_strategy"
EVIDENCE = ROOT / "bench-cross-engine-structural-strategy.json"


class CrossEngineStructuralStrategyTests(unittest.TestCase):
    def generate(self, *args: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["python3", str(GENERATOR), *args],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    def report(self) -> dict[str, object]:
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        return json.loads(result.stdout)

    def test_only_portable_backend_has_sufficient_modeled_reach(self) -> None:
        report = self.report()
        summary = report["summary"]
        self.assertEqual(
            summary["decision"],
            "portable-backend-performance-capable-but-no-concrete-mechanism",
        )
        self.assertEqual(summary["performance_capable_route_count"], 1)
        self.assertEqual(summary["concrete_admitted_route_count"], 0)
        self.assertFalse(summary["prototype_admitted"])
        self.assertEqual(summary["selected_rows"], 128)
        self.assertEqual(summary["target_misses"], 117)
        self.assertAlmostEqual(summary["total_target_excess_seconds"], 230.8515789473684)

    def test_typed_specialization_fails_every_row_gate(self) -> None:
        route = self.report()["routes"]["typed_bytecode_semantic_specialization"]
        self.assertEqual(route["material_row_count"], 3)
        self.assertEqual(route["governing_row_count"], 6)
        self.assertFalse(route["passes_every_governing_row"])
        failed = {row["benchmark"] for row in route["rows"] if not row["passes_materiality_gate"]}
        self.assertEqual(
            failed,
            {"concurrent_event_routing", "word_frequency", "reverse_complement"},
        )

    def test_portable_backend_proxy_passes_phase_one_but_is_not_concrete(self) -> None:
        route = self.report()["routes"]["portable_lower_level_vm_backend"]
        self.assertEqual(route["phase_one_material_row_count"], 5)
        self.assertEqual(route["phase_one_governing_row_count"], 5)
        self.assertEqual(route["full_corpus_material_row_count"], 6)
        self.assertEqual(route["full_corpus_row_count"], 6)
        self.assertTrue(route["passes_phase_one_performance_gate"])
        self.assertFalse(route["concrete_mechanism"])
        self.assertFalse(route["prototype_admitted"])
        concurrency = next(
            row for row in route["rows"] if row["benchmark"] == "concurrent_event_routing"
        )
        self.assertFalse(concurrency["phase_one_governing"])
        self.assertGreater(concurrency["target_excess_reduction_percent"], 25)

    def test_compiled_nominal_route_has_no_three_family_mechanism(self) -> None:
        route = self.report()["routes"]["compiled_general_nominal_abi_simplification"]
        self.assertEqual(route["maximum_material_unlike_family_count"], 2)
        self.assertEqual(route["eligible_mechanism_count"], 0)
        self.assertEqual(route["typed_boundary_intersection_application_count"], 3)
        self.assertFalse(route["passes_breadth_gate"])

    def test_closures_guards_and_next_lane_are_explicit(self) -> None:
        report = self.report()
        self.assertEqual(report["summary"]["closure_count"], 23)
        self.assertEqual(report["summary"]["invalidated_closure_count"], 0)
        self.assertEqual(len(report["semantic_obligations"]), 7)
        self.assertEqual(
            {row["benchmark"] for row in report["target_guards"]}, {"json", "pidigits"}
        )
        self.assertEqual(
            report["next_lane"]["id"], "portable-vm-backend-abi-dependency-adr"
        )

    def test_checked_artifacts_are_current(self) -> None:
        result = self.generate("--check")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_source_fingerprint_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["sources"]["native_tier_budget"]["sha256"] = "0" * 64
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("fingerprint drifted", result.stderr)


if __name__ == "__main__":
    unittest.main()
