#!/usr/bin/env python3
"""Fast tests for the shared-value production-pilot decision record."""

from __future__ import annotations

import importlib.machinery
import importlib.util
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
SCRIPT = ROOT / "bench_shared_value_heap_production_pilot"


def load_module():
    loader = importlib.machinery.SourceFileLoader(
        "shared_value_heap_production_pilot", str(SCRIPT)
    )
    spec = importlib.util.spec_from_loader(loader.name, loader)
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


class SharedValueHeapProductionPilotTest(unittest.TestCase):
    def test_checked_record_recomputes_all_means(self) -> None:
        module = load_module()
        report = module.load(module.DEFAULT_REPORT)
        module.validate_report(report)
        self.assertEqual(
            report["decision"],
            "reject-live-cell-call-veneer-require-native-shared-ownership",
        )
        self.assertEqual(report["measurement"]["verified_selection_processes"], 120)
        self.assertFalse(report["summary"]["candidate_admitted"])

    def test_rejected_live_path_is_absent(self) -> None:
        module = load_module()
        module.validate_revert()


if __name__ == "__main__":
    unittest.main()
