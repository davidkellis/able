#!/usr/bin/env python3
"""Fast tests for retained scorecard dependency hermeticity."""

from __future__ import annotations

import hashlib
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


def load_scoreboard() -> ModuleType:
    path = SCRIPT_DIR / "bench_external_scoreboard"
    loader = importlib.machinery.SourceFileLoader("external_scoreboard_tested", str(path))
    spec = importlib.util.spec_from_loader(loader.name, loader)
    if spec is None:
        raise RuntimeError(f"cannot load {path}")
    module = importlib.util.module_from_spec(spec)
    loader.exec_module(module)
    return module


scoreboard = load_scoreboard()


class RetainedDependencyTests(unittest.TestCase):
    def check_fixture(
        self, external_reference: bool, relocated: bool = False
    ) -> tuple[int, int]:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            perf_dir = root / "v12" / "docs" / "perf-baselines"
            perf_dir.mkdir(parents=True)
            reference = (root if external_reference else perf_dir) / "reference.json"
            reference.write_text("{}\n", encoding="utf-8")
            source = perf_dir / "source.json"
            source.write_text(
                json.dumps(
                    {
                        "go_reference_json": str(
                            Path("/retired/checkout/v12/docs/perf-baselines/reference.json")
                            if relocated
                            else reference
                        )
                    }
                )
                + "\n",
                encoding="utf-8",
            )
            scorecard = perf_dir / "current.json"
            scorecard.write_text(
                json.dumps(
                    {
                        "sources": [
                            {
                                "path": str(
                                    Path(
                                        "/retired/checkout/v12/docs/perf-baselines/source.json"
                                    )
                                    if relocated
                                    else source
                                )
                            }
                        ]
                    }
                )
                + "\n",
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

    def test_relocated_retained_reports_are_accepted(self) -> None:
        self.assertEqual(
            self.check_fixture(external_reference=False, relocated=True),
            (1, 1),
        )

    def test_existing_external_reference_report_is_rejected(self) -> None:
        with self.assertRaisesRegex(ValueError, "is not retained under"):
            self.check_fixture(external_reference=True)

    def test_scoreboard_relocates_measured_able_source(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            target = root / "v12" / "examples" / "benchmarks" / "fib.able"
            target.parent.mkdir(parents=True)
            target.write_text("module Main\n", encoding="utf-8")
            source_sha = hashlib.sha256(target.read_bytes()).hexdigest()
            row = {
                "benchmark": "fib",
                "mode": "compiled",
                "able": {
                    "source": {
                        "path": "/retired/checkout/v12/examples/benchmarks/fib.able",
                        "sha256": source_sha,
                    }
                },
            }
            old_root = scoreboard.REPO_ROOT
            scoreboard.REPO_ROOT = root
            try:
                self.assertEqual(
                    scoreboard.scorecard_source_fingerprint(row, target),
                    {
                        "path": "v12/examples/benchmarks/fib.able",
                        "sha256": source_sha,
                        "provenance": "measured",
                    },
                )
            finally:
                scoreboard.REPO_ROOT = old_root


if __name__ == "__main__":
    unittest.main()
