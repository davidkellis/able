#!/usr/bin/env python3
"""Contract tests for the bytecode native hot-tier design budget."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO_ROOT = ROOT.parent
GENERATOR = ROOT / "bench_bytecode_native_hot_tier_budget"
EVIDENCE = ROOT / "bench-bytecode-native-hot-tier-design.json"


class BytecodeNativeHotTierBudgetTests(unittest.TestCase):
    def generate(self, *args: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            ["python3", str(GENERATOR), *args],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    def test_current_design_does_not_admit_a_prototype(self) -> None:
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        summary = report["summary"]
        self.assertEqual(
            summary["decision"],
            "no-go-native-tier-prototype-current-evidence",
        )
        self.assertFalse(summary["prototype_admitted"])
        self.assertEqual(summary["common_bytecode_compiled_application_count"], 66)
        self.assertEqual(summary["compiled_native_proxy_target_meets"], 36)
        self.assertEqual(summary["compiled_native_proxy_target_misses"], 30)
        self.assertAlmostEqual(
            summary["compiled_native_proxy_target_excess_seconds"],
            16.818105262,
            places=9,
        )
        self.assertGreater(
            summary["compiled_native_proxy_excess_reduction_percent"], 75
        )
        self.assertEqual(summary["known_region_application_count"], 4)
        self.assertEqual(summary["known_region_material_reduction_count"], 2)
        self.assertEqual(summary["known_region_target_closure_count"], 0)
        self.assertEqual(summary["hot_function_census_application_count"], 6)
        self.assertEqual(summary["hot_function_contract_eligible_count"], 0)
        self.assertEqual(summary["scalar_proof_census_application_count"], 6)
        self.assertEqual(summary["proven_integer_load_application_count"], 5)
        self.assertTrue(summary["scalar_slot_hint_rejected"])
        self.assertTrue(summary["array_slot_value_role_candidate_closed"])
        self.assertEqual(summary["eligible_hot_function_class_count"], 0)
        self.assertEqual(summary["selected_backend_count"], 0)
        regions = {row["benchmark"]: row for row in report["known_region_budget"]}
        self.assertGreater(
            regions["fixed_width_128"][
                "required_compiled_equivalent_fraction_to_target"
            ],
            0.96,
        )
        self.assertGreater(
            regions["rms_norm"]["required_compiled_equivalent_fraction_to_target"],
            0.84,
        )
        self.assertTrue(regions["monte_carlo_pi"]["material_reduction_in_model"])

    def test_source_fingerprint_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["sources"]["vm_dispatch"]["sha256"] = "0" * 64
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("fingerprint drifted", result.stderr)

    def test_three_family_reach_without_backend_does_not_admit(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            by_name = {
                row["benchmark"]: row
                for row in evidence["hot_function_reach_evidence"]
            }
            evidence["hot_function_reach_evidence"] = []
            for idx, benchmark in enumerate(
                ("fixed_width_128", "distance_field", "word_frequency")
            ):
                row = dict(by_name[benchmark])
                row["family"] = f"family-{idx}"
                row["candidate_class"] = "typed-hot-function"
                row["contract_eligible"] = True
                evidence["hot_function_reach_evidence"].append(row)
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(report["summary"]["eligible_hot_function_class_count"], 1)
        self.assertFalse(report["summary"]["prototype_admitted"])

    def test_checked_artifacts_are_current(self) -> None:
        result = self.generate("--check")
        self.assertEqual(result.returncode, 0, result.stderr)


if __name__ == "__main__":
    unittest.main()
