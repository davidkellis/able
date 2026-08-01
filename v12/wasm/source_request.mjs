import { loadStaticModuleClosure } from "./module_loader.mjs";

// buildSourceEvaluationRequest composes a host-owned parser and source
// provider into the JSON-ready request accepted by __able_eval_request_json.
// It deliberately does not start WASM or select host APIs, so the same module
// can be used by a browser embedding and the Node prototype CLI.
export async function buildSourceEvaluationRequest({
  entryPath,
  moduleRoots = [],
  execMode = "treewalker",
  parseSource,
  sourceProvider,
}) {
  const closure = await loadStaticModuleClosure({
    entryPath,
    moduleRoots,
    parseSource,
    sourceProvider,
  });
  return {
    execMode,
    setupModules: closure.setupModules,
    module: closure.entry.module,
    entryOrigin: closure.entry.origin,
  };
}
