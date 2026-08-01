import path from "node:path";
import { fileURLToPath } from "node:url";

import { Language, Parser } from "web-tree-sitter";

export { parseSourceToAstModule } from "./ast_adapter.mjs";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const DEFAULT_LANGUAGE_WASM_PATH = path.resolve(
  __dirname,
  "../parser/tree-sitter-able/tree-sitter-able.wasm",
);

// createAbleParser is the Node-only tree-sitter bootstrap used by the CLI.
// Browser callers initialize web-tree-sitter themselves, then share the AST
// conversion from ast_adapter.mjs.
export async function createAbleParser(languageWasmPath = DEFAULT_LANGUAGE_WASM_PATH) {
  await Parser.init();
  const language = await Language.load(languageWasmPath);
  const parser = new Parser();
  parser.setLanguage(language);
  return parser;
}
