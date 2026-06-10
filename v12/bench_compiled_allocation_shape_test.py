#!/usr/bin/env python3

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
GENERATOR = ROOT / "bench_compiled_allocation_shape"
EVIDENCE = ROOT / "bench-compiled-allocation-shape.json"


class CompiledAllocationShapeTests(unittest.TestCase):
    def run_generator(self, evidence: Path = EVIDENCE):
        return subprocess.run(
            [str(GENERATOR), "--evidence", str(evidence)],
            cwd=ROOT.parent,
            text=True,
            capture_output=True,
            check=False,
        )

    def test_current_evidence_rejects_a_candidate(self):
        result = self.run_generator()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(report["summary"]["application_count"], 5)
        self.assertEqual(report["summary"]["unlike_family_count"], 5)
        self.assertEqual(report["summary"]["eligible_mechanism_count"], 0)
        self.assertEqual(
            report["summary"]["decision"],
            "no-go-no-three-family-cpu-material-allocation-shape",
        )

    def test_source_drift_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["applications"][0]["able_source_sha256"] = "0" * 64
            path = Path(directory) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.run_generator(path)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("fingerprint drifted", result.stderr)

    def test_single_process_evidence_is_rejected(self):
        with tempfile.TemporaryDirectory() as directory:
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["applications"][0]["able_allocation_samples"] = [[1, 1]]
            path = Path(directory) / "evidence.json"
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.run_generator(path)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("at least two processes", result.stderr)


if __name__ == "__main__":
    unittest.main()
