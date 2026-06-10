#!/usr/bin/env python3
"""Fast contract tests for the shared-runtime closed-region decision."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO_ROOT = ROOT.parent
GENERATOR = ROOT / "bench_shared_runtime_closed_region_cutover"
EVIDENCE = ROOT / "bench-shared-runtime-closed-region-cutover.json"


class SharedRuntimeClosedRegionCutoverTests(unittest.TestCase):
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

    def test_production_migration_closes_on_breadth_not_performance(self) -> None:
        summary = self.report()["summary"]
        self.assertEqual(
            summary["decision"],
            "close-shared-runtime-production-migration-no-three-unlike-closed-cut",
        )
        self.assertEqual(summary["target_model_row_count"], 5)
        self.assertEqual(summary["target_model_material_row_count"], 5)
        self.assertEqual(summary["bounded_closed_region_candidate_count"], 1)
        self.assertEqual(summary["bounded_closed_region_unlike_family_count"], 1)
        self.assertEqual(summary["minimum_unlike_families"], 3)
        self.assertFalse(summary["breadth_gate_passes"])
        self.assertFalse(summary["implementation_admitted"])

    def test_distance_is_only_bounded_function_entry_cut(self) -> None:
        report = self.report()
        bounded = [
            row
            for row in report["applications"]
            if row["bounded_closed_region_candidate"]
        ]
        self.assertEqual(
            [(row["benchmark"], row["family"]) for row in bounded],
            [("distance_field", "float-numeric")],
        )
        self.assertFalse(bounded[0]["crosses_legacy_identity_boundary"])
        self.assertEqual(
            bounded[0]["minimum_closed_scope"],
            "function-plus-transitive-primitive-kernel",
        )

    def test_identity_rows_expand_to_interpreter_instance(self) -> None:
        report = self.report()
        wholesale = [
            row
            for row in report["applications"]
            if row["minimum_closed_scope"] == "interpreter-instance-runtime"
        ]
        self.assertEqual(len(wholesale), 4)
        self.assertTrue(
            all(row["crosses_legacy_identity_boundary"] for row in wholesale)
        )
        self.assertEqual(
            {row["benchmark"] for row in wholesale},
            {
                "fixed_width_128",
                "word_frequency",
                "array_slice_window",
                "reverse_complement",
            },
        )

    def test_static_ownership_and_frontier_are_complete(self) -> None:
        report = self.report()
        self.assertEqual(report["summary"]["ownership_domain_count"], 7)
        self.assertEqual(len(report["ownership_budget"]), 7)
        self.assertEqual(
            report["next_lane"]["frontier_families"],
            [
                "concurrency",
                "control_flow",
                "expressions_arrays_text_files",
            ],
        )
        self.assertEqual(report["next_lane"]["frontier_current_depth"], 3)
        self.assertFalse(report["summary"]["wasm_change_retained"])

    def test_checked_artifacts_are_current(self) -> None:
        result = self.generate("--check")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_source_or_scope_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["applications"][1]["source_markers"].append(
                "not a real distance-field source marker"
            )
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("source markers missing", result.stderr)


if __name__ == "__main__":
    unittest.main()
