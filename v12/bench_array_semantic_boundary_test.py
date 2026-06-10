#!/usr/bin/env python3
"""Contract tests for the bytecode Array semantic-boundary gate."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
GENERATOR = SCRIPT_DIR / "bench_array_semantic_boundary"
EVIDENCE = SCRIPT_DIR / "bench-array-semantic-boundary.json"


class ArraySemanticBoundaryTests(unittest.TestCase):
    def generate(self, evidence: Path = EVIDENCE) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [str(GENERATOR), "--evidence", str(evidence)],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    def test_current_gate_closes_array_generalization(self) -> None:
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        summary = report["summary"]
        self.assertEqual(summary["decision"], "no-go-array-wrapper-or-storage-generalization")
        self.assertEqual(summary["application_count"], 3)
        self.assertEqual(summary["unlike_family_count"], 3)
        self.assertEqual(summary["total_array_slot_lookups"], 12959267)
        self.assertEqual(summary["total_push_lookups"], 8325866)
        self.assertEqual(summary["total_fast_path_misses"], 0)
        self.assertEqual(summary["distinct_dominant_push_descendants"], 3)
        self.assertTrue(summary["prior_generic_candidate_rejected"])

    def test_member_kind_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["applications"][0]["array_slot"]["push"] += 1
            path = Path(directory) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate(path)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("member-kind counts do not sum", result.stderr)

    def test_fast_path_miss_is_not_silently_accepted(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["applications"][1]["array_slot"]["fast_path_misses"] = 1
            path = Path(directory) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate(path)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("no fast miss", result.stderr)


if __name__ == "__main__":
    unittest.main()
