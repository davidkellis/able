#!/usr/bin/env python3
"""Fast tests for retained scorecard dependency hermeticity."""

from __future__ import annotations

import importlib.machinery
import importlib.util
import json
import sys
import tempfile
import unittest
from pathlib import Path
from types import ModuleType


SCRIPT_DIR = Path(__file__).resolve().parent
sys.path.insert(0, str(SCRIPT_DIR))


def load_script() -> ModuleType:
    path = SCRIPT_DIR / "bench_scorecard_evidence_check.py"
    loader = importlib.machinery.SourceFileLoader("scorecard_evidence_check_tested", str(path))
    spec = importlib.util.spec_from_loader(loader.name, loader)
    if spec is None:
        raise RuntimeError(f"cannot load {path}")
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


evidence = load_script()


class RetainedDependencyTests(unittest.TestCase):
    def check_fixture(self, external_reference: bool) -> tuple[int, int]:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            perf_dir = root / "v12" / "docs" / "perf-baselines"
            perf_dir.mkdir(parents=True)
            reference = (root if external_reference else perf_dir) / "reference.json"
            reference.write_text("{}\n", encoding="utf-8")
            source = perf_dir / "source.json"
            source.write_text(
                json.dumps({"go_reference_json": str(reference)}) + "\n",
                encoding="utf-8",
            )
            scorecard = perf_dir / "current.json"
            scorecard.write_text(
                json.dumps({"sources": [{"path": str(source)}]}) + "\n",
                encoding="utf-8",
            )
            old_root, old_perf = evidence.REPO_ROOT, evidence.PERF_DIR
            evidence.REPO_ROOT, evidence.PERF_DIR = root, perf_dir
            try:
                return evidence.validate_retained_dependencies(scorecard)
            finally:
                evidence.REPO_ROOT, evidence.PERF_DIR = old_root, old_perf

    def test_retained_reference_report_is_accepted(self) -> None:
        self.assertEqual(self.check_fixture(external_reference=False), (1, 1))

    def test_existing_external_reference_report_is_rejected(self) -> None:
        with self.assertRaisesRegex(ValueError, "is not retained under"):
            self.check_fixture(external_reference=True)


if __name__ == "__main__":
    unittest.main()
