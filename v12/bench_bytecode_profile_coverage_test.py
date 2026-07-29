#!/usr/bin/env python3
"""Tests for the bytecode CPU/allocation source-identity ledger."""

from __future__ import annotations

import hashlib
import json
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
sys.path.insert(0, str(SCRIPT_DIR))

from bench_bytecode_profile_coverage import build_report  # noqa: E402
from bench_scorecard_selection import semantic_sha256  # noqa: E402


class BytecodeProfileCoverageTests(unittest.TestCase):
    def fixture(self, root: Path) -> tuple[Path, Path, Path, Path, Path]:
        source = root / "source.able"
        source.write_text("fn main() {}\n", encoding="utf-8")
        evidence = root / "evidence.md"
        evidence.write_text("evidence\n", encoding="utf-8")

        selection = {
            "kind": "external-benchmark-selection-manifest",
            "schema_version": 1,
            "modes": {"compiled": ["guard"], "bytecode": ["alpha"]},
        }
        selection_sha = semantic_sha256(selection)
        selection_path = root / "selection.json"
        selection_path.write_text(json.dumps(selection), encoding="utf-8")

        scorecard = {
            "targets": {"bytecode": {"max_able_to_reference_ratio": 1 / 0.95}},
            "selection_manifest": {"selection_sha256": selection_sha},
            "rows": [
                {
                    "benchmark": "alpha",
                    "mode": "bytecode",
                    "target_status": "miss",
                    "able": {"real_seconds": 2.0},
                    "able_source": {
                        "path": str(source),
                        "sha256": hashlib.sha256(source.read_bytes()).hexdigest(),
                    },
                    "comparisons": {"python": {"real_seconds": 1.0}},
                }
            ],
        }
        scorecard_path = root / "scorecard.json"
        scorecard_path.write_text(json.dumps(scorecard), encoding="utf-8")

        frontier = {
            "selection_sha256": selection_sha,
            "groups": [
                {
                    "id": "bytecode-alpha",
                    "rows": ["alpha/bytecode"],
                    "disposition": "closed-no-shared-leaf",
                }
            ],
        }
        frontier_path = root / "frontier.json"
        frontier_path.write_text(json.dumps(frontier), encoding="utf-8")

        closure = {
            "selection_sha256": selection_sha,
            "closures": [{"id": "closed"}],
        }
        closure_path = root / "closure.json"
        closure_path.write_text(json.dumps(closure), encoding="utf-8")

        coverage = {
            "kind": "able-bytecode-profile-coverage-manifest",
            "schema_version": 1,
            "selection_sha256": selection_sha,
            "groups": [
                {
                    "id": "bytecode-alpha",
                    "cpu_status": "current",
                    "allocation_status": "current",
                    "evidence": [str(evidence)],
                }
            ],
        }
        coverage_path = root / "coverage.json"
        coverage_path.write_text(json.dumps(coverage), encoding="utf-8")
        return (
            scorecard_path,
            selection_path,
            frontier_path,
            closure_path,
            coverage_path,
        )

    def test_builds_complete_source_checked_miss_ledger(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            report = build_report(*self.fixture(Path(raw_root)))
        self.assertEqual(report["summary"]["bytecode_target_misses"], 1)
        self.assertEqual(report["summary"]["cpu_current_rows"], 1)
        self.assertEqual(report["summary"]["allocation_current_rows"], 1)
        self.assertEqual(report["summary"]["uncovered_rows"], 0)
        self.assertEqual(report["summary"]["closure_ledger_entries"], 1)
        self.assertEqual(report["rows"][0]["benchmark"], "alpha")

    def test_rejects_stale_selected_source(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            paths = self.fixture(root)
            scorecard = json.loads(paths[0].read_text(encoding="utf-8"))
            scorecard["rows"][0]["able_source"]["sha256"] = "0" * 64
            paths[0].write_text(json.dumps(scorecard), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "source identity is stale"):
                build_report(*paths)

    def test_rejects_incomplete_miss_coverage(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            paths = self.fixture(root)
            coverage = json.loads(paths[4].read_text(encoding="utf-8"))
            coverage["groups"] = []
            paths[4].write_text(json.dumps(coverage), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "needs groups"):
                build_report(*paths)


if __name__ == "__main__":
    unittest.main()
