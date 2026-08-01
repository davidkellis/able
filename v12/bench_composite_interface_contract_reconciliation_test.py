#!/usr/bin/env python3
"""Fast checks for the composite-interface performance evidence rebase."""

from __future__ import annotations

import importlib.machinery
import importlib.util
import json
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
SCRIPT = ROOT / "bench_composite_interface_contract_reconciliation"


def load_module():
    loader = importlib.machinery.SourceFileLoader(
        "composite_interface_contract_reconciliation", str(SCRIPT)
    )
    spec = importlib.util.spec_from_loader(loader.name, loader)
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


class CompositeInterfaceContractReconciliationTest(unittest.TestCase):
    def test_checked_evidence_covers_every_closure(self) -> None:
        module = load_module()
        report = module.build_report()
        self.assertEqual(report["decision"], "rebaseline-reviewed-v12-spec-scope")
        self.assertEqual(report["summary"]["closure_count"], 23)
        self.assertEqual(report["summary"]["compiled_rows"], 10)
        self.assertEqual(report["summary"]["bytecode_rows"], 9)
        self.assertEqual(report["summary"]["verified_processes"], 470)
        self.assertEqual(report["summary"]["failed_processes"], 0)
        self.assertEqual(report["summary"]["timed_out_processes"], 0)
        self.assertFalse(report["summary"]["performance_candidate_admitted"])
        self.assertTrue(report["summary"]["closure_rebaseline_admitted"])
        self.assertEqual(
            report["scope_review"]["generated_modules_matching_prior"], 66
        )
        self.assertEqual(
            report["scope_review"]["boundary_guard"]["interpreter_dependencies"], 0
        )

    def test_json_round_trip_is_deterministic(self) -> None:
        module = load_module()
        report = module.build_report()
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "report.json"
            path.write_text(
                json.dumps(report, indent=2, sort_keys=True) + "\n",
                encoding="utf-8",
            )
            self.assertEqual(json.loads(path.read_text(encoding="utf-8")), report)


if __name__ == "__main__":
    unittest.main()
