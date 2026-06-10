#!/usr/bin/env python3

import json
import tempfile
import unittest
from pathlib import Path

import bench_preserved_compiled_report as report


class PreservedCompiledReportTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.root = Path(self.temp.name)
        self.source = self.root / "main.able"
        self.source.write_text("fn main() { 0 }\n")
        self.binary = self.root / "compiled"
        self.binary.write_bytes(b"binary")
        self.generated = self.root / "generated"
        self.generated.mkdir()
        (self.generated / "main.go").write_text("package main\n")
        self.contract = {
            "mode": "compiled",
            "logical_cpu_budget": 1,
            "cpu_affinity": "3",
            "executor_policy": "serial",
        }
        self.go_reference = self.root / "go.json"
        self.go_reference.write_text(json.dumps({
            "go_version": "go test",
            "rows": [{
                "benchmark": "sample",
                "source": "/tmp/app.go",
                "source_sha256": "a" * 64,
                "avg_real_seconds": 1.0,
                "execution_contract": self.contract,
            }],
        }))

    def tearDown(self):
        self.temp.cleanup()

    def cohort(self, name, values, order_index):
        path = self.root / name
        samples = [
            {"run": index, "status": "ok", "real_seconds": value, "stdout_sha256": "b" * 64}
            for index, value in enumerate(values, 1)
        ]
        path.write_text(json.dumps({
            "preserved_order_index": order_index,
            "results": [{
                "mode": "compiled", "ok_runs": len(values), "timeouts": 0, "failures": 0,
                "samples": samples,
                "validation": {
                    "status": "verified", "verified_runs": len(values), "failed_runs": 0,
                    "stdout_sha256": ["b" * 64],
                },
            }],
        }))
        return str(path)

    def manifest(self, first, second, limit=15.0):
        return {
            "suite": "test", "benchmark_repo": "/tmp/benchmarks", "cpu_affinity_pool": "3",
            "runs_per_cohort": 2, "cohort_count": 2, "max_cohort_spread_percent": limit,
            "go_reference_json": str(self.go_reference),
            "benchmarks": [{
                "benchmark": "sample", "target": str(self.source), "binary": str(self.binary),
                "generated_dir": str(self.generated), "compiled_build_args": [],
                "benchmark_contract": {"sha256": "c" * 64},
                "execution_contract": self.contract, "cohort_results": [first, second],
            }],
        }

    def test_balanced_verified_cohorts_are_eligible(self):
        payload = report.compose(self.manifest(
            self.cohort("a.json", [2.0, 2.2], 0),
            self.cohort("b.json", [2.1, 2.3], 0),
        ))
        row = payload["rows"][0]
        self.assertTrue(payload["build_phase_completed_before_timing"])
        self.assertTrue(row["aggregate"]["promotion_eligible"])
        self.assertAlmostEqual(row["aggregate"]["mean_real_seconds"], 2.15)
        self.assertAlmostEqual(row["go_reference"]["able_to_go_ratio"], 2.15)

    def test_material_cohort_disagreement_rejects_row(self):
        payload = report.compose(self.manifest(
            self.cohort("a.json", [1.0, 1.0], 0),
            self.cohort("b.json", [1.4, 1.4], 0),
            limit=15.0,
        ))
        row = payload["rows"][0]
        self.assertFalse(row["aggregate"]["promotion_eligible"])
        self.assertAlmostEqual(row["aggregate"]["cohort_spread_percent"], 40.0)
        self.assertIn("cohort_spread_exceeds_limit", row["rejection_reasons"])


if __name__ == "__main__":
    unittest.main()
