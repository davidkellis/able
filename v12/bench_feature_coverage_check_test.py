#!/usr/bin/env python3
"""Fast failure-mode tests for the feature coverage contract."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path
from typing import Any


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
CHECKER = SCRIPT_DIR / "bench_feature_coverage_check"
DEFAULT_MANIFEST = SCRIPT_DIR / "bench-feature-coverage.json"


class FeatureCoverageTests(unittest.TestCase):
    def run_manifest(self, manifest: dict[str, Any]) -> subprocess.CompletedProcess[str]:
        with tempfile.TemporaryDirectory() as raw_root:
            path = Path(raw_root) / "coverage.json"
            path.write_text(json.dumps(manifest), encoding="utf-8")
            return subprocess.run(
                [str(CHECKER), "--manifest", str(path)],
                cwd=REPO_ROOT,
                text=True,
                capture_output=True,
                check=False,
            )

    def manifest(self) -> dict[str, Any]:
        return json.loads(DEFAULT_MANIFEST.read_text(encoding="utf-8"))

    def test_checked_in_manifest_passes(self) -> None:
        result = subprocess.run(
            [str(CHECKER)],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("16 normative sections, 63 portable applications", result.stdout)

    def test_uncovered_spec_section_fails(self) -> None:
        manifest = self.manifest()
        manifest["families"] = [
            family for family in manifest["families"] if family["id"] != "testing_framework"
        ]
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("uncovered normative spec sections: 17", result.stderr)

    def test_unmapped_portable_application_fails(self) -> None:
        manifest = self.manifest()
        for family in manifest["families"]:
            family["portable_benchmarks"] = [
                benchmark for benchmark in family["portable_benchmarks"] if benchmark != "fib"
            ]
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("portable applications absent from feature coverage: fib", result.stderr)

    def test_local_only_family_cannot_claim_portable_timing(self) -> None:
        manifest = self.manifest()
        family = next(
            family for family in manifest["families"] if family["id"] == "host_interop"
        )
        family["portable_benchmarks"] = ["fib"]
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("local_only coverage cannot claim a portable benchmark", result.stderr)

    def test_missing_local_guard_fails(self) -> None:
        manifest = self.manifest()
        family = next(
            family for family in manifest["families"] if family["id"] == "testing_framework"
        )
        family["local_fixtures"] = ["fixtures/exec/not_present"]
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("local fixture has no main.able", result.stderr)


if __name__ == "__main__":
    unittest.main()
