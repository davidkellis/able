import path from "node:path";
import { fileURLToPath } from "node:url";

type Parser = import("web-tree-sitter").Parser;

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const REPO_ROOT = path.resolve(__dirname, "../../..");
const WASM_PATH = path.join(REPO_ROOT, "parser10", "tree-sitter-able", "tree-sitter-able.wasm");

let parserPromise: Promise<Parser> | null = null;

export async function getTreeSitterParser(): Promise<Parser> {
  if (!parserPromise) {
    parserPromise = initParser();
  }
  return parserPromise;
}

async function initParser(): Promise<Parser> {
  const { Parser, Language } = await import("web-tree-sitter");
  await Parser.init();
  const parser = new Parser();
  const language = await Language.load(WASM_PATH);
  parser.setLanguage(language);
  return parser;
}

export { WASM_PATH };
