#!/usr/bin/env python3
"""Tests for the complete cross-mode performance-frontier ledger."""

from __future__ import annotations

import json
import sys
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
sys.path.insert(0, str(SCRIPT_DIR))

from bench_performance_frontier import build_frontier  # noqa: E402
from bench_scorecard_selection import semantic_sha256  # noqa: E402


class PerformanceFrontierTests(unittest.TestCase):
    def fixture(self, root: Path) -> tuple[Path, Path, Path, Path]:
        selection = {
            "kind": "external-benchmark-selection-manifest",
            "schema_version": 1,
            "modes": {"compiled": ["alpha"], "bytecode": ["beta", "gamma"]},
        }
        selection_path = root / "selection.json"
        selection_path.write_text(json.dumps(selection), encoding="utf-8")
        selection_sha = semantic_sha256(selection)
        evidence_doc = root / "evidence.md"
        evidence_doc.write_text("evidence\n", encoding="utf-8")
        evidence = {
            "kind": "able-performance-frontier-evidence",
            "schema_version": 1,
            "selection_sha256": selection_sha,
            "groups": [
                {
                    "id": "closed",
                    "rows": ["alpha/compiled"],
                    "profile_freshness": "current-exact",
                    "artifact_identity": "matched",
                    "disposition": "closed-no-shared-leaf",
                    "exact_leaf_breadth": 0,
                    "owner": "compiled owner",
                    "rationale": "closed rationale",
                    "evidence": [str(evidence_doc)],
                },
                {
                    "id": "smaller-refresh",
                    "rows": ["beta/bytecode"],
                    "profile_freshness": "stale",
                    "artifact_identity": "not retained",
                    "disposition": "refresh-required",
                    "exact_leaf_breadth": 0,
                    "owner": "beta owner",
                    "rationale": "beta rationale",
                    "evidence": [str(evidence_doc)],
                },
                {
                    "id": "larger-refresh",
                    "rows": ["gamma/bytecode"],
                    "profile_freshness": "stale",
                    "artifact_identity": "not retained",
                    "disposition": "refresh-required",
                    "exact_leaf_breadth": 0,
                    "owner": "gamma owner",
                    "rationale": "gamma rationale",
                    "evidence": [str(evidence_doc)],
                },
            ],
        }
        evidence_path = root / "evidence.json"
        evidence_path.write_text(json.dumps(evidence), encoding="utf-8")

        def row(benchmark: str, mode: str, able: float, reference: float) -> dict:
            return {
                "benchmark": benchmark,
                "mode": mode,
                "able": {"status": "verified", "real_seconds": able},
                "able_source": {"path": benchmark, "sha256": "a" * 64},
                "comparisons": {
                    "go" if mode == "compiled" else "python": {
                        "real_seconds": reference,
                        "ratio": able / reference,
                        "source": {"sha256": "b" * 64},
                    }
                },
                "target_status": "meets" if able / reference <= 1 / 0.95 else "miss",
            }

        scorecard = {
            "generated_at": "test-time",
            "targets": {
                "compiled": {"max_able_to_reference_ratio": 1 / 0.95},
                "bytecode": {"max_able_to_reference_ratio": 1 / 0.95},
            },
            "selection_manifest": {"selection_sha256": selection_sha},
            "canonical_stdlib_source_state": {"source_tree_sha256": "c" * 64},
            "rows": [
                row("alpha", "compiled", 1.0, 1.0),
                row("beta", "bytecode", 2.0, 1.0),
                row("gamma", "bytecode", 4.0, 1.0),
                row("unselected", "bytecode", 100.0, 1.0),
            ],
        }
        scorecard_path = root / "scorecard.json"
        scorecard_path.write_text(json.dumps(scorecard), encoding="utf-8")
        stability = {
            "kind": "able-performance-stability",
            "schema_version": 1,
            "selection_sha256": selection_sha,
            "scorecard_stdlib_source_tree_sha256": "c" * 64,
            "entries": [
                {
                    "benchmark": "alpha",
                    "mode": "compiled",
                    "classification": "established-meet",
                    "pooled_ratio": 1.0,
                    "cohort_ratios": [1.0, 1.01],
                    "limiting_reference": "go",
                    "able_samples": 10,
                    "reference_samples": 10,
                    "able_source_sha256": "a" * 64,
                    "reference_source_sha256": {"go": "b" * 64},
                    "evidence_stdlib_source_tree_sha256": "c" * 64,
                    "evidence": [str(evidence_doc)],
                    "rationale": "stable test guard",
                }
            ],
        }
        stability_path = root / "stability.json"
        stability_path.write_text(json.dumps(stability), encoding="utf-8")
        return scorecard_path, selection_path, evidence_path, stability_path

    def test_complete_selection_and_excess_rank_refresh_groups(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            paths = self.fixture(Path(raw_root))
            report = build_frontier(*paths)
        self.assertEqual(report["summary"]["selected_rows"], 3)
        self.assertEqual(report["summary"]["compiled_rows"], 1)
        self.assertEqual(report["summary"]["bytecode_rows"], 2)
        self.assertEqual(report["summary"]["target_meets"], 1)
        self.assertEqual(report["summary"]["established_guards"], 1)
        self.assertEqual(report["summary"]["unestablished_snapshot_meets"], 0)
        self.assertEqual(
            [group["id"] for group in report["actionable_groups"]],
            ["larger-refresh", "smaller-refresh"],
        )
        self.assertEqual(report["recommendation"]["group"], "larger-refresh")
        beta = next(row for row in report["rows"] if row["benchmark"] == "beta")
        self.assertAlmostEqual(beta["excess_seconds"], 2 - 1 / 0.95)
        self.assertNotIn("unselected", {row["benchmark"] for row in report["rows"]})

    def test_evidence_must_cover_every_selected_row_once(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            scorecard, selection, evidence, stability = self.fixture(root)
            value = json.loads(evidence.read_text(encoding="utf-8"))
            value["groups"][2]["rows"] = ["beta/bytecode"]
            evidence.write_text(json.dumps(value), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "assigns beta/bytecode"):
                build_frontier(scorecard, selection, evidence, stability)

    def test_stability_must_cover_each_snapshot_meet(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            scorecard, selection, evidence, stability = self.fixture(root)
            value = json.loads(scorecard.read_text(encoding="utf-8"))
            beta = next(row for row in value["rows"] if row["benchmark"] == "beta")
            beta["able"]["real_seconds"] = 1.0
            beta["comparisons"]["python"]["ratio"] = 1.0
            beta["target_status"] = "meets"
            scorecard.write_text(json.dumps(value), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "missing beta/bytecode"):
                build_frontier(scorecard, selection, evidence, stability)

    def test_volatile_crossing_is_not_an_established_guard(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            scorecard, selection, evidence, stability = self.fixture(root)
            value = json.loads(stability.read_text(encoding="utf-8"))
            entry = value["entries"][0]
            entry["classification"] = "volatile-crossing"
            entry["cohort_ratios"] = [1.0, 1.2]
            stability.write_text(json.dumps(value), encoding="utf-8")
            report = build_frontier(scorecard, selection, evidence, stability)
            self.assertEqual(report["summary"]["established_guards"], 0)
            self.assertEqual(report["summary"]["unestablished_snapshot_meets"], 1)

    def test_stability_source_fingerprint_must_match_scorecard(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            root = Path(raw_root)
            scorecard, selection, evidence, stability = self.fixture(root)
            value = json.loads(stability.read_text(encoding="utf-8"))
            value["entries"][0]["able_source_sha256"] = "d" * 64
            stability.write_text(json.dumps(value), encoding="utf-8")
            with self.assertRaisesRegex(ValueError, "Able source is stale"):
                build_frontier(scorecard, selection, evidence, stability)


if __name__ == "__main__":
    unittest.main()
