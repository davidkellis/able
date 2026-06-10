#!/usr/bin/env python3
"""Fast contract tests for the bytecode semantic-work amplification audit."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
GENERATOR = SCRIPT_DIR / "bench_semantic_work_amplification"
EVIDENCE = SCRIPT_DIR / "bench-semantic-work-amplification.json"


class SemanticWorkAmplificationTests(unittest.TestCase):
    def generate(self, evidence: Path = EVIDENCE) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [str(GENERATOR), "--evidence", str(evidence)],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    def write_evidence(self, directory: str, value: dict[str, object]) -> Path:
        path = Path(directory) / "evidence.json"
        path.write_text(json.dumps(value), encoding="utf-8")
        return path

    def test_current_evidence_rejects_a_candidate_from_the_six_rows(self) -> None:
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        summary = report["summary"]
        self.assertEqual(summary["decision"], "no-go-current-six-semantic-candidate")
        self.assertEqual(summary["application_count"], 6)
        self.assertEqual(summary["unlike_family_count"], 6)
        self.assertEqual(summary["total_dynamic_operations"], 246759975)
        self.assertEqual(summary["total_transport_operations"], 89338836)
        self.assertEqual(summary["total_semantic_operations"], 157421139)
        self.assertEqual(summary["total_allocations"], 65842215)
        self.assertEqual(summary["eligible_mechanism_count"], 0)

    def test_category_sum_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["applications"][0]["category_counts"]["transport"] += 1
            path = self.write_evidence(directory, evidence)
            result = self.generate(path)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("category counts do not sum", result.stderr)

    def test_derived_logical_unit_drift_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            word = next(
                app
                for app in evidence["applications"]
                if app["benchmark"] == "word_frequency"
            )
            word["logical_units"] += 1
            path = self.write_evidence(directory, evidence)
            result = self.generate(path)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("logical units drifted", result.stderr)

    def test_unmarked_admitted_mechanism_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as directory:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            mechanism = evidence["mechanisms"][0]
            mechanism["material_amplification_applications"] = mechanism[
                "applications"
            ]
            path = self.write_evidence(directory, evidence)
            result = self.generate(path)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("clears admission but is not marked admitted", result.stderr)


if __name__ == "__main__":
    unittest.main()
