#!/usr/bin/env python3
"""Fast contract tests for shared runtime semantic-ABI feasibility."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO_ROOT = ROOT.parent
GENERATOR = ROOT / "bench_shared_runtime_semantic_abi"
EVIDENCE = ROOT / "bench-shared-runtime-semantic-abi.json"


class SharedRuntimeSemanticABITests(unittest.TestCase):
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

    def test_feasibility_admits_only_codec_layout_spike(self) -> None:
        summary = self.report()["summary"]
        self.assertEqual(
            summary["decision"], "conditionally-feasible-admit-codec-layout-spike"
        )
        self.assertTrue(summary["feasibility_passes"])
        self.assertTrue(summary["codec_layout_spike_admitted"])
        self.assertFalse(summary["runtime_or_backend_implementation_admitted"])
        self.assertEqual(summary["program_image_section_count"], 8)
        self.assertEqual(summary["current_live_opcode_count"], 144)

    def test_runtime_kind_mapping_is_exhaustive(self) -> None:
        mapping = self.report()["runtime_kind_mapping"]
        self.assertTrue(mapping["exhaustive"])
        self.assertEqual(mapping["runtime_kind_count"], 31)
        self.assertEqual(mapping["mapped_kind_count"], 31)
        self.assertEqual(mapping["duplicates"], [])
        self.assertEqual(mapping["missing"], [])
        self.assertEqual(mapping["unknown"], [])
        self.assertEqual(
            mapping["group_counts"],
            {
                "immediate_or_immediate_with_heap_overflow": 7,
                "shared_heap_or_immutable_metadata": 20,
                "host_effect_registry": 4,
            },
        )

    def test_all_serial_target_model_rows_clear_materiality(self) -> None:
        report = self.report()
        summary = report["summary"]
        self.assertEqual(summary["target_model_row_count"], 5)
        self.assertEqual(summary["target_model_material_row_count"], 5)
        self.assertTrue(all(row["passes_25_percent_gate"] for row in report["target_model"]))
        self.assertEqual(
            {row["benchmark"] for row in report["target_model"]},
            {
                "fixed_width_128",
                "distance_field",
                "word_frequency",
                "array_slice_window",
                "reverse_complement",
            },
        )

    def test_ownership_effect_fallback_and_migration_are_explicit(self) -> None:
        report = self.report()
        self.assertEqual(report["value_cell"]["size_bytes"], 16)
        self.assertEqual(len(report["ownership"]), 6)
        self.assertEqual(len(report["effect_resume"]["host_effects"]), 5)
        self.assertEqual(len(report["migration"]), 4)
        self.assertFalse(report["migration"][0]["production_execution_change"])
        self.assertEqual(report["next_lane"]["id"], "semantic-abi-codec-layout-spike")
        self.assertEqual(len(report["next_lane"]["governing_applications"]), 3)

    def test_checked_artifacts_are_current(self) -> None:
        result = self.generate("--check")
        self.assertEqual(result.returncode, 0, result.stderr)

    def test_kind_source_or_evidence_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["runtime_kind_groups"]["host_effect_registry"].remove("KindFuture")
            path = Path(raw_dir) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate("--evidence", str(path))
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("runtime kind mapping is not exhaustive", result.stderr)


if __name__ == "__main__":
    unittest.main()
