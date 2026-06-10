#!/usr/bin/env python3
"""Fast contract tests for the portable VM backend ADR."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO_ROOT = ROOT.parent
GENERATOR = ROOT / "bench_portable_vm_backend_adr"
EVIDENCE = ROOT / "bench-portable-vm-backend-adr.json"


class PortableVMBackendADRTests(unittest.TestCase):
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

    def test_current_runtime_contract_selects_no_foreign_backend(self) -> None:
        summary = self.report()["summary"]
        self.assertEqual(
            summary["decision"],
            "close-portable-foreign-backend-under-current-runtime-contract",
        )
        self.assertFalse(summary["prototype_admitted"])
        self.assertEqual(summary["selected_backend_count"], 0)
        self.assertEqual(summary["backend_class_count"], 3)
        self.assertEqual(summary["ownership_variant_count"], 6)

    def test_performance_reach_is_not_mistaken_for_abi_feasibility(self) -> None:
        summary = self.report()["summary"]
        self.assertTrue(summary["phase_one_performance_capable"])
        self.assertEqual(summary["phase_one_material_rows"], 5)
        self.assertEqual(summary["phase_one_governing_rows"], 5)
        self.assertEqual(summary["contract_eligible_hot_function_count"], 0)
        self.assertEqual(summary["maximum_static_primitive_span"], 7)

    def test_each_backend_class_checks_both_ownership_models(self) -> None:
        report = self.report()
        self.assertEqual(
            {row["id"] for row in report["backend_classes"]},
            {
                "whole-engine-c-abi",
                "portable-jit-library",
                "direct-machine-code-generation",
            },
        )
        for backend in report["backend_classes"]:
            self.assertEqual(len(backend["variants"]), 2)
            self.assertEqual(backend["admitted_variants"], [])
            self.assertEqual(backend["disposition"], "reject-under-current-runtime-contract")
            self.assertFalse(any(row["passes_all_gates"] for row in backend["variants"]))

    def test_abi_obligations_and_next_lane_are_explicit(self) -> None:
        report = self.report()
        self.assertGreaterEqual(len(report["decision_gates"]), 8)
        self.assertEqual(len(report["current_abi_observations"]), 6)
        self.assertEqual(
            report["next_lane"]["id"], "shared-runtime-semantic-abi-feasibility"
        )
        self.assertEqual(len(report["next_lane"]["required_outputs"]), 6)

    def test_checked_artifacts_are_current(self) -> None:
        result = self.generate("--check")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_source_fingerprint_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["sources"]["runtime_values"]["sha256"] = "0" * 64
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("fingerprint drifted", result.stderr)


if __name__ == "__main__":
    unittest.main()
