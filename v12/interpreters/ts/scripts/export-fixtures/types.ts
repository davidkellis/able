import { AST } from "../index";

export interface Fixture {
  /**
   * Folder relative to the fixture root (fixtures/ast).
   */
  name: string;
  module: AST.Module;
  setupModules?: Record<string, AST.Module>;
  manifest?: FixtureManifest;
}

export interface FixtureManifest {
  description: string;
  entry?: string;
  expect?: Record<string, unknown>;
  setup?: string[];
}
