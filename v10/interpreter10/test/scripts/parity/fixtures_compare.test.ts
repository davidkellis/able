import { describe, expect, test } from "bun:test";

import {
  compareFixtureOutcomes,
  type GoOutcome,
  type TSOutcome,
} from "../../../scripts/parity/fixtures";

describe("compareFixtureOutcomes diagnostics parity", () => {
  const baseTS: TSOutcome = { stdout: [] };
  const baseGo: GoOutcome = { stdout: [] };

  test("flags mismatched diagnostics even without manifest expectations", () => {
    const diff = compareFixtureOutcomes(
      { ...baseTS, diagnostics: ["typechecker: foo"] },
      { ...baseGo, diagnostics: [] },
      "fixtures/sample",
      null,
    );
    expect(diff).not.toBeNull();
    expect(diff?.kind).toBe("diagnostics");
  });

  test("returns null when diagnostics match", () => {
    const diff = compareFixtureOutcomes(
      { ...baseTS, diagnostics: ["typechecker: foo"] },
      { ...baseGo, diagnostics: ["typechecker: foo"] },
      "fixtures/sample",
      null,
    );
    expect(diff).toBeNull();
  });
});
