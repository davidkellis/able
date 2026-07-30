#!/usr/bin/env python3
"""Fast contract tests for the feature-interaction matrix generator."""

from __future__ import annotations

import json
import subprocess
import tempfile
import unittest
from pathlib import Path


SCRIPT_DIR = Path(__file__).resolve().parent
REPO_ROOT = SCRIPT_DIR.parent
GENERATOR = SCRIPT_DIR / "bench_feature_interaction_matrix"
BASELINE_WITHOUT = (
    "backup_dedup,binary_event_log,concurrent_document_pipeline,"
    "concurrent_text_index,validated_job_pipeline,dependency_wave_validation,"
    "concurrent_event_routing,manifest_normalization,policy_record_dispatch,"
    "sensor_calibration,concurrent_stencil_reduction,concurrent_signal_dispatch,"
    "concurrent_transform_chain,concurrent_policy_callbacks,"
    "concurrent_graph_visitors,concurrent_audio_voices,"
    "concurrent_packet_codecs,concurrent_scene_tiles,concurrent_tree_folds,"
    "concurrent_state_machines,concurrent_stateful_pipeline,"
    "discrete_event_simulation,transaction_ledger_audit"
)


class FeatureInteractionMatrixTests(unittest.TestCase):
    def test_interaction_applications_close_real_pairwise_gaps(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            output = Path(raw_dir) / "matrix.json"
            result = subprocess.run(
                [
                    str(GENERATOR),
                    "--baseline-without",
                    BASELINE_WITHOUT,
                    "--json-out",
                    str(output),
                ],
                cwd=REPO_ROOT,
                text=True,
                capture_output=True,
                check=False,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            report = json.loads(output.read_text(encoding="utf-8"))
            self.assertEqual(report["family_count"], 11)
            self.assertEqual(report["pair_count"], 55)
            self.assertEqual(report["baseline_zero_pairs"], 29)
            self.assertEqual(report["current_zero_pairs"], 0)
            self.assertGreater(report["improved_pairs"], 0)
            pairs = {
                (pair["left"], pair["right"]): pair for pair in report["pairs"]
            }
            concurrency_errors = pairs[
                ("concurrency", "option_result_exceptions")
            ]
            self.assertEqual(concurrency_errors["baseline_count"], 0)
            self.assertEqual(concurrency_errors["current_count"], 16)
            self.assertEqual(
                concurrency_errors["portable_benchmarks"],
                [
                    "concurrent_audio_voices",
                    "concurrent_document_pipeline",
                    "concurrent_event_routing",
                    "concurrent_graph_visitors",
                    "concurrent_packet_codecs",
                    "concurrent_policy_callbacks",
                    "concurrent_scene_tiles",
                    "concurrent_signal_dispatch",
                    "concurrent_state_machines",
                    "concurrent_stateful_pipeline",
                    "concurrent_stencil_reduction",
                    "concurrent_text_index",
                    "concurrent_transform_chain",
                    "concurrent_tree_folds",
                    "dependency_wave_validation",
                    "validated_job_pipeline",
                ],
            )
            control_concurrency = pairs[("concurrency", "control_flow")]
            self.assertEqual(control_concurrency["baseline_count"], 0)
            self.assertEqual(control_concurrency["current_count"], 15)
            self.assertEqual(
                control_concurrency["portable_benchmarks"],
                [
                    "concurrent_audio_voices",
                    "concurrent_document_pipeline",
                    "concurrent_event_routing",
                    "concurrent_graph_visitors",
                    "concurrent_packet_codecs",
                    "concurrent_policy_callbacks",
                    "concurrent_scene_tiles",
                    "concurrent_signal_dispatch",
                    "concurrent_state_machines",
                    "concurrent_stateful_pipeline",
                    "concurrent_stencil_reduction",
                    "concurrent_text_index",
                    "concurrent_transform_chain",
                    "concurrent_tree_folds",
                    "dependency_wave_validation",
                ],
            )
            lexical_concurrency = pairs[
                ("concurrency", "lexical_blocks_bindings_patterns")
            ]
            self.assertEqual(lexical_concurrency["baseline_count"], 0)
            self.assertEqual(lexical_concurrency["current_count"], 15)
            self.assertEqual(
                lexical_concurrency["portable_benchmarks"],
                [
                    "concurrent_audio_voices",
                    "concurrent_document_pipeline",
                    "concurrent_event_routing",
                    "concurrent_graph_visitors",
                    "concurrent_packet_codecs",
                    "concurrent_policy_callbacks",
                    "concurrent_scene_tiles",
                    "concurrent_signal_dispatch",
                    "concurrent_state_machines",
                    "concurrent_stateful_pipeline",
                    "concurrent_stencil_reduction",
                    "concurrent_text_index",
                    "concurrent_transform_chain",
                    "concurrent_tree_folds",
                    "dependency_wave_validation",
                ],
            )
            lexical_functions = pairs[
                ("functions_closures_callables", "lexical_blocks_bindings_patterns")
            ]
            self.assertEqual(lexical_functions["baseline_count"], 0)
            self.assertEqual(lexical_functions["current_count"], 16)
            self.assertEqual(
                lexical_functions["portable_benchmarks"],
                [
                    "backup_dedup",
                    "binary_event_log",
                    "concurrent_audio_voices",
                    "concurrent_document_pipeline",
                    "concurrent_event_routing",
                    "concurrent_graph_visitors",
                    "concurrent_packet_codecs",
                    "concurrent_policy_callbacks",
                    "concurrent_scene_tiles",
                    "concurrent_state_machines",
                    "concurrent_stateful_pipeline",
                    "concurrent_transform_chain",
                    "concurrent_tree_folds",
                    "dependency_wave_validation",
                    "manifest_normalization",
                    "policy_record_dispatch",
                ],
            )

    def test_single_family_membership_baseline_is_exact(self) -> None:
        with tempfile.TemporaryDirectory() as raw_dir:
            output = Path(raw_dir) / "matrix.json"
            result = subprocess.run(
                [
                    str(GENERATOR),
                    "--baseline-remove",
                    "lexical_blocks_bindings_patterns=dependency_wave_validation",
                    "--json-out",
                    str(output),
                ],
                cwd=REPO_ROOT,
                text=True,
                capture_output=True,
                check=False,
            )
            self.assertEqual(result.returncode, 0, result.stderr)
            report = json.loads(output.read_text(encoding="utf-8"))
            self.assertEqual(report["baseline_zero_pairs"], 0)
            self.assertEqual(report["current_zero_pairs"], 0)
            self.assertEqual(report["baseline_min_depth"], 13)
            self.assertEqual(report["current_min_depth"], 13)
            self.assertEqual(report["baseline_depth_one_pairs"], 0)
            self.assertEqual(report["current_depth_one_pairs"], 0)
            self.assertEqual(report["improved_pairs"], 8)
            self.assertEqual(
                report["baseline_removed_memberships"],
                [
                    {
                        "family": "lexical_blocks_bindings_patterns",
                        "portable_benchmarks": ["dependency_wave_validation"],
                    }
                ],
            )
            pairs = {
                (pair["left"], pair["right"]): pair for pair in report["pairs"]
            }
            concurrency_lexical = pairs[
                ("concurrency", "lexical_blocks_bindings_patterns")
            ]
            self.assertEqual(concurrency_lexical["baseline_count"], 14)
            self.assertEqual(concurrency_lexical["current_count"], 15)
            self.assertEqual(
                concurrency_lexical["portable_benchmarks"],
                [
                    "concurrent_audio_voices",
                    "concurrent_document_pipeline",
                    "concurrent_event_routing",
                    "concurrent_graph_visitors",
                    "concurrent_packet_codecs",
                    "concurrent_policy_callbacks",
                    "concurrent_scene_tiles",
                    "concurrent_signal_dispatch",
                    "concurrent_state_machines",
                    "concurrent_stateful_pipeline",
                    "concurrent_stencil_reduction",
                    "concurrent_text_index",
                    "concurrent_transform_chain",
                    "concurrent_tree_folds",
                    "dependency_wave_validation",
                ],
            )

    def test_membership_baseline_rejects_unknown_memberships(self) -> None:
        cases = [
            (
                "missing_family=dependency_wave_validation",
                "unknown portable family",
            ),
            (
                "lexical_blocks_bindings_patterns=missing_benchmark",
                "benchmarks absent",
            ),
        ]
        for removal, expected_error in cases:
            with self.subTest(removal=removal):
                result = subprocess.run(
                    [str(GENERATOR), "--baseline-remove", removal],
                    cwd=REPO_ROOT,
                    text=True,
                    capture_output=True,
                    check=False,
                )
                self.assertNotEqual(result.returncode, 0)
                self.assertIn(expected_error, result.stderr)


if __name__ == "__main__":
    unittest.main()
