#!/usr/bin/env python3
"""Fast contract tests for the bytecode architecture target-budget model."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
GENERATOR = SCRIPT_DIR / "bench_bytecode_architecture_budget"
EVIDENCE = SCRIPT_DIR / "bench-bytecode-architecture-budget.json"


class BytecodeArchitectureBudgetTests(unittest.TestCase):
    def generate(self, evidence: Path = EVIDENCE) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [str(GENERATOR), "--evidence", str(evidence)],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    def test_current_budget_closes_register_representation_only_route(self) -> None:
        result = self.generate()
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        summary = report["summary"]
        self.assertEqual(summary["decision"], "no-go-register-representation-only")
        self.assertEqual(summary["application_count"], 6)
        self.assertEqual(summary["unlike_family_count"], 6)
        self.assertEqual(summary["live_opcode_count"], 144)
        self.assertEqual(summary["transport_opcode_count"], 6)
        self.assertEqual(summary["semantic_opcode_count"], 138)
        self.assertEqual(summary["semantic_closure_area_count"], 8)
        self.assertEqual(summary["total_dynamic_operations"], 246759975)
        self.assertEqual(summary["total_transport_operations"], 89338836)
        self.assertAlmostEqual(
            summary["weighted_transport_share_percent"], 36.204752, places=6
        )
        self.assertGreater(summary["minimum_remaining_speedup_after_transport"], 7)
        self.assertLess(summary["maximum_uniform_cost_transport_speedup"], 2)
        self.assertTrue(summary["all_empirical_gates_rejected"])

    def test_missing_live_opcode_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            path = Path(raw_dir) / "evidence.json"
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["opcode_categories"][0]["opcodes"].remove("Pop")
            evidence["transport_opcodes"].remove("Pop")
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate(path)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("missing Pop", result.stderr)

    def test_duplicate_opcode_category_is_rejected(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            path = Path(raw_dir) / "evidence.json"
            evidence = json.loads(EVIDENCE.read_text(encoding="utf-8"))
            evidence["opcode_categories"][1]["opcodes"].append("Const")
            path.write_text(json.dumps(evidence), encoding="utf-8")
            result = self.generate(path)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("appears in both", result.stderr)


if __name__ == "__main__":
    unittest.main()
