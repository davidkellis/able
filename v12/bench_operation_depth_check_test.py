#!/usr/bin/env python3
"""Fast failure-mode tests for the operation-depth contract."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path
from typing import Any


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
CHECKER = SCRIPT_DIR / "bench_operation_depth_check"
DEFAULT_MANIFEST = SCRIPT_DIR / "bench-operation-depth.json"


class OperationDepthTests(unittest.TestCase):
    def manifest(self) -> dict[str, Any]:
        return json.loads(DEFAULT_MANIFEST.read_text(encoding="utf-8"))

    def run_manifest(self, manifest: dict[str, Any]) -> subprocess.CompletedProcess[str]:
        with tempfile.TemporaryDirectory() as raw_root:
            path = Path(raw_root) / "depth.json"
            path.write_text(json.dumps(manifest), encoding="utf-8")
            return subprocess.run(
                [str(CHECKER), "--manifest", str(path)],
                cwd=REPO_ROOT,
                text=True,
                capture_output=True,
                check=False,
            )

    def operation(self, manifest: dict[str, Any], operation_id: str) -> dict[str, Any]:
        return next(item for item in manifest["operations"] if item["id"] == operation_id)

    def test_checked_in_manifest_passes(self) -> None:
        result = subprocess.run(
            [str(CHECKER)],
            cwd=REPO_ROOT,
            text=True,
            capture_output=True,
            check=False,
        )
        self.assertEqual(result.returncode, 0, result.stderr)
        self.assertIn("21 operations, 18 sufficient, 0 insufficient, 3 local-only", result.stdout)
        self.assertIn("0 actionable", result.stdout)

    def test_unknown_feature_family_fails(self) -> None:
        manifest = self.manifest()
        self.operation(manifest, "recursive_user_function_calls")["feature_family"] = "not_a_family"
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("unknown feature family: not_a_family", result.stderr)

    def test_stale_source_marker_fails(self) -> None:
        manifest = self.manifest()
        evidence = self.operation(manifest, "recursive_user_function_calls")["portable_evidence"][0]
        evidence["source_markers"] = ["not present in fib source"]
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("source marker not found in fib", result.stderr)

    def test_inflated_unlike_breadth_fails(self) -> None:
        manifest = self.manifest()
        operation = self.operation(manifest, "recursive_user_function_calls")
        for evidence in operation["portable_evidence"]:
            evidence["workload"] = "one_related_workload"
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("does not match 1 unlike workloads (insufficient)", result.stderr)

    def test_unknown_portable_benchmark_fails(self) -> None:
        manifest = self.manifest()
        evidence = self.operation(manifest, "hash_map_lookup_and_update")["portable_evidence"][0]
        evidence["benchmark"] = "not_in_coverage"
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("benchmark is outside coverage: not_in_coverage", result.stderr)

    def test_unknown_frontier_group_fails(self) -> None:
        manifest = self.manifest()
        operation = self.operation(manifest, "hash_map_lookup_and_update")
        operation["frontier_groups"] = ["not-a-frontier-group"]
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("unknown frontier group: not-a-frontier-group", result.stderr)

    def test_local_only_operation_cannot_claim_portable_evidence(self) -> None:
        manifest = self.manifest()
        operation = self.operation(manifest, "user_authored_host_interop")
        operation["portable_evidence"] = [
            {
                "benchmark": "fib",
                "workload": "numeric_recursion",
                "source_markers": ["fn fib(n: i32)"],
            }
        ]
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("local_only coverage needs fixtures and no portable evidence", result.stderr)

    def test_every_feature_family_needs_an_operation(self) -> None:
        manifest = self.manifest()
        manifest["operations"] = [
            operation
            for operation in manifest["operations"]
            if operation["id"] != "test_reporter_lifecycle"
        ]
        result = self.run_manifest(manifest)
        self.assertEqual(result.returncode, 2)
        self.assertIn("feature families absent from operation depth: testing_framework", result.stderr)


if __name__ == "__main__":
    unittest.main()
