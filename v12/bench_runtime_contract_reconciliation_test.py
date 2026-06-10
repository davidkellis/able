#!/usr/bin/env python3
"""Fast contract tests for runtime-contract evidence reconciliation."""

from __future__ import annotations

import importlib.machinery
import importlib.util
import json
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
SCRIPT = ROOT / "bench_runtime_contract_reconciliation"


def load_module():
    loader = importlib.machinery.SourceFileLoader("runtime_contract_reconciliation", str(SCRIPT))
    spec = importlib.util.spec_from_loader(loader.name, loader)
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


class RuntimeContractReconciliationTest(unittest.TestCase):
    def test_checked_report_has_complete_repeated_processes(self) -> None:
        module = load_module()
        report = module.build_report(
            module.DEFAULT_COMPILED,
            module.DEFAULT_BYTECODE_FORWARD,
            module.DEFAULT_BYTECODE_REVERSE,
            module.DEFAULT_INVALIDATION,
        )
        self.assertEqual(report["decision"], "rebaseline-reviewed-shared-runtime-scope")
        self.assertEqual(report["summary"]["total_verified_processes"], 120)
        self.assertEqual(report["summary"]["failed_processes"], 0)
        self.assertEqual(report["summary"]["timed_out_processes"], 0)
        self.assertEqual(report["summary"]["benchmark_mode_rows"], 12)
        self.assertFalse(report["summary"]["performance_candidate_admitted"])
        self.assertTrue(report["summary"]["closure_rebaseline_admitted"])
        self.assertEqual(
            report["scope_review"]["new_runtime_api_ordinary_production_call_count"], 0
        )

    def test_json_round_trip_is_deterministic(self) -> None:
        module = load_module()
        report = module.build_report(
            module.DEFAULT_COMPILED,
            module.DEFAULT_BYTECODE_FORWARD,
            module.DEFAULT_BYTECODE_REVERSE,
            module.DEFAULT_INVALIDATION,
        )
        with tempfile.TemporaryDirectory() as directory:
            path = Path(directory) / "report.json"
            path.write_text(json.dumps(report, indent=2, sort_keys=True) + "\n", encoding="utf-8")
            self.assertEqual(json.loads(path.read_text(encoding="utf-8")), report)


if __name__ == "__main__":
    unittest.main()
