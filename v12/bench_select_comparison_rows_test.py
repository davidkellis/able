#!/usr/bin/env python3

from __future__ import annotations

import importlib.machinery
import importlib.util
import json
import tempfile
import unittest
from pathlib import Path


SCRIPT = Path(__file__).with_name("bench_select_comparison_rows")
LOADER = importlib.machinery.SourceFileLoader("bench_select_comparison_rows", str(SCRIPT))
SPEC = importlib.util.spec_from_loader(LOADER.name, LOADER)
assert SPEC is not None
MODULE = importlib.util.module_from_spec(SPEC)
LOADER.exec_module(MODULE)


class ComparisonRowSelectionTest(unittest.TestCase):
    def test_selects_requested_order_and_records_source_identity(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            source_path = Path(raw_root) / "source.json"
            source = {
                "generated_at": "2026-07-29T00:00:00Z",
                "suite": "custom",
                "benchmarks": ["alpha", "beta", "gamma"],
                "rows": [
                    {"benchmark": "alpha", "mode": "compiled"},
                    {"benchmark": "beta", "mode": "compiled"},
                    {"benchmark": "gamma", "mode": "compiled"},
                ],
            }
            source_path.write_text(json.dumps(source), encoding="utf-8")

            selected = MODULE.selected_report(source, ["gamma", "alpha"], source_path)

            self.assertEqual(selected["benchmarks"], ["gamma", "alpha"])
            self.assertEqual(
                [row["benchmark"] for row in selected["rows"]],
                ["gamma", "alpha"],
            )
            self.assertEqual(
                selected["row_selection"]["source_path"],
                str(source_path),
            )
            self.assertEqual(len(selected["row_selection"]["source_sha256"]), 64)

    def test_rejects_missing_or_duplicate_rows(self) -> None:
        with tempfile.TemporaryDirectory() as raw_root:
            source_path = Path(raw_root) / "source.json"
            source_path.write_text("{}", encoding="utf-8")
            source = {
                "rows": [
                    {"benchmark": "alpha"},
                    {"benchmark": "alpha"},
                ]
            }
            with self.assertRaisesRegex(ValueError, "repeats benchmark"):
                MODULE.selected_report(source, ["alpha"], source_path)

            source = {"rows": [{"benchmark": "alpha"}]}
            with self.assertRaisesRegex(ValueError, "missing selected rows"):
                MODULE.selected_report(source, ["beta"], source_path)


if __name__ == "__main__":
    unittest.main()
