#!/usr/bin/env python3
import hashlib
import json
import unittest
from pathlib import Path


ROOT = Path(__file__).resolve().parent.parent
V12 = ROOT / "v12"
REPORT = (
    V12
    / "docs/perf-baselines/2026-07-30-full-compiled-static-native-boundary-census.json"
)


def sha256(path: Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


class CompiledStaticBoundaryCensusTest(unittest.TestCase):
    @classmethod
    def setUpClass(cls) -> None:
        cls.report = json.loads(REPORT.read_text())
        cls.manifest = json.loads((V12 / "bench-selection-manifest.json").read_text())

    def test_report_covers_current_compiled_selection(self) -> None:
        selected = self.manifest["modes"]["compiled"]
        self.assertEqual(self.report["coverage"]["selected_rows"], 66)
        self.assertEqual(self.report["coverage"]["successful_rows"], 66)
        self.assertEqual(self.report["coverage"]["failed_rows"], 0)
        self.assertEqual(
            [row["benchmark"] for row in self.report["rows"]],
            selected,
        )
        self.assertTrue(all(row["generation_exit"] == 0 for row in self.report["rows"]))
        self.assertTrue(all(row["module_sha256"] for row in self.report["rows"]))

    def test_inputs_are_content_addressed(self) -> None:
        for key in ("selection_manifest", "compiled_frontier", "analyzer", "runner"):
            entry = self.report["inputs"][key]
            self.assertEqual(entry["sha256"], sha256(ROOT / entry["path"]))

    def test_scope_contract_and_decision_are_fail_closed(self) -> None:
        for row in self.report["rows"]:
            scopes = row["scopes"]
            self.assertIn("compiled_body", scopes)
            self.assertIn("main_direct_reachable", scopes)
            self.assertGreater(scopes["compiled_body"]["functions"], 0)
            self.assertGreater(scopes["main_direct_reachable"]["functions"], 0)
        decision = self.report["decision"]
        self.assertIsNone(decision["admitted_candidate"])
        self.assertFalse(decision["production_change"])
        self.assertFalse(decision["prototype_or_ab_cohort_run"])
        self.assertEqual(
            decision["disposition"],
            "closed-no-new-shared-lowerable-boundary",
        )

    def test_recurrent_shapes_are_explicitly_disposed(self) -> None:
        candidates = {
            candidate["id"]: candidate
            for candidate in self.report["candidate_review"]
        }
        self.assertEqual(
            candidates["host-output-and-argument-abi"]["callees"][
                "__able_call_named"
            ]["applications"],
            62,
        )
        self.assertEqual(
            candidates["residual-runtime-method-call"]["callees"][
                "__able_method_call_node"
            ]["applications"],
            17,
        )
        callable_shapes = candidates["callable-runtime-conversion"]["callees"]
        self.assertLess(
            max(shape["applications"] for shape in callable_shapes.values()),
            3,
        )
        self.assertEqual(
            self.report["heap_nominal_review"]["disposition"],
            "not-a-shared-removable-boundary",
        )


if __name__ == "__main__":
    unittest.main()
