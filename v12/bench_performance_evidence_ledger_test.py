#!/usr/bin/env python3
"""Fast contract tests for performance-evidence invalidation selection."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent
REPO_ROOT = ROOT.parent
GENERATOR = ROOT / "bench_performance_evidence_ledger"
CHECKED_REPORT = (
    ROOT
    / "docs/perf-baselines/2026-07-21-performance-evidence-invalidation-ledger.json"
)


class PerformanceEvidenceLedgerTests(unittest.TestCase):
    def run_ledger(self, *args: str) -> subprocess.CompletedProcess[str]:
        return subprocess.run(
            [str(GENERATOR), *args],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )

    def current_report(self) -> dict[str, object]:
        result = self.run_ledger()
        self.assertEqual(result.returncode, 0, result.stderr)
        return json.loads(result.stdout)

    def bootstrap_ledger(self, path: Path) -> dict[str, object]:
        result = self.run_ledger("--bootstrap", "--json-out", str(path))
        self.assertEqual(result.returncode, 0, result.stderr)
        return json.loads(path.read_text(encoding="utf-8"))

    def test_current_ledger_matches_checked_selection(self) -> None:
        report = self.current_report()
        checked = json.loads(CHECKED_REPORT.read_text(encoding="utf-8"))
        self.assertEqual(report, checked)
        self.assertEqual(report["summary"]["closure_count"], 21)
        self.assertEqual(
            report["summary"]["current_count"]
            + report["summary"]["invalidated_count"],
            21,
        )

    def test_one_evidence_identity_invalidates_only_its_closure(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            path = Path(raw_dir) / "ledger.json"
            ledger = self.bootstrap_ledger(path)
            closure = next(
                item for item in ledger["closures"] if item["id"] == "compiled-current-control"
            )
            closure["evidence"][0]["sha256"] = "0" * 64
            path.write_text(json.dumps(ledger), encoding="utf-8")
            result = self.run_ledger("--ledger", str(path))
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(
            set(report["summary"]["selected_closures"]),
            {"compiled-current-control"},
        )
        selected = next(
            item for item in report["closures"] if item["id"] == "compiled-current-control"
        )
        self.assertTrue(any("evidence-content-drift" in reason for reason in selected["reasons"]))

    def test_compiler_production_drift_selects_only_compiled_and_cross_mode(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            path = Path(raw_dir) / "ledger.json"
            ledger = self.bootstrap_ledger(path)
            compiler_closures = {
                closure["id"]
                for closure in ledger["closures"]
                if "compiled" in closure["modes"]
            }
            scope = next(
                item for item in ledger["scopes"] if item["id"] == "compiler-production"
            )
            scope["tree_sha256"] = "0" * 64
            path.write_text(json.dumps(ledger), encoding="utf-8")
            result = self.run_ledger("--ledger", str(path))
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        selected = set(report["summary"]["selected_closures"])
        self.assertEqual(selected, compiler_closures)
        self.assertNotIn("bytecode-register-architecture", selected)
        self.assertEqual(
            set(report["summary"]["selected_by_mode"]["bytecode"]),
            {
                closure["id"]
                for closure in ledger["closures"]
                if closure["id"] in selected and "bytecode" in closure["modes"]
            },
        )

    def test_invalidated_markdown_describes_the_active_selection(self) -> None:
        report = self.current_report()
        if report["summary"]["invalidated_count"] == 0:
            self.skipTest("current checked report has no invalidated closures")
        with tempfile.TemporaryDirectory() as raw_dir:
            path = Path(raw_dir) / "report.md"
            result = self.run_ledger("--markdown-out", str(path))
            self.assertEqual(result.returncode, 0, result.stderr)
            rendered = path.read_text(encoding="utf-8")
        self.assertIn(
            f"currently selects {report['summary']['invalidated_count']} closures",
            rendered,
        )
        self.assertNotIn("Because it currently selects nothing", rendered)

    def test_test_only_change_does_not_invalidate_production_scope(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            directory = Path(raw_dir)
            scope_root = directory / "compiler"
            scope_root.mkdir()
            (scope_root / "prod.go").write_text("package compiler\n", encoding="utf-8")
            (scope_root / "prod_test.go").write_text(
                "package compiler\n", encoding="utf-8"
            )
            baseline = directory / "ledger.json"
            override = f"compiler-production={scope_root}"
            boot = self.run_ledger(
                "--bootstrap", "--scope-override", override, "--json-out", str(baseline)
            )
            self.assertEqual(boot.returncode, 0, boot.stderr)
            (scope_root / "prod_test.go").write_text(
                "package compiler\n// irrelevant test change\n", encoding="utf-8"
            )
            result = self.run_ledger(
                "--ledger", str(baseline), "--scope-override", override
            )
        self.assertEqual(result.returncode, 0, result.stderr)
        report = json.loads(result.stdout)
        self.assertEqual(report["summary"]["selected_closures"], [])

    def test_partial_advance_updates_only_named_closures(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            directory = Path(raw_dir)
            baseline = directory / "ledger.json"
            advanced = directory / "advanced.json"
            ledger = self.bootstrap_ledger(baseline)
            stale_ids = {
                "compiled-target-guards",
                "bytecode-target-guards",
                "compiled-current-control",
            }
            for closure in ledger["closures"]:
                if closure["id"] in stale_ids:
                    closure["definition_sha256"] = "0" * 64
            baseline.write_text(json.dumps(ledger), encoding="utf-8")
            result = self.run_ledger(
                "--ledger",
                str(baseline),
                "--advance-closure",
                "compiled-target-guards",
                "--advance-closure",
                "bytecode-target-guards",
                "--json-out",
                str(advanced),
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            report = self.run_ledger("--ledger", str(advanced))
        self.assertEqual(report.returncode, 0, report.stderr)
        selection = json.loads(report.stdout)["summary"]["selected_closures"]
        self.assertEqual(selection, ["compiled-current-control"])

    def test_partial_advance_refuses_to_mask_shared_scope_drift(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            directory = Path(raw_dir)
            baseline = directory / "ledger.json"
            advanced = directory / "advanced.json"
            ledger = self.bootstrap_ledger(baseline)
            scope = next(item for item in ledger["scopes"] if item["id"] == "v12-spec")
            scope["tree_sha256"] = "0" * 64
            baseline.write_text(json.dumps(ledger), encoding="utf-8")
            result = self.run_ledger(
                "--ledger",
                str(baseline),
                "--advance-closure",
                "compiled-target-guards",
                "--json-out",
                str(advanced),
            )
        self.assertNotEqual(result.returncode, 0)
        self.assertIn("would mask shared scope drift", result.stderr)


if __name__ == "__main__":
    unittest.main()
